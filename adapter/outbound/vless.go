package outbound

type (
	Vless       = Vmess
	VlessOption = VmessOption
)

func NewVless(option VlessOption) (*Vless, error) {
	return newVmess(option, true)
}
