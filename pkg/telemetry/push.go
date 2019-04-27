package telemetry

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/ustackq/indagate/pkg/metrics"
)

type Pusher struct {
	URL        string
	Gather     prometheus.Gatherer
	Client     *http.Client
	PushFormat expfmt.Format
}

func NewPusher(g prometheus.Gatherer) *Pusher {
	return &Pusher{
		URL: "https://ustack.io/metrics/job/indagate",
		Gather: &metrics.Filter{
			Gatherer: g,
			Matcher:  telemetryMatcher,
		},
		Client: &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   10 * time.Second,
		},
		PushFormat: expfmt.FmtText,
	}
}

func (p *Pusher) Push(ctx context.Context) error {
	if p.PushFormat == "" {
		p.PushFormat = expfmt.FmtText
	}
	res := make(chan error)
	go func() {
		res <- p.push(ctx)
	}()
	select {
	case err := <-res:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (p *Pusher) push(ctx context.Context) error {
	encode, err := p.encode()
	if err != nil {
		return err
	}

	if encode == nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, p.URL, encode)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", string(p.PushFormat))

	res, err := p.Client.Do(req)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err != nil {
		return err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		body, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("unable to POST metrics,recived status %s, %s", http.StatusText(res.StatusCode), body)
	}
	return nil
}

func (p *Pusher) encode() (io.Reader, error) {
	mfs, err := p.Gather.Gather()
	if err != nil {
		return nil, err
	}

	b, err := metrics.EncodeExpfmt(mfs, p.PushFormat)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(b), nil
}
