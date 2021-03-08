package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/goproxy/goproxy"
	"github.com/goproxy/goproxy/cacher"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

func main() {
	_ = godotenv.Load()

	app := cli.NewApp()
	app.Name = "goproxy"
	app.Usage = "Proxy server for go modules"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "listen",
			Value:   ":8080",
			Usage:   "listen address",
			EnvVars: []string{"LISTEN"},
		},
		&cli.StringFlag{
			Name:    "cache-dir",
			Usage:   "enable disk cache",
			EnvVars: []string{"CACHE_DIR"},
		},
		&cli.IntFlag{
			Name:    "cache-max-megabytes",
			Usage:   "Max number of megabytes to cache",
			EnvVars: []string{"CACHE_MAX_MEGABYTES"},
		},
	}
	var lis net.Listener
	app.Action = func(c *cli.Context) error {
		var err error
		lis, err = net.Listen("tcp", c.String("listen"))
		if err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}
		log.Printf("listening on: %v", lis.Addr())
		gp := goproxy.New()

		cacheDir := c.String("cache-dir")
		if cacheDir != "" {
			stat, err := os.Stat(cacheDir)
			if err != nil {
				return fmt.Errorf("unable to use the configured cache-dir: %v", err)
			}
			if !stat.IsDir() {
				return fmt.Errorf("unable to use the configured cache-dir: not a directory")
			}
			log.Printf("Using disk cache on: %s", cacheDir)
			gp.Cacher = &cacher.Disk{
				Root: cacheDir,
			}
			gp.CacherMaxCacheBytes = c.Int("cache-max-megabytes") * 1024 * 1024
		}

		err = http.Serve(lis, gp)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return fmt.Errorf("http serve error: %w", err)
		}
		return nil
	}

	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		log.Printf("Signal received: %s\n", sig)
		log.Println("Shutting down")
		err := lis.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}
	}()

	err := app.Run(os.Args)
	if err != nil {
		log.Println("error:", err)
		os.Exit(1)
	}
}
