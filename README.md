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
    10*time.Second,             // interval for uploads
    "token",                    // AppOptics API token
    map[string]string{
        "hostname": "localhost", // tags
    }, 
    []float64{0.95},            // percentiles to send
    time.Millisecond,           // time units for timers
    "myservicename.",            // prefix on reported metric names
)
```

### Limitations
Tags are attached at the batch level, not to the individual metrics/measurements within
the batch. You can work around this by using a different Registry and appoptics.AppOptics 
goroutine for metrics that need different tags. See 
[#1](https://github.com/ysamlan/go-metrics-appoptics/issues/1) for potential approaches for fixing
this.

### Migrating from `rcrowley/go-metrics` / `mihasya/go-metrics-librato` implementation

* Change the import to `"github.com/ysamlan/go-metrics-appoptics"`
* Change `librato.Librato` to `appoptics.AppOptics`
* Remove the email argument from the `appoptics.AppOptics` function call (the updated AppOptics API
  only requires the token)
* Change the `source` argument from a `string` into a `map[string]string` for tags - e.g. 
  `"myhostname"` would become `map[string]string{"source":"myhostname"}`).
* Add a prefix argument (use "" for none).
