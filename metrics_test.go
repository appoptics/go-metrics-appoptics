package appoptics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecodeMetric(t *testing.T) {
	name, tags := decodeMetricName("myMetric")
	assert.Equal(t, "myMetric", name)
	assert.Equal(t, map[string]string{}, tags)

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
	assert.Equal(t, "myMetric#foo=1,bar=2", Metric("myMetric").Tag("foo", 1).Tag("bar", 2).String())

}
