package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func newCollector(logger *zap.Logger) *collector {
	return &collector{
		logger: logger,
		messageCount: prometheus.NewCounter(prometheus.CounterOpts{
			Subsystem: "ont_watch_bot",
			Name:      "message_count",
			Help:      "how many requests performed bot",
		}),
	}
}

type collector struct {
	logger       *zap.Logger
	messageCount prometheus.Counter
}

func (c *collector) IncMessage() {
	c.messageCount.Inc()
}
func (c *collector) prometheusRegister() {
	metrics := [...]prometheus.Collector{
		c.messageCount,
	}
	for _, m := range metrics {
		if err := prometheus.Register(m); err != nil {
			c.logger.Error("failed to register in Prometheus", zap.Error(err))
		}
	}
}

func main() {
	sigKillCh := make(chan os.Signal, 1) // Kills the process immediately (not implemented)
	sigIntCh := make(chan os.Signal, 1)  // Stops the server gracefully
	sigHupCh := make(chan os.Signal, 1)  // Reloads server configuration file (not implemented)
	signal.Notify(sigKillCh, syscall.SIGKILL)
	signal.Notify(sigIntCh, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(sigHupCh, syscall.SIGHUP)

	debug := false
	debugEnv := os.Getenv("DEBUG")
	if "true" == debugEnv || "1" == debugEnv {
		debug = true
	}

	InitGlobalLogger(debug)

	logger := zap.L()
	collector := newCollector(logger)
	collector.prometheusRegister()

	logger.Info("starting otn watch bot")

	apiToken := os.Getenv("API_TOKEN")
	if "" == apiToken {
		logger.Info("You should expect API_TOKEN env variable")
		message, err := getData()
		logger.Info(message)
		if nil != err {
			logger.Info(err.Error())
		}
		return
	}

	bot, err := tgbotapi.NewBotAPI(apiToken)
	if nil != err {
		logger.Fatal("cannot create bot", zap.Error(err))
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Info("starting prometheus handler")
		_ = http.ListenAndServe(":9123", nil)
	}()

	bot.Debug = debug
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		logger.Fatal("cannot get update channel", zap.Error(err))
	}

	for {
		select {
		case update := <-updates:
			if nil == update.Message {
				continue
			}

			switch update.Message.Text {
			case "/rate":
				message, err := getData()

				if nil != err {
					logger.Error("cannot get exchange data", zap.Error(err))
					continue
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
				msg.ReplyToMessageID = update.Message.MessageID

				_, err = bot.Send(msg)
				if err != nil {
					logger.Error("cannot send message", zap.Error(err))
				}
				collector.IncMessage()
			}
		case <-sigIntCh:
			logger.Info("got SIGINT")
			return
		case <-sigHupCh:
			logger.Info("got SIGHUP")
			return
		case <-sigKillCh:
			logger.Info("got SIGKILL")
			return
		}
	}
}

func InitGlobalLogger(debug bool) {
	var zapConfig zap.Config
	if debug {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}
	zapConfig.DisableStacktrace = true
	zapConfig.DisableCaller = true
	zapConfig.EncoderConfig.LevelKey = "level"
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.MessageKey = "message"
	zapConfig.EncoderConfig.CallerKey = "caller"
	zapConfig.EncoderConfig.NameKey = "name"
	zapConfig.EncoderConfig.StacktraceKey = "stack"

	zapLogger, err := zapConfig.Build()
	if err != nil {
		panic("can not create logger: " + err.Error())
	}

	// After logger is configurated replace global logger
	zap.ReplaceGlobals(zapLogger)
}
