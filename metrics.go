package appoptics

import (
	"fmt"
	"github.com/rcrowley/go-metrics"
	"log"
	"regexp"
	"sort"
	"strings"
)

var metricNameRegex = regexp.MustCompile(`[^-.:_\w]`)
var tagNameRegex = regexp.MustCompile(`[^-.:_\w]`)
var tagValueRegex = regexp.MustCompile(`[^-.:_\\/\w\? ]`)

type TaggedMetric struct {
	name string
	tags map[string]string
	sampleFunc func() metrics.Sample
}

func Metric(name string) *TaggedMetric {
	return &TaggedMetric{name: sanitizeMetricName(name), tags: map[string]string{}}
}

func (t *TaggedMetric) Tag(name string, value interface{}) *TaggedMetric {
	tagName := sanitizeTagName(name)
	tagValue := sanitizeTagValue(fmt.Sprintf("%v", value))

	if tagName == "" || tagValue == "" {
		log.Printf("Empty tag name or value: name=%v value=%v", tagName, tagValue)
		return t
	}

	t.tags[tagName] = tagValue
	return t
}

func (t *TaggedMetric) WithSample(s func() metrics.Sample) *TaggedMetric {
	t.sampleFunc = s
	return t
}

func (t *TaggedMetric) String() string {
	sb := strings.Builder{}

	sb.WriteString(t.name)

	if len(t.tags) > 0 {
		sb.WriteString("#")
	}

	// Sort tag map for consistent ordering in encoded string
	var keys []string
	for key := range t.tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for i, key := range keys {
		if i != 0 {
			sb.WriteString(",")
		}
		sb.WriteString(key + "=" + t.tags[key])
	}

	return sb.String()
}

func (t *TaggedMetric) Counter() metrics.Counter {
	return metrics.GetOrRegisterCounter(t.String(), metrics.DefaultRegistry)
}

func (t *TaggedMetric) Meter() metrics.Meter {
	return metrics.GetOrRegisterMeter(t.String(), metrics.DefaultRegistry)
}

func (t *TaggedMetric) Timer() metrics.Timer {
	return metrics.GetOrRegisterTimer(t.String(), metrics.DefaultRegistry)
}

func (t *TaggedMetric) Histogram() metrics.Histogram {
	var sample func() metrics.Sample
	if t.sampleFunc != nil {
		sample = t.sampleFunc
	} else {
		sample = func() metrics.Sample {
			return metrics.NewExpDecaySample(1028, 0.015)
		}
	}

	return metrics.GetOrRegister(t.String(), func() metrics.Histogram {return metrics.NewHistogram(sample())}).(metrics.Histogram)
}

func (t *TaggedMetric) Gauge() metrics.Gauge {
	return metrics.GetOrRegisterGauge(t.String(), metrics.DefaultRegistry)
}

func (t *TaggedMetric) Gauge64() metrics.GaugeFloat64 {
	return metrics.GetOrRegisterGaugeFloat64(t.String(), metrics.DefaultRegistry)
}

// decodeMetricName decodes the metricName#a=foo,b=bar format and returns the TaggedMetric name
// as a string and the tags as a map
func decodeMetricName(encoded string) (string, map[string]string) {
	split := strings.SplitN(encoded, "#", 2)
	name := split[0]
	if len(split) == 1 {
		return name, nil
	}

	tagPart := split[1]
	pairs := strings.Split(tagPart, ",")

	tags := map[string]string{}
	for _, pair := range pairs {
		pairList := strings.SplitN(pair, "=", 2)
		if len(pairList) != 2 {
			log.Printf("Tag name `%v` is missing its value", pairList[0])
			continue
		}
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

func sanitizeMetricName(value string) string {
	if len(value) > 255 {
		value = value[:252] + "..."
	}
	return metricNameRegex.ReplaceAllString(value, "_")
}
