package bolt

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ustackq/indagate/pkg/store"
	bolt "go.etcd.io/bbolt"
)

var _ prometheus.Collector = (*Client)(nil)

var (
	orgDesc = prometheus.NewDesc(
		"indagate_organizations_total",
		"Number of total organizations on the server",
		nil, nil)
)

func (c *Client) Describe(ch chan<- *prometheus.Desc) {
	ch <- orgDesc
}

func (c *Client) Collect(ch chan<- prometheus.Metric) {
	orgs := 0
	c.db.View(func(tx *bolt.Tx) error {
		orgs = tx.Bucket(store.OrgBucket).Stats().KeyN
		return nil
	})

	ch <- prometheus.MustNewConstMetric(
		orgDesc,
		prometheus.CounterValue,
		float64(orgs),
	)
}
