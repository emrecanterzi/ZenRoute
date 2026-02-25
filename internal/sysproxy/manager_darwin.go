package sysproxy

import (
	"fmt"
	"os/exec"
)

type DarwinManager struct {
	service string
}

func NewManager(service string) Manager {
	return &DarwinManager{service: service}
}

func (m *DarwinManager) SetProxy(addr, port string) error {
	fmt.Printf("sysproxy: enabling proxy on %s (%s:%s)\n", m.service, addr, port)
	if err := exec.Command("networksetup", "-setwebproxy", m.service, addr, port).Run(); err != nil {
		return err
	}
	return exec.Command("networksetup", "-setsecurewebproxy", m.service, addr, port).Run()
}

func (m *DarwinManager) UnsetProxy() error {
	fmt.Printf("sysproxy: disabling proxy on %s\n", m.service)
	if err := exec.Command("networksetup", "-setwebproxystate", m.service, "off").Run(); err != nil {
		return err
	}
	return exec.Command("networksetup", "-setsecurewebproxystate", m.service, "off").Run()
}
