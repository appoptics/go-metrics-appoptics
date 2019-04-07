package appoptics

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/rcrowley/go-metrics"
)

// a regexp for extracting the unit from time.Duration.String
var unitRegexp = regexp.MustCompile("[^\\d]+$")

// a helper that turns a time.Duration into AppOptics display attributes for timer metrics
func translateTimerAttributes(d time.Duration) (attrs map[string]interface{}) {
	attrs = make(map[string]interface{})
	attrs[DisplayTransform] = fmt.Sprintf("x/%d", int64(d))
	attrs[DisplayUnitsShort] = string(unitRegexp.Find([]byte(d.String())))
	return
}

type Reporter struct {
	Token                     string
	Tags                      map[string]string
	Interval                  time.Duration
	Registry                  metrics.Registry
	Percentiles               []float64              // percentiles to report on histogram metrics
	Prefix                    string                 // prefix metric names for upload (eg "servicename.")
	WhitelistedRuntimeMetrics map[string]bool        // runtime.* metrics to upload (nil = allow all)
	TimerAttributes           map[string]interface{} // units in which timers will be displayed
	intervalSec               int64
}

func NewReporter(registry metrics.Registry, interval time.Duration, token string, tags map[string]string,
	percentiles []float64, timeUnits time.Duration, prefix string, whitelistedRuntimeMetrics []string) *Reporter {
	// set up lookups for our whitelist. Translate from []string to map[string]bool for easy lookups
	// nil = allow all; empty slice = block all
	var whitelist map[string]bool
	if whitelistedRuntimeMetrics != nil {
		whitelist = map[string]bool{}
		for _, name := range whitelistedRuntimeMetrics {
			whitelist[name] = true
		}
	}

	return &Reporter{token, tags, interval, registry, percentiles, prefix,
		whitelist, translateTimerAttributes(timeUnits),
		int64(interval / time.Second)}
}

// Call in a goroutine to start metric uploading.
// Using whitelistedRuntimeMetrics: a non-nil value sets this reporter to upload only a subset
// of the runtime.* metrics that are gathered by go-metrics runtime memory stats
// (CaptureRuntimeMemStats). The full list of possible values is at
// https://github.com/rcrowley/go-metrics/blob/master/runtime.go#L181-L211
// Passing an empty slice disables uploads for all runtime.* metrics.
func AppOptics(registry metrics.Registry, interval time.Duration, token string, tags map[string]string,
	percentiles []float64, timeUnits time.Duration, prefix string, whitelistedRuntimeMetrics []string) {
	NewReporter(registry, interval, token, tags, percentiles, timeUnits, prefix, whitelistedRuntimeMetrics).Run()
}

func (self *Reporter) Run() {
	ticker := time.Tick(self.Interval)
	metricsApi := &AppOpticsClient{self.Token}
	for now := range ticker {
		var metrics Batch
		var err error
		if metrics, err = self.BuildRequest(now, self.Registry); err != nil {
			log.Printf("ERROR constructing AppOptics request body %s", err)
			continue
		}
		if err := metricsApi.PostMetrics(metrics); err != nil {
			log.Printf("ERROR sending metrics to AppOptics %s", err)
			continue
		}
	}
}

