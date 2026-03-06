//go:build windows

package sysproxy

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

type WindowsManager struct {
	service string
}

func NewManager(service string) Manager {
	return &WindowsManager{service: service}
}

func (m *WindowsManager) SetProxy(addr, port string) error {
	fmt.Printf("sysproxy: enabling proxy on windows (%s:%s)\n", addr, port)
	k, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		registry.SET_VALUE,
	)
	if err != nil {
		return err
	}
	defer k.Close()

	if err := k.SetStringValue("ProxyServer", addr+":"+port); err != nil {
		return err
	}
	return k.SetDWordValue("ProxyEnable", 1)
}

func (m *WindowsManager) UnsetProxy() error {
	fmt.Printf("sysproxy: disabling proxy on windows\n")
	k, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		registry.SET_VALUE,
	)
	if err != nil {
		return err
	}
	defer k.Close()

	return k.SetDWordValue("ProxyEnable", 0)
}
