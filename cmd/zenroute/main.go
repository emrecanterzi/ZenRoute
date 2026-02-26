package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/emrecanterzi/internal/cache"
	"github.com/emrecanterzi/internal/config"
	"github.com/emrecanterzi/internal/dns"
	"github.com/emrecanterzi/internal/proxy"
	"github.com/emrecanterzi/internal/sysproxy"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("zenroute: config error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("zenroute: starting engine (os: %s)\n", runtime.GOOS)

	if runtime.GOOS != "darwin" {
		fmt.Printf("zenroute: unsupported os: %s\n", runtime.GOOS)
		os.Exit(1)
	}

	cache := cache.NewInMemoryCache()
	resolver := dns.NewCloudflareDoH(cache)
	server := proxy.NewServer(proxy.Options{
		Addr:          fmt.Sprintf("%s:%s", cfg.ProxyAddr, cfg.ProxyPort),
		FragmentSize:  cfg.FragmentSize,
		BypassDomains: cfg.BypassDomains,
		BypassAll:     cfg.BypassAll,
	}, resolver)

	sysMgr := sysproxy.NewManager(cfg.SystemServiceName)

	ctx, cancel := context.WithCancel(context.Background())

	if err := sysMgr.SetProxy(cfg.ProxyAddr, cfg.ProxyPort); err != nil {
		fmt.Printf("zenroute: failed to set system proxy: %v\n", err)
		os.Exit(1)
	}

	if cfg.BypassAll {
		fmt.Printf("zenroute: bypass all enabled\n")
	} else {
		fmt.Printf("zenroute: bypass domains: %v\n", len(cfg.BypassDomains))
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
