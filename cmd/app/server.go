package app

import (
	"fmt"
	"net/http"

	"github.com/coreos/go-systemd/daemon"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/ustackq/indagate/cmd/app/options"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/route"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/ustackq/indagate/api"
)

const (
	component = "indagate"
)

var (
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "indagate_http_request_duration_seconds",
			Help:    "Histogram of latencies for HTTP requests.",
			Buckets: []float64{.05, 0.1, .25, .5, .75, 1, 2, 5, 20, 60},
		},
		[]string{"handler", "method"},
	)
	responseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "indagate_http_response_size_bytes",
			Help:    "Histogram of response size for HTTP requests.",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7),
		},
		[]string{"handler", "method"},
	)
	responseErr = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "reponse_error",
		Help: "1 if there was an error while request container metrics, 0 otherwise",
	})
)

func init() {
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(responseSize)
}

// NewIndagateCommand create  coomand object with default parameters
func NewIndagateCommand(stopCh <-chan struct{}, logger log.Logger) *cobra.Command {
	cleanFlagSet := pflag.NewFlagSet(component, pflag.ContinueOnError)
	indagateFlags := options.NewIndagateFlags()

	cmd := &cobra.Command{
		Use:  component,
		Long: ``,
		Run: func(cmd *cobra.Command, args []string) {
			help, err := cleanFlagSet.GetBool("help")
			if err != nil {
				level.Error(logger).Log("msg", "unable parse", "err", err)
			}
			if help {
				cmd.Help()
				return
			}

			// validate flag

			// run indagate server
			level.Info(logger).Log("msg", "run server")
			if err := run(indagateFlags, stopCh); err != nil {
				level.Error(logger).Log("msg", "unable run server", "err", err)
			}

		},
	}
	indagateFlags.AddFlags(cleanFlagSet)
	cleanFlagSet.BoolP("help", "h", false, fmt.Sprintf("help for %s", cmd.Name()))

	const usageFmt = "Usage:\n %s\n\nFlags:\n%s"
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		return nil
	})
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		//
	})

	return cmd
}

func run(ing *options.Indagate, stopCh <-chan struct{}) error {
	if err := options.ValidataIndagate(ing); err != nil {
		return nil
	}

	api, err := api.New(api.Options{
		Timeout:     ing.Timeout,
		Concurrency: ing.Concurrency,
		Logger:      log.With(ing.Logger, "indagate", "api"),
	})

	if err != nil {
		level.Error(ing.Logger).Log("err", fmt.Errorf("failed to create API: %v", err.Error()))
	}

	addr, err := parseStrToAddr(ing.Linsten)
	if err != nil {
		level.Error(ing.Logger).Log("err", err)
		return err
	}
	router := route.New().WithInstrumentation(instrumentHandler)

	mux := api.Registry(router)

	srv := http.Server{
		Addr:    addr,
		Handler: mux,
	}
	srcv := make(chan struct{})
	go func() {
		level.Info(ing.Logger).Log("msg", "Listening", "address", addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			level.Error(ing.Logger).Log("msg", "Listen error", "err", err)
			close(srcv)
		}

		defer func() {
			if err := srv.Close(); err != nil {
				level.Error(ing.Logger).Log("msg", "Error on closing the server ", "err", err)
			}
		}()
	}()

	done := make(chan struct{})
	// If systemd is used, notify it that we have started
	go daemon.SdNotify(false, "READY=1")

	select {
	case <-done:
		break
	case <-stopCh:
		break
	}

	return nil
}

func parseStrToAddr(listen string) (addr string, err error) {
	return "", nil
}

func instrumentHandler(handlerName string, handler http.HandlerFunc) http.HandlerFunc {
	label := prometheus.Labels{
		"handler": handlerName,
	}
	return promhttp.InstrumentHandlerDuration(
		requestDuration.MustCurryWith(label),
		promhttp.InstrumentHandlerResponseSize(
			responseSize.MustCurryWith(label),
			handler,
		),
	)
}
