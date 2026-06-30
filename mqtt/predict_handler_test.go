package mqtt

import "testing"

func TestParsePredictPayloadFlatFields(t *testing.T) {
	payload := []byte(`{"sncode":"DEV001","baseTime":1717200000000,"rain1h":0.1,"rain12h":1.2,"rain24h":2.4}`)

	data, ok := ParsePredictPayload("rain/predict", payload)
	if !ok {
		t.Fatal("expected payload to be parsed")
	}
	if data.Sncode != "DEV001" {
		t.Fatalf("unexpected sncode: %s", data.Sncode)
	}
	if data.BaseTime != 1717200000000 {
		t.Fatalf("unexpected base time: %d", data.BaseTime)
	}
	if len(data.Predictions) != 3 {
		t.Fatalf("unexpected predictions count: %d", len(data.Predictions))
	}
	assertPredictItem(t, data.Predictions, 1, 0.1)
	assertPredictItem(t, data.Predictions, 12, 1.2)
	assertPredictItem(t, data.Predictions, 24, 2.4)
}

func TestParsePredictPayloadArray(t *testing.T) {
	payload := []byte(`{"snCode":"DEV002","time":1717200000,"data":[{"hour":1,"rain":0.2},{"hour":12,"rainfall":1.3},{"hour":24,"predictRain":2.5}]}`)

	data, ok := ParsePredictPayload("rain/predict", payload)
	if !ok {
		t.Fatal("expected payload to be parsed")
	}
	if data.Sncode != "DEV002" {
		t.Fatalf("unexpected sncode: %s", data.Sncode)
	}
	if data.BaseTime != 1717200000000 {
		t.Fatalf("unexpected base time: %d", data.BaseTime)
	}
	assertPredictItem(t, data.Predictions, 1, 0.2)
	assertPredictItem(t, data.Predictions, 12, 1.3)
	assertPredictItem(t, data.Predictions, 24, 2.5)
}

func TestParsePredictPayloadNestedMap(t *testing.T) {
	payload := []byte(`{"sncode":"DEV003","base_time":"2024-06-01 00:00:00","rain":{"1":0.3,"12":1.4,"24":2.6}}`)

	data, ok := ParsePredictPayload("rain/predict", payload)
	if !ok {
		t.Fatal("expected payload to be parsed")
	}
	if data.Sncode != "DEV003" {
		t.Fatalf("unexpected sncode: %s", data.Sncode)
	}
	assertPredictItem(t, data.Predictions, 1, 0.3)
	assertPredictItem(t, data.Predictions, 12, 1.4)
	assertPredictItem(t, data.Predictions, 24, 2.6)
}

func assertPredictItem(t *testing.T, items []predictItem, hour int, rain float64) {
	t.Helper()
	for _, item := range items {
		if item.Hour == hour {
			if item.Rain != rain {
				t.Fatalf("unexpected rain for hour %d: %v", hour, item.Rain)
			}
			return
		}
	}
	t.Fatalf("missing prediction hour %d", hour)
}
