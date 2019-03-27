package genesis

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/insolar/insolar/configuration"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/utils"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/instrumentation/instracer"
	"github.com/insolar/insolar/log"
	"github.com/insolar/insolar/version"
)

type Server struct {
	cfgPath           string
	trace             bool
	genesisConfigPath string
	genesisKeyOut     string
}

func New(cfgPath string, trace bool, genesisConfigPath, genesisKeyOut string) *Server {
	return &Server{
		cfgPath:           cfgPath,
		trace:             trace,
		genesisConfigPath: genesisConfigPath,
		genesisKeyOut:     genesisKeyOut,
	}
}

func (s *Server) Serve() {
	cfgHolder := configuration.NewHolder()
	var err error
	if len(s.cfgPath) != 0 {
		err = cfgHolder.LoadFromFile(s.cfgPath)
	} else {
		err = cfgHolder.Load()
	}
	if err != nil {
		log.Warn("failed to load configuration from file: ", err.Error())
	}

	cfg := &cfgHolder.Configuration
	cfg.Metrics.Namespace = "insolard"

	traceID := "main_" + utils.RandTraceID()
	ctx, inslog := initLogger(context.Background(), cfg.Log, traceID)
	log.SetGlobalLogger(inslog)
	fmt.Println("Starts with configuration:\n", configuration.ToString(cfgHolder.Configuration))

	removeLedgerDataDir(ctx, cfg)
	cfg.Ledger.PulseManager.HeavySyncEnabled = false

	bootstrapComponents := initBootstrapComponents(ctx, *cfg)
	certManager := initCertificateManager(
		ctx,
		*cfg,
		true,
		bootstrapComponents.CryptographyService,
		bootstrapComponents.KeyProcessor,
	)

	jaegerflush := func() {}
	if s.trace {
		jconf := cfg.Tracer.Jaeger
		log.Infof("Tracing enabled. Agent endpoint: '%s', collector endpoint: '%s'\n", jconf.AgentEndpoint, jconf.CollectorEndpoint)
		jaegerflush = instracer.ShouldRegisterJaeger(
			ctx,
			certManager.GetCertificate().GetRole().String(),
			certManager.GetCertificate().GetNodeRef().String(),
			jconf.AgentEndpoint,
			jconf.CollectorEndpoint,
			jconf.ProbabilityRate)
		ctx = instracer.SetBaggage(ctx, instracer.Entry{Key: "traceid", Value: traceID})
	}
	defer jaegerflush()

	cm, err := initComponents(
		ctx,
		*cfg,
		bootstrapComponents.CryptographyService,
		bootstrapComponents.PlatformCryptographyScheme,
		bootstrapComponents.KeyStore,
		bootstrapComponents.KeyProcessor,
		certManager,
		true,
		s.genesisConfigPath,
		s.genesisKeyOut,
	)
	checkError(ctx, err, "failed to init components")

	ctx, inslog = inslogger.WithField(ctx, "nodeid", certManager.GetCertificate().GetNodeRef().String())
	ctx, inslog = inslogger.WithField(ctx, "role", certManager.GetCertificate().GetRole().String())
	ctx = inslogger.SetLogger(ctx, inslog)
	log.SetGlobalLogger(inslog)

	err = cm.Init(ctx)
	checkError(ctx, err, "failed to init components")

	var gracefulStop = make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	var waitChannel = make(chan bool)

	go func() {
		sig := <-gracefulStop
		inslog.Debug("caught sig: ", sig)

		inslog.Warn("GRACEFULL STOP APP")
		err = cm.Stop(ctx)
		checkError(ctx, err, "failed to graceful stop components")
		close(waitChannel)
	}()

	err = cm.Start(ctx)
	checkError(ctx, err, "failed to start components")
	fmt.Println("Version: ", version.GetFullVersion())
	fmt.Println("All components were started")
	<-waitChannel
}

func initLogger(ctx context.Context, cfg configuration.Log, traceid string) (context.Context, insolar.Logger) {
	inslog, err := log.NewLog(cfg)
	if err != nil {
		panic(err)
	}

	if newInslog, err := inslog.WithLevel(cfg.Level); err != nil {
		inslog.Error(err.Error())
	} else {
		inslog = newInslog
	}

	ctx = inslogger.SetLogger(ctx, inslog)
	ctx, inslog = inslogger.WithTraceField(ctx, traceid)
	return ctx, inslog
}

func checkError(ctx context.Context, err error, message string) {
	if err == nil {
		return
	}
	inslogger.FromContext(ctx).Fatalf("%v: %v", message, err.Error())
}

func removeLedgerDataDir(ctx context.Context, cfg *configuration.Configuration) {
	_, err := exec.Command(
		"rm", "-rfv",
		cfg.Ledger.Storage.DataDirectory,
	).CombinedOutput()
	checkError(ctx, err, "failed to delete ledger storage data directory")
}
