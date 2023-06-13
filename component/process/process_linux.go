package process

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/netip"
	"os"
	"unsafe"

	"github.com/Dreamacro/clash/common/pool"

	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

type inetDiagRequest struct {
	Family   byte
	Protocol byte
	Ext      byte
	Pad      byte
	States   uint32

	SrcPort [2]byte
	DstPort [2]byte
	Src     [16]byte
	Dst     [16]byte
	If      uint32
	Cookie  [2]uint32
}

type inetDiagResponse struct {
	Family  byte
	State   byte
	Timer   byte
	ReTrans byte

	SrcPort [2]byte
	DstPort [2]byte
	Src     [16]byte
	Dst     [16]byte
	If      uint32
	Cookie  [2]uint32

	Expires uint32
	RQueue  uint32
	WQueue  uint32
	UID     uint32
	INode   uint32
}

func findProcessPath(network string, from netip.AddrPort, to netip.AddrPort) (string, error) {
	inode, uid, err := resolveSocketByNetlink(network, from, to)
	if err != nil {
		return "", err
	}

	return resolveProcessPathByProcSearch(inode, uid)
}

func resolveSocketByNetlink(network string, from netip.AddrPort, to netip.AddrPort) (inode uint32, uid uint32, err error) {
	request := &inetDiagRequest{
		States: 0xffffffff,
		Cookie: [2]uint32{0xffffffff, 0xffffffff},
	}

	if from.Addr().Is4() {
		request.Family = unix.AF_INET
	} else {
		request.Family = unix.AF_INET6
	}

	// Swap src & dst for udp
	// See also https://www.mail-archive.com/netdev@vger.kernel.org/msg248638.html
	switch network {
	case TCP:
		request.Protocol = unix.IPPROTO_TCP

		copy(request.Src[:], from.Addr().AsSlice())
		copy(request.Dst[:], to.Addr().AsSlice())

		binary.BigEndian.PutUint16(request.SrcPort[:], from.Port())
		binary.BigEndian.PutUint16(request.DstPort[:], to.Port())
	case UDP:
		request.Protocol = unix.IPPROTO_UDP

		copy(request.Dst[:], from.Addr().AsSlice())
		copy(request.Src[:], to.Addr().AsSlice())

		binary.BigEndian.PutUint16(request.DstPort[:], from.Port())
		binary.BigEndian.PutUint16(request.SrcPort[:], to.Port())
	default:
		return 0, 0, ErrInvalidNetwork
	}

	conn, err := netlink.Dial(unix.NETLINK_INET_DIAG, nil)
	if err != nil {
		return 0, 0, err
	}
	defer conn.Close()

	message := netlink.Message{
		Header: netlink.Header{
			Type:  20, // SOCK_DIAG_BY_FAMILY
			Flags: netlink.Request,
		},
		Data: (*(*[unsafe.Sizeof(*request)]byte)(unsafe.Pointer(request)))[:],
	}

	messages, err := conn.Execute(message)
	if err != nil {
		return 0, 0, err
	}

	for _, msg := range messages {
		if len(msg.Data) < int(unsafe.Sizeof(inetDiagResponse{})) {
			continue
		}

		response := (*inetDiagResponse)(unsafe.Pointer(&msg.Data[0]))

		return response.INode, response.UID, nil
	}

	return 0, 0, ErrNotFound
}

func resolveProcessPathByProcSearch(inode, uid uint32) (string, error) {
	procDir, err := os.Open("/proc")
	if err != nil {
		return "", err
	}
	defer procDir.Close()

	pids, err := procDir.Readdirnames(-1)
	if err != nil {
		return "", err
	}

	expectedSocketName := fmt.Appendf(nil, "socket:[%d]", inode)

	pathBuffer := pool.Get(64)
	defer pool.Put(pathBuffer)

	readlinkBuffer := pool.Get(32)
	defer pool.Put(readlinkBuffer)

	copy(pathBuffer, "/proc/")

	for _, pid := range pids {
		if !isPid(pid) {
			continue
		}

		pathBuffer = append(pathBuffer[:len("/proc/")], pid...)

		stat := &unix.Stat_t{}
		err = unix.Stat(string(pathBuffer), stat)
		if err != nil {
			continue
		} else if stat.Uid != uid {
			continue
		}

		pathBuffer = append(pathBuffer, "/fd/"...)
		fdsPrefixLength := len(pathBuffer)

		fdDir, err := os.Open(string(pathBuffer))
		if err != nil {
			continue
		}

		fds, err := fdDir.Readdirnames(-1)
		fdDir.Close()
		if err != nil {
			continue
		}

		for _, fd := range fds {
			pathBuffer = pathBuffer[:fdsPrefixLength]

			pathBuffer = append(pathBuffer, fd...)

			n, err := unix.Readlink(string(pathBuffer), readlinkBuffer)
			if err != nil {
				continue
			}

			if bytes.Equal(readlinkBuffer[:n], expectedSocketName) {
				return os.Readlink("/proc/" + pid + "/exe")
			}
		}
	}

	return "", fmt.Errorf("inode %d of uid %d not found", inode, uid)
}

func isPid(name string) bool {
	for _, c := range name {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}
