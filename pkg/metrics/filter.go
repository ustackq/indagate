package metrics

import (
	"sort"

	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// Filter filters the metrics from Gather using Matcher.
type Filter struct {
	Gatherer prometheus.Gatherer
	Matcher  Matcher
}

func (f *Filter) Gather() ([]*dto.MetricFamily, error) {
	mfs, err := f.Gatherer.Gather()
	if err != nil {
		return nil, err
	}
	return f.Matcher.Match(mfs), err
}

type Matcher map[string]Labels

func (m Matcher) Match(mfs []*dto.MetricFamily) []*dto.MetricFamily {
	if len(mfs) == 0 {
		return mfs
	}
	fF := []*dto.MetricFamily{}
	for _, mf := range mfs {
		labels, ok := m[mf.GetName()]
		if !ok {
			continue
		}
		metrics := []*dto.Metric{}
		match := false
		for _, metric := range mf.Metric {
			if labels.Match(metric) {
				match = true
				metrics = append(metrics, metric)
			}
		}

		if match {
			fF = append(fF, &dto.MetricFamily{
				Name:   mf.Name,
				Help:   mf.Help,
				Type:   mf.Type,
				Metric: metrics,
			})
		}
	}
	// sort
	sort.Sort(familySorter(fF))
	return fF
}

type Labels map[string]bool

// L is used with Family to create a series of label pairs for matching.
func L(name, value string) *dto.LabelPair {
	return &dto.LabelPair{
		Name:  proto.String(name),
		Value: proto.String(value),
	}
}

// Match checks if the metric's labels matches this set of labels.
func (ls Labels) Match(metric *dto.Metric) bool {
	lp := &labelPairs{metric.Label}
	return ls[lp.String()] || ls[""] // match empty string so no labels can be matched.
}

// labelPairs is used to serialize a portion of dto.Metric into a serializable
// string.
type labelPairs struct {
	Label []*dto.LabelPair `protobuf:"bytes,1,rep,name=label" json:"label,omitempty"`
}

func (l *labelPairs) Reset()         {}
func (l *labelPairs) String() string { return proto.CompactTextString(l) }
func (*labelPairs) ProtoMessage()    {}

// familySorter implements sort.Interface. It is used to sort a slice of
// dto.MetricFamily pointers.
type familySorter []*dto.MetricFamily

func (s familySorter) Len() int {
	return len(s)
}

func (s familySorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s familySorter) Less(i, j int) bool {
	return s[i].GetName() < s[j].GetName()
}

func NewMacther() Matcher {
	return Matcher{}
}

// Family helps constuct match by adding a metric family to match to.
func (m Matcher) Family(name string, lps ...*dto.LabelPair) Matcher {
	// prometheus metrics labels are sorted by label name.
	sort.Slice(lps, func(i, j int) bool {
		return lps[i].GetName() < lps[j].GetName()
	})

	pairs := &labelPairs{
		Label: lps,
	}

	family, ok := m[name]
	if !ok {
		family = make(Labels)
	}

	family[pairs.String()] = true
	m[name] = family
	return m
}
