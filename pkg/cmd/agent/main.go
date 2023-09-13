package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/promlog"
	promlogflag "github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"prober/pkg/agent"
)

var (
	app = kingpin.New(filepath.Base(os.Args[0]), "The FITI-agent")
	//You should modify it to the address of the Server (Controller) in Default()
	grpcServerAddress = app.Flag("grpc.server-address", "server addr").Default("").String()
)

func main() {

	promlogConfig := promlog.Config{}

	app.Version(version.Print("agent"))
	app.HelpFlag.Short('h')
	promlogflag.AddFlags(app, &promlogConfig)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	var logger log.Logger
	logger = func(config *promlog.Config) log.Logger {
		var (
			l  log.Logger
			le level.Option
		)
		if config.Format.String() == "logfmt" {
			l = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
		} else {
			l = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
		}

		switch config.Level.String() {
		case "debug":
			le = level.AllowDebug()
		case "info":
			le = level.AllowInfo()
		case "warn":
			le = level.AllowWarn()
		case "error":
			le = level.AllowError()
		}
		l = level.NewFilter(l, le)
		l = log.With(l, "ts", log.TimestampFormat(
			func() time.Time { return time.Now().Local() },
			"2006-01-02T15:04:05.000Z07:00",
		), "caller", log.DefaultCaller)
		return l
	}(&promlogConfig)

	if ipSucc := agent.GetLocalIp(logger); ipSucc == false {
		level.Error(logger).Log("msg", "failed_to_get_ip_exit...")
		return
	}
	if nodeNameSucc := agent.GetLocalNodeName(logger); nodeNameSucc == false {
		level.Error(logger).Log("msg", "failed_to_get_node_name_exit...")
		return
	}

	level.Info(logger).Log("msg", "agent_metadata", "ip", agent.LocalIp, "region", agent.LocalNodeName)
	// init rpc pool
	//ctx, cancelAll := context.WithCancel(context.Background())
	isSuccess := agent.InitRpcPool(*grpcServerAddress, logger)
	if isSuccess == false {
		level.Error(logger).Log("msg", "init_rpc_pool_failed_and_exit")
		os.Exit(1)
	}
	level.Info(logger).Log("msg", "init_rpc_pool_success")
	// report ip

	go agent.ReportIp(logger)

	//定期更新pinglist
	agent.Init(logger)
	go agent.RefreshTarget(logger)

	//上传结果
	go agent.PushWork(logger)

	// term handler
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-term:
			level.Info(logger).Log("msg", "Received SIGTERM, exiting gracefully...")
			agent.GrpcPool.Close()
			return
		}
	}

}
