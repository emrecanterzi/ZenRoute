//go:build darwin

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

	if err := exec.Command("networksetup", "-setsecurewebproxy", m.service, addr, port).Run(); err != nil {
		return err
	}

	proxyURL := fmt.Sprintf("http://%s:%s", addr, port)
	exec.Command("launchctl", "setenv", "HTTPS_PROXY", proxyURL).Run()
	exec.Command("launchctl", "setenv", "HTTP_PROXY", proxyURL).Run()

	return nil
}

func (m *DarwinManager) UnsetProxy() error {
	fmt.Printf("sysproxy: disabling proxy on %s\n", m.service)
	if err := exec.Command("networksetup", "-setwebproxystate", m.service, "off").Run(); err != nil {
		return err
	}

	exec.Command("launchctl", "unsetenv", "HTTPS_PROXY").Run()
	exec.Command("launchctl", "unsetenv", "HTTP_PROXY").Run()

	return exec.Command("networksetup", "-setsecurewebproxystate", m.service, "off").Run()
}
