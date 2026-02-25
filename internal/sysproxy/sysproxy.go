package sysproxy

type Manager interface {
	SetProxy(addr, port string) error
	UnsetProxy() error
}