func (self *Reporter) BuildRequest(now time.Time, r metrics.Registry) (snapshot Batch, err error) {
	snapshot = Batch{
		// coerce timestamps to a stepping fn so that they line up in AppOptics graphs
		Time: (now.Unix() / self.intervalSec) * self.intervalSec,
		Tags: self.Tags,
	}
	snapshot.Measurements = make([]Measurement, 0)
	histogramMeasurementCount := 1 + len(self.Percentiles)
	r.Each(func(name string, metric interface{}) {
		// if whitelis is set (non-nil), only upload runtime.* metrics from the list
		if strings.HasPrefix(name, "runtime.") && self.WhitelistedRuntimeMetrics != nil &&
			!self.WhitelistedRuntimeMetrics[name] {
			return
		}

		name = self.Prefix + name
		measurement := Measurement{}
		measurement[Period] = self.Interval.Seconds()
		switch m := metric.(type) {
		case metrics.Counter:
			if m.Count() > 0 {
				measurement[Name] = fmt.Sprintf("%s.%s", name, "count")
				measurement[Value] = float64(m.Count())
				measurement[Attributes] = map[string]interface{}{
					DisplayUnitsLong:  Operations,
					DisplayUnitsShort: OperationsShort,
					DisplayMin:        "0",
				}
				snapshot.Measurements = append(snapshot.Measurements, measurement)
			}
		case metrics.Gauge:
			measurement[Name] = name
			measurement[Value] = float64(m.Value())
			snapshot.Measurements = append(snapshot.Measurements, measurement)
		case metrics.GaugeFloat64:
			measurement[Name] = name
			measurement[Value] = float64(m.Value())
			snapshot.Measurements = append(snapshot.Measurements, measurement)
		case metrics.Histogram:
			if m.Count() > 0 {
				measurements := make([]Measurement, histogramMeasurementCount, histogramMeasurementCount)
				s := m.Sample()
				measurement[Name] = fmt.Sprintf("%s.%s", name, "hist")
				// For AppOptics, count must be the number of measurements in this sample. It will show sum/count as the mean.
				// Sample.Size() gives us this. Sample.Count() gives the total number of measurements ever recorded for the
				// life of the histogram, which means the AppOptics graph will trend toward 0 as more measurements are recored.
				measurement[Count] = uint64(s.Size())
				measurement[Max] = float64(s.Max())
				measurement[Min] = float64(s.Min())
				measurement[Sum] = float64(s.Sum())
				measurement[StdDev] = float64(s.StdDev())
				measurements[0] = measurement
				for i, p := range self.Percentiles {
					measurements[i+1] = Measurement{
						Name:   fmt.Sprintf("%s.%.2f", measurement[Name], p),
						Value:  s.Percentile(p),
						Period: measurement[Period],
					}
				}
				snapshot.Measurements = append(snapshot.Measurements, measurements...)
			}
		case metrics.Meter:
			measurement[Name] = name
			measurement[Value] = float64(m.Count())
			snapshot.Measurements = append(snapshot.Measurements, measurement)
			snapshot.Measurements = append(snapshot.Measurements,
				Measurement{
					Name:   fmt.Sprintf("%s.%s", name, "1min"),
					Value:  m.Rate1(),
					Period: int64(self.Interval.Seconds()),
					Attributes: map[string]interface{}{
						DisplayUnitsLong:  Operations,
						DisplayUnitsShort: OperationsShort,
						DisplayMin:        "0",
					},
				},
				Measurement{
					Name:   fmt.Sprintf("%s.%s", name, "5min"),
					Value:  m.Rate5(),
					Period: int64(self.Interval.Seconds()),
					Attributes: map[string]interface{}{
						DisplayUnitsLong:  Operations,
						DisplayUnitsShort: OperationsShort,
						DisplayMin:        "0",
					},
				},
				Measurement{
					Name:   fmt.Sprintf("%s.%s", name, "15min"),
					Value:  m.Rate15(),
					Period: int64(self.Interval.Seconds()),
					Attributes: map[string]interface{}{
						DisplayUnitsLong:  Operations,
						DisplayUnitsShort: OperationsShort,
						DisplayMin:        "0",
					},
				},
			)
		case metrics.Timer:
			measurement[Name] = name
			measurement[Value] = float64(m.Count())
			snapshot.Measurements = append(snapshot.Measurements, measurement)
			if m.Count() > 0 {
				appOpticsName := fmt.Sprintf("%s.%s", name, "timer.mean")
				measurements := make([]Measurement, histogramMeasurementCount, histogramMeasurementCount)
				measurements[0] = Measurement{
					Name:       appOpticsName,
					Count:      uint64(m.Count()),
					Sum:        m.Mean() * float64(m.Count()),
					Max:        float64(m.Max()),
					Min:        float64(m.Min()),
					StdDev:     float64(m.StdDev()),
					Period:     int64(self.Interval.Seconds()),
					Attributes: self.TimerAttributes,
				}
				for i, p := range self.Percentiles {
					measurements[i+1] = Measurement{
						Name:       fmt.Sprintf("%s.timer.%2.0f", name, p*100),
						Value:      m.Percentile(p),
						Period:     int64(self.Interval.Seconds()),
						Attributes: self.TimerAttributes,
					}
				}
				snapshot.Measurements = append(snapshot.Measurements, measurements...)
				snapshot.Measurements = append(snapshot.Measurements,
					Measurement{
						Name:   fmt.Sprintf("%s.%s", name, "rate.1min"),
						Value:  m.Rate1(),
						Period: int64(self.Interval.Seconds()),
						Attributes: map[string]interface{}{
							DisplayUnitsLong:  Operations,
							DisplayUnitsShort: OperationsShort,
							DisplayMin:        "0",
						},
					},
					Measurement{
						Name:   fmt.Sprintf("%s.%s", name, "rate.5min"),
						Value:  m.Rate5(),
						Period: int64(self.Interval.Seconds()),
						Attributes: map[string]interface{}{
							DisplayUnitsLong:  Operations,
							DisplayUnitsShort: OperationsShort,
							DisplayMin:        "0",
						},
					},
					Measurement{
						Name:   fmt.Sprintf("%s.%s", name, "rate.15min"),
						Value:  m.Rate15(),
						Period: int64(self.Interval.Seconds()),
						Attributes: map[string]interface{}{
							DisplayUnitsLong:  Operations,
							DisplayUnitsShort: OperationsShort,
							DisplayMin:        "0",
						},
					},
				)
			}
		}
	})
	return
}
