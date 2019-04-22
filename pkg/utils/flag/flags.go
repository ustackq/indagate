package flag

import (
	"github.com/spf13/pflag"
	"k8s.io/klog"
)

func PrintFlags(flags *pflag.FlagSet) {
	flags.Visit(func(flag *pflag.Flag) {
		klog.V(1).Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})
}
