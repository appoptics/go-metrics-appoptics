package appoptics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const Operations = "operations"
const OperationsShort = "ops"

type AppOpticsClient struct {
	Token string
}

// property strings
const (
	// display attributes
	Color             = "color"
	DisplayMax        = "display_max"
	DisplayMin        = "display_min"
	DisplayUnitsLong  = "display_units_long"
	DisplayUnitsShort = "display_units_short"
	DisplayStacked    = "display_stacked"
	DisplayTransform  = "display_transform"
	// special gauge display attributes
	SummarizeFunction = "summarize_function"
	Aggregate         = "aggregate"

	// metric keys
	Name        = "name"
	Period      = "period"
	Description = "description"
	DisplayName = "display_name"
	Attributes  = "attributes"

	// measurement keys
	Time  = "time"
	Tags  = "tags"
	Value = "value"

	// special gauge keys
	Count  = "count"
	Sum    = "sum"
	Max    = "max"
	Min    = "min"
	StdDev = "stddev"

	// batch keys
	Measurements = "measurements"

	MetricsPostUrl = "https://api.appoptics.com/v1/measurements"
)

type Measurement map[string]interface{}
type Metric map[string]interface{}

type Batch struct {
	Measurements []Measurement     `json:"measurements,omitempty"`
	Time         int64             `json:"time"`
	Tags         map[string]string `json:"tags"`
}

var client = http.DefaultClient

func SetHTTPClient(c *http.Client) {
	client = c
}

func (self *AppOpticsClient) PostMetrics(batch Batch) (err error) {
	var (
		js   []byte
		req  *http.Request
		resp *http.Response
	)

	if len(batch.Measurements) == 0 {
		return nil
	}

	if js, err = json.Marshal(batch); err != nil {
		return
	}

	if req, err = http.NewRequest("POST", MetricsPostUrl, bytes.NewBuffer(js)); err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(self.Token, "")

	if resp, err = client.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var body []byte
		if body, err = ioutil.ReadAll(resp.Body); err != nil {
			body = []byte(fmt.Sprintf("(could not fetch response body for error: %s)", err))
		}
		err = fmt.Errorf("Unable to post to AppOptics: %d %s %s", resp.StatusCode, resp.Status, string(body))
	}
	return
}
