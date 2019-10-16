This is a reporter for the [go-metrics](https://github.com/rcrowley/go-metrics)
library which posts the metrics to [AppOptics](https://www.appoptics.com/).
It is based on [ysamlan's AppOptics reporter](https://github.com/ysamlan/go-metrics-appoptics),
which was based on [mihaysa's Librato reporter](https://github.com/mihasya/go-metrics-librato).

### Usage

```go
import "github.com/appoptics/go-metrics-appoptics"

go appoptics.AppOptics(metrics.DefaultRegistry,
    10*time.Second,              // interval for uploads
    "token",                     // AppOptics API token
    map[string]string{
        "hostname": "localhost", // tags
    }, 
    []float64{0.95},             // percentiles to send
    time.Millisecond,            // time units for timers
    "myservicename.",            // prefix on reported metric names
    nil,                         // (optional) go-metrics runtime.* stats upload whitelist
)
```

### Features

*Metric Name Prefix*: This reporter supports a `prefix` argument when initializing. All uploaded 
metrics will have that prefix prepended to their names. Use `""` if you don't want this behavior.

*Tags*: Tags passed during the initialization are attached to all this reporter's measurements to
AppOptics.

Tags can also be created on a per-metric level using the [`.Tag()`](https://godoc.org/github.com/appoptics/go-metrics-appoptics#TaggedMetric.Tag) function, for example:
```go
appoptics.Metric("myMetric").Tag("tagA", "foo").Tag("tagB", "bar").Meter().Mark()
```
Or, to create a histogram with a custom sample type:
```go
appoptics.Metric("myMetric").Tag("tag", "foo").WithSample(func() metrics.Sample { return metrics.NewUniformSample(1000) }).Histogram().Update(100)
```

*Selective runtime metric uploading*: If you're using go-metrics' `CaptureRuntimeMemStats` feature,
it's great and automates collecting a lot of useful data. Unfortunately, it also adds 30 metrics, 
which can eat up a lot of metric hours with AppOptics. The `runtimeMetricsWhiteleist` parameter lets
you cherry-pick which metrics actually get uploaded, without needing to manually collect them 
yourself. See [the source](https://github.com/rcrowley/go-metrics/blob/master/runtime.go) for 
possible values. Pass `nil` to allow all, and an empty slice to disable uploads for all `runtime.` 
metrics.

### Migrating from Librato and the `rcrowley/go-metrics` / `mihasya/go-metrics-librato` implementation

Source-based metrics are not supported in AppOptics. To migrate from the old Librato reporter (only with tags instead of sources):

* Change the import to `"github.com/appoptics/go-metrics-appoptics"`
* Change `librato.Librato` to `appoptics.AppOptics`
* Remove the email argument from the `appoptics.AppOptics` function call (the updated AppOptics API
  only requires the token)
* Change the source argument from a `string` into a `map[string]string` for tags - e.g. a source
  `"myhostname"` could become the tag `map[string]string{"host":"myhostname"}`).
* Use `""` for the metric name prefix.
* Use `nil` for the runtime-metric-name whitelist (allow-all).
