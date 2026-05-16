//go:build linux

package sysproxy

import (
	"fmt"
	"os/exec"
)

type LinuxManager struct {
}

func NewManager(service string) Manager {
	return &LinuxManager{}
}

func (m *LinuxManager) SetProxy(addr, port string) error {
	fmt.Printf("sysproxy: enabling proxy on (%s:%s)\n", addr, port)

	if err := exec.Command("gsettings", "set", "org.gnome.system.proxy.http", "host", addr).Run(); err != nil {
		return err
	}

	if err := exec.Command("gsettings", "set", "org.gnome.system.proxy.http", "port", port).Run(); err != nil {
		return err
	}

	if err := exec.Command("gsettings", "set", "org.gnome.system.proxy.https", "host", addr).Run(); err != nil {
		return err
	}
	
	if err := exec.Command("gsettings", "set", "org.gnome.system.proxy.https", "port", port).Run(); err != nil {
		return err
	}

	return exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "manual").Run()
}


func (m *LinuxManager) UnsetProxy() error {
	fmt.Println("sysproxy: disabling proxy on")

	return exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "none").Run()
}
