package metrics

import (
	"bytes"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func EncodeExpfmt(mfs []*dto.MetricFamily, opts ...expfmt.Format) ([]byte, error) {
	format := expfmt.FmtProtoDelim
	if len(opts) != 0 && opts[0] != "" {
		format = opts[0]
	}

	buf := &bytes.Buffer{}
	encoder := expfmt.NewEncoder(buf, format)
	for _, mf := range mfs {
		if err := encoder.Encode(mf); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}
