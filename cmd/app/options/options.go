package options

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/spf13/pflag"
)

// Indagate contains configuration flags for the Indagate.
type Indagate struct {
	Config      string
	Timeout     time.Duration
	Concurrency int
	Logger      log.Logger
	Linsten     string
}

// NewIndagateFlags will create a new IndagateFlags with default values.
func NewIndagateFlags() *Indagate {
	return &Indagate{}
}

// ValidataIndagate validate Indagate flag config
func ValidataIndagate(ing *Indagate) error {
	return nil
}

// AddFlags adds flag
func (f *Indagate) AddFlags(mainfs *pflag.FlagSet) {
	fs := pflag.NewFlagSet("", pflag.ExitOnError)
	defer func() {
		mainfs.AddFlagSet(fs)
	}()

	fs.StringVar(&f.Config, "config", f.Config, "The Indagate config file")
}
