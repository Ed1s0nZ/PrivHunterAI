package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lqqyt2423/go-mitmproxy/proxy"
	"yuequanScan/internal/config"
	"yuequanScan/internal/httpapi"
	"yuequanScan/internal/scanner"
	"yuequanScan/internal/store"
)

func main() {
	addr := flag.String("listen", "127.0.0.1:8222", "控制台监听地址")
	dbPath := flag.String("database", "data/privhunter.db", "SQLite 数据库文件")
	frontend := flag.String("frontend", "frontend/dist", "React 构建目录")
	secure := flag.Bool("secure-cookie", false, "HTTPS 部署使用 Secure Cookie")
	proxyAddr := flag.String("proxy", "127.0.0.1:9080", "被动代理监听地址")
	importPath := flag.String("import-config", "", "显式导入旧配置到本地 SQLite（保持暂停）")
	flag.Parse()
	gin.SetMode(gin.ReleaseMode)
	db, e := store.Open(*dbPath)
	if e != nil {
		log.Fatal("数据库初始化失败")
	}
	defer db.Close()
	if *importPath != "" {
		settings, err := config.ImportLegacy(*importPath)
		if err != nil {
			log.Fatal(err)
		}
		if err = db.SaveSettings(settings); err != nil {
			log.Fatal("旧配置保存失败")
		}
		log.Print("旧配置已本地导入，扫描保持暂停")
	}
	engine := scanner.New(db)
	apiServer := httpapi.New(db, *secure)
	apiServer.ScannerStatus = engine.Status
	router := apiServer.Router()
	router.Static("/assets", *frontend+"/assets")
	router.GET("/", func(c *gin.Context) { c.File(*frontend + "/index.html") })
	server := &http.Server{Addr: *addr, Handler: router, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	engine.Run(ctx)
	passive, err := proxy.NewProxy(&proxy.Options{Addr: *proxyAddr, StreamLargeBodies: 1 << 20})
	if err != nil {
		log.Fatal("代理初始化失败")
	}
	passive.AddAddon(engine)
	go func() {
		if err := passive.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Print("代理已停止")
			stop()
		}
	}()
	go func() {
		log.Printf("PrivHunter 控制台监听 %s", *addr)
		if e := server.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
			log.Print("HTTP 服务启动失败")
			stop()
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	passive.Close()
	engine.Wait()
	if server.Shutdown(shutdown) != nil {
		log.Print("HTTP 服务关闭超时")
	}
}
