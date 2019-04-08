This is a reporter for the [go-metrics](https://github.com/rcrowley/go-metrics)
library which posts the metrics to [AppOptics](https://www.appoptics.com/)/Librato. 
It is forked from [go-metrics-librato](https://github.com/mihasya/go-metrics-librato) and
was originally part of the `go-metrics` library itself.

This library supports tagged metrics only; source-based metrics have been deprecated by Librato.
The original [go-metrics-librato](https://github.com/mihasya/go-metrics-librato) library can be used
if you need to upload source-based metrics instead.

### Usage

```go
import "github.com/ysamlan/go-metrics-appoptics"

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

*Selective runtime metric uploading*: If you're using go-metrics' `CaptureRuntimeMemStats` feature,
it's great and automates collecting a lot of useful data. Unfortunately, it also adds 30 metrics, 
which can eat up a lot of metric hours with AppOptics. The `runtimeMetricsWhiteleist` parameter lets
you cherry-pick which metrics actually get uploaded, without needing to manually collect them 
yourself. See [the source](https://github.com/rcrowley/go-metrics/blob/master/runtime.go) for 
possible values. Pass `nil` to allow all, and an empty slice to disable uploads for all `runtime.` 
metrics.

### Limitations
Tags are attached at the batch level, not to the individual metrics/measurements within
the batch. You can work around this by using a different Registry and appoptics.AppOptics 
goroutine for metrics that need different tags. See 
[#1](https://github.com/ysamlan/go-metrics-appoptics/issues/1) for potential approaches for fixing
this.

### Migrating from `rcrowley/go-metrics` / `mihasya/go-metrics-librato` implementation

To get the same behavior you're used to from the original Librato reporter (only with tags instead
of sources):

* Change the import to `"github.com/ysamlan/go-metrics-appoptics"`
* Change `librato.Librato` to `appoptics.AppOptics`
* Remove the email argument from the `appoptics.AppOptics` function call (the updated AppOptics API
  only requires the token)
* Change the source argument from a `string` into a `map[string]string` for tags - e.g. a source
  `"myhostname"` could become the tag `map[string]string{"host":"myhostname"}`).
* Use `""` for the metric name prefix.
* Use `nil` for the runtime-metric-name whitelist (allow-all).