package metrics

import (
	"runtime"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ustackq/indagate/pkg/version"
)

// indagateCollector define build info
type indagateCollector struct {
	indagateInfoDesc   *prometheus.Desc
	indagateUptimeDesc *prometheus.Desc
	start              time.Time
}

func NewIndagateCollector(name string, info version.Version) prometheus.Collector {

	return &indagateCollector{
		indagateInfoDesc: prometheus.NewDesc(
			"indagate_info",
			"indagate build info.",
			nil,
			prometheus.Labels{
				"version":   info.GitVersion,
				"commit":    info.GitCommit,
				"buildDate": info.BuildDate,
				"os":        info.Platform,
				"compiler":  info.Compiler,
				"cpus":      strconv.Itoa(runtime.NumCPU()),
			},
		),
		indagateUptimeDesc: prometheus.NewDesc(
			"indagate_uptime_seconds",
			"indagate start time in seconds",
			nil,
			prometheus.Labels{
				"service": name,
			},
		),
		start: time.Now(),
	}
}

func (c *indagateCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.indagateInfoDesc
	ch <- c.indagateUptimeDesc
}

func (c *indagateCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.indagateInfoDesc, prometheus.GaugeValue, 1)
	uptime := time.Since(c.start).Seconds()
	ch <- prometheus.MustNewConstMetric(c.indagateUptimeDesc, prometheus.GaugeValue, float64(uptime))
}
