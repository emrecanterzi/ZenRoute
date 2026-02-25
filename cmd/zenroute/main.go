package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/emrecanterzi/internal/config"
	"github.com/emrecanterzi/internal/dns"
	"github.com/emrecanterzi/internal/proxy"
	"github.com/emrecanterzi/internal/sysproxy"
)

func main() {
	cfg := config.Load()

	fmt.Printf("zenroute: starting engine (os: %s)\n", runtime.GOOS)

	if runtime.GOOS != "darwin" {
		fmt.Printf("zenroute: unsupported os: %s\n", runtime.GOOS)
		os.Exit(1)
	}

	resolver := dns.NewCloudflareDoH()
	server := proxy.NewServer(fmt.Sprintf("%s:%s", cfg.ProxyAddr, cfg.ProxyPort), resolver, cfg.FragmentSize)

	sysMgr := sysproxy.NewManager(cfg.SystemServiceName)

	ctx, cancel := context.WithCancel(context.Background())

	if err := sysMgr.SetProxy(cfg.ProxyAddr, cfg.ProxyPort); err != nil {
		fmt.Printf("zenroute: failed to set system proxy: %v\n", err)
		os.Exit(1)
	}

	go func() {
		if err := server.Start(ctx); err != nil {
			fmt.Printf("zenroute: proxy server err: %v\n", err)
			cancel()
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("zenroute: stopping...")
	cancel()
	sysMgr.UnsetProxy()
	fmt.Println("zenroute: exited")
}
