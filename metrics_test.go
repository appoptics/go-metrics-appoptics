package appoptics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecodeMetric(t *testing.T) {
	name, tags := decodeMetricName("myMetric")
	assert.Equal(t, "myMetric", name)
	assert.Equal(t, 0, len(tags))

	name, tags = decodeMetricName("myMetric#foo=1")
	assert.Equal(t, "myMetric", name)
	assert.Equal(t, map[string]string{"foo": "1"}, tags)

	name, tags = decodeMetricName("myMetric#foo=1,bar=2")
	assert.Equal(t, "myMetric", name)
	assert.Equal(t, map[string]string{"foo": "1", "bar": "2"}, tags)
}

func TestEncodeMetric(t *testing.T) {
	assert.Equal(t, "myMetric", Metric("myMetric").String())
	assert.Equal(t, "myMetric#foo=1", Metric("myMetric").Tag("foo", 1).String())
	assert.Equal(t, "myMetric#bar=2,foo=1", Metric("myMetric").Tag("foo", 1).Tag("bar", 2).String())
}

func TestTagSanitation(t *testing.T) {
	metric := Metric("myMetric")
	assert.Equal(t, "myMetric#foo=1_2", metric.Tag("foo", "1&2").String())

	metric = Metric("myMetric")
	assert.Equal(t, "myMetric#f_o=1", metric.Tag("f=o", "1").String())
}

func TestMetricSanitation(t *testing.T) {
	metric := Metric("myMetric#")
	assert.Equal(t, "myMetric_", metric.String())

	metric = Metric("myMetric#").Tag("foo", "bar")
	assert.Equal(t, "myMetric_#foo=bar", metric.String())
}
