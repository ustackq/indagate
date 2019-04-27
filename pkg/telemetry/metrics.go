package telemetry

import (
	"github.com/ustackq/indagate/pkg/metrics"
)

var telemetryMatcher = metrics.NewMacther().
	Family("indagate_info")
