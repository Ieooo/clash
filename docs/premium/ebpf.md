---
sidebarTitle: "Feature: eBPF Redirect to TUN"
sidebarOrder: 3
---

# eBPF Redirect to TUN

eBPF redirect to TUN is a feature that intercepts all network traffic on a specific network interface and redirects it to the TUN interface.

::: warning
This feature conflicts with `tun.auto-route`.
:::

It requires [kernel support](https://github.com/iovisor/bcc/blob/master/INSTALL.md#kernel-configuration) and is less tested, however it would bring better performance compared to `tun.auto-redir` and `tun.auto-route`.

## Configuration

```yaml
ebpf:
  redirect-to-tun:
    - eth0
```
