package appoptics

import (
	"fmt"
	"github.com/rcrowley/go-metrics"
	"regexp"
	"strings"
)

var tagNameRegex = regexp.MustCompile(`[^-.:_\w]`)
var tagValueRegex = regexp.MustCompile(`[^-.:_\\/\w\? ]`)

type metric struct {
	name string
	tags map[string]string
	sample *metrics.Sample
}

func Metric(name string) *metric {
	return &metric{name: name}
}

func (m *metric) Tag(name string, value interface{}) *metric {
	tagName := sanitizeTagName(fmt.Sprintf("%v", name))
	tagValue := sanitizeTagValue(fmt.Sprintf("%v", value))
	m.tags[tagName] = tagValue
	return m
}

func (m *metric) Sample(s metrics.Sample) *metric {
	m.sample = &s
	return m
}

func (m *metric) String() string {
	sb := strings.Builder{}

	sb.WriteString(m.name + "#")

	// TODO: sort, for consistent ordering
	for name, value := range m.tags {
		sb.WriteString(name + "=" + value + ",")
	}

	// Hacky way to remove trailing comma
	return strings.TrimSuffix(sb.String(), ",")
}

func (m *metric) Counter() metrics.Counter {
	return metrics.GetOrRegisterCounter(m.String(), metrics.DefaultRegistry)
}

func (m *metric) Meter() metrics.Meter {
	return metrics.GetOrRegisterMeter(m.String(), metrics.DefaultRegistry)
}

func (m *metric) Timer() metrics.Timer {
	return metrics.GetOrRegisterTimer(m.String(), metrics.DefaultRegistry)
}

func (m *metric) Histogram() metrics.Histogram {
	sample := func(s *metrics.Sample) metrics.Sample {
		if s == nil {
			return metrics.NewExpDecaySample(1028, 0.015)
		} else {
			return *s
		}
	}

	return metrics.GetOrRegister(m.String(), sample(m.sample)).(metrics.Histogram)
}

func (m *metric) Gauge() metrics.Gauge {
	return metrics.GetOrRegisterGauge(m.String(), metrics.DefaultRegistry)
}

func (m *metric) Gauge64() metrics.GaugeFloat64 {
	return metrics.GetOrRegisterGaugeFloat64(m.String(), metrics.DefaultRegistry)
}

// decodeMetricName decodes the metricName#a=foo,b=bar format and returns the metric name
// as a string and the tags as a map
func decodeMetricName(encoded string) (string, map[string]string) {
	split := strings.SplitN(encoded, "#", 2)
	name := split[0]
	tagPart := split[1]
	pairs := strings.Split(tagPart, ",")

	tags := map[string]string{}
	for _, pair := range pairs {
		pairList := strings.SplitN(pair, "=", 2)
		tags[pairList[0]] = pairList[1]
	}

	return name, tags
}

func sanitizeTagName(value string) string {
	if len(value) > 64 {
		value = value[:64]
	}
	value = strings.ToLower(value)
	return tagNameRegex.ReplaceAllString(value, "_")
}

func sanitizeTagValue(value string) string {
	if len(value) > 255 {
		value = value[:252] + "..."
	}
	value = strings.ToLower(value)
	return tagValueRegex.ReplaceAllString(value, "_")
}
