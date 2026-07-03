package services

import (
	"encoding/binary"
	"testing"
	"time"

	"nav-rain-grid-go/domains"
)

type testNCVariableHeader struct {
	name  string
	vsize uint32
	begin uint32
}

func TestBuildGridDiffNCDataVariableOffsets(t *testing.T) {
	baseTime := time.Date(2026, 7, 2, 7, 0, 0, 0, time.Local).UnixMilli()
	result := domains.GridDiffResult{
		GridName: "寿命格网",
		BaseTime: baseTime,
		Points: []domains.GridDiffPointResult{
			testGridDiffPointResult(114.01, 30.01, 1, 12, 24),
			testGridDiffPointResult(114.02, 30.02, 2, 13, 25),
		},
	}

	data, err := buildGridDiffNCData(result, completeGridDiffPoints(result.Points))
	if err != nil {
		t.Fatalf("build nc data: %v", err)
	}

	variables := readTestNCVariables(t, data)
	if len(variables) != 5 {
		t.Fatalf("unexpected variable count: %d", len(variables))
	}
	expectedNames := []string{"longitude", "latitude", "forecast_1h", "forecast_12h", "forecast_24h"}
	for i, variable := range variables {
		if variable.name != expectedNames[i] {
			t.Fatalf("unexpected variable %d: got %s, want %s", i, variable.name, expectedNames[i])
		}
		if variable.vsize != 16 {
			t.Fatalf("unexpected vsize for %s: %d", variable.name, variable.vsize)
		}
		if int(variable.begin+variable.vsize) > len(data) {
			t.Fatalf("variable %s exceeds file size: begin=%d vsize=%d len=%d", variable.name, variable.begin, variable.vsize, len(data))
		}
	}
}

func testGridDiffPointResult(lng, lat, rain1H, rain12H, rain24H float64) domains.GridDiffPointResult {
	return domains.GridDiffPointResult{
		CenterLng:   lng,
		CenterLat:   lat,
		Forecast1H:  &domains.GridDiffForecast{PredictRain: rain1H},
		Forecast12H: &domains.GridDiffForecast{PredictRain: rain12H},
		Forecast24H: &domains.GridDiffForecast{PredictRain: rain24H},
	}
}

func readTestNCVariables(t *testing.T, data []byte) []testNCVariableHeader {
	t.Helper()
	offset := 0
	if len(data) < 8 || string(data[:4]) != "CDF\x01" {
		t.Fatalf("invalid nc magic")
	}
	offset += 8

	dimTag := readTestNCInt32(t, data, &offset)
	if dimTag != ncTagDimension {
		t.Fatalf("unexpected dimension tag: %d", dimTag)
	}
	dimCount := readTestNCInt32(t, data, &offset)
	for i := int32(0); i < dimCount; i++ {
		_ = readTestNCString(t, data, &offset)
		_ = readTestNCInt32(t, data, &offset)
	}

	skipTestNCAttributes(t, data, &offset)

	varTag := readTestNCInt32(t, data, &offset)
	if varTag != ncTagVariable {
		t.Fatalf("unexpected variable tag: %d", varTag)
	}
	varCount := readTestNCInt32(t, data, &offset)
	result := make([]testNCVariableHeader, 0, varCount)
	for i := int32(0); i < varCount; i++ {
		name := readTestNCString(t, data, &offset)
		dimIDCount := readTestNCInt32(t, data, &offset)
		offset += int(dimIDCount) * 4
		skipTestNCAttributes(t, data, &offset)
		ncType := readTestNCInt32(t, data, &offset)
		if ncType != ncDouble {
			t.Fatalf("unexpected variable type for %s: %d", name, ncType)
		}
		vsize := readTestNCUint32(t, data, &offset)
		begin := readTestNCUint32(t, data, &offset)
		result = append(result, testNCVariableHeader{name: name, vsize: vsize, begin: begin})
	}
	return result
}

func skipTestNCAttributes(t *testing.T, data []byte, offset *int) {
	t.Helper()
	attrTag := readTestNCInt32(t, data, offset)
	if attrTag == 0 {
		return
	}
	if attrTag != ncTagAttribute {
		t.Fatalf("unexpected attr tag: %d", attrTag)
	}
	attrCount := readTestNCInt32(t, data, offset)
	for i := int32(0); i < attrCount; i++ {
		_ = readTestNCString(t, data, offset)
		attrType := readTestNCInt32(t, data, offset)
		valueCount := readTestNCInt32(t, data, offset)
		valueSize := int(valueCount)
		if attrType == ncDouble {
			valueSize *= 8
		}
		*offset += paddedTestNCSize(valueSize)
	}
}

func readTestNCString(t *testing.T, data []byte, offset *int) string {
	t.Helper()
	size := int(readTestNCInt32(t, data, offset))
	if *offset+size > len(data) {
		t.Fatalf("nc string exceeds data")
	}
	value := string(data[*offset : *offset+size])
	*offset += paddedTestNCSize(size)
	return value
}

func readTestNCInt32(t *testing.T, data []byte, offset *int) int32 {
	t.Helper()
	if *offset+4 > len(data) {
		t.Fatalf("nc int32 exceeds data")
	}
	value := int32(binary.BigEndian.Uint32(data[*offset : *offset+4]))
	*offset += 4
	return value
}

func readTestNCUint32(t *testing.T, data []byte, offset *int) uint32 {
	t.Helper()
	if *offset+4 > len(data) {
		t.Fatalf("nc uint32 exceeds data")
	}
	value := binary.BigEndian.Uint32(data[*offset : *offset+4])
	*offset += 4
	return value
}

func paddedTestNCSize(size int) int {
	return size + (4-(size%4))%4
}
