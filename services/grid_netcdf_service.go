package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"nav-rain-grid-go/domains"
	"os"
	"path/filepath"
	"time"
)

const (
	ncOutputDir = "data/nc"

	ncTagDimension = int32(10)
	ncTagAttribute = int32(12)
	ncTagVariable  = int32(11)

	ncChar   = int32(2)
	ncDouble = int32(6)
)

var gridDiffNCOutputDir = ncOutputDir

type gridDiffNCFile struct {
	Path     string
	Name     string
	Size     int64
	Checksum string
}

type ncDimension struct {
	name string
	size int32
}

type ncAttribute struct {
	name   string
	values string
}

type ncVariable struct {
	name   string
	dimIDs []int32
	attrs  []ncAttribute
	values []float64
	vsize  uint32
	begin  uint32
}

func generateGridDiffNCFile(result domains.GridDiffResult) (gridDiffNCFile, error) {
	points := completeGridDiffPoints(result.Points)
	if len(points) == 0 {
		return gridDiffNCFile{}, errors.New("grid diff result has no complete points")
	}

	fileName := buildGridDiffNCFileName(result)
	filePath := filepath.Join(gridDiffNCOutputDir, fileName)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return gridDiffNCFile{}, err
	}

	data, err := buildGridDiffNCData(result, points)
	if err != nil {
		return gridDiffNCFile{}, err
	}

	tmpPath := filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return gridDiffNCFile{}, err
	}
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if err := os.Rename(tmpPath, filePath); err != nil {
		return gridDiffNCFile{}, err
	}

	checksum := sha256.Sum256(data)
	return gridDiffNCFile{
		Path:     filepath.ToSlash(filePath),
		Name:     fileName,
		Size:     int64(len(data)),
		Checksum: fmt.Sprintf("%x", checksum),
	}, nil
}

func completeGridDiffPoints(points []domains.GridDiffPointResult) []domains.GridDiffPointResult {
	result := make([]domains.GridDiffPointResult, 0, len(points))
	for _, point := range points {
		if point.Forecast1H == nil || point.Forecast12H == nil || point.Forecast24H == nil {
			continue
		}
		result = append(result, point)
	}
	return result
}

func buildGridDiffNCFileName(result domains.GridDiffResult) string {
	gridIdentifier := normalizeGridIdentifier(result.GridIdentifier, result.GridName)
	coordinateSystem := normalizeGridCoordinateSystem(result.CoordinateSystem)
	return fmt.Sprintf(
		"shouming_hourly_precipitation_forecast_%s_%s_%s.nc",
		gridIdentifier,
		coordinateSystem,
		formatGridNCBaseTime(result.BaseTime),
	)
}

func formatGridNCBaseTime(baseTime int64) string {
	return time.UnixMilli(baseTime).In(time.Local).Format("200601021504")
}

func buildGridDiffNCData(result domains.GridDiffResult, points []domains.GridDiffPointResult) ([]byte, error) {
	longitudes := make([]float64, 0, len(points))
	latitudes := make([]float64, 0, len(points))
	forecast1H := make([]float64, 0, len(points))
	forecast12H := make([]float64, 0, len(points))
	forecast24H := make([]float64, 0, len(points))

	for _, point := range points {
		longitudes = append(longitudes, point.CenterLng)
		latitudes = append(latitudes, point.CenterLat)
		forecast1H = append(forecast1H, point.Forecast1H.PredictRain)
		forecast12H = append(forecast12H, point.Forecast12H.PredictRain)
		forecast24H = append(forecast24H, point.Forecast24H.PredictRain)
	}

	globalAttrs := []ncAttribute{
		{name: "Conventions", values: "CF-1.8"},
		{name: "title", values: "hourly precipitation forecast grid points"},
		{name: "grid_name", values: result.GridName},
		{name: "grid_identifier", values: normalizeGridIdentifier(result.GridIdentifier, result.GridName)},
		{name: "coordinate_system", values: normalizeGridCoordinateSystem(result.CoordinateSystem)},
		{name: "base_time", values: formatGridNCBaseTime(result.BaseTime)},
		{name: "base_time_millis", values: fmt.Sprintf("%d", result.BaseTime)},
	}
	variables := []ncVariable{
		{
			name:   "longitude",
			dimIDs: []int32{0},
			attrs: []ncAttribute{
				{name: "long_name", values: "经度"},
				{name: "units", values: "degrees_east"},
				{name: "standard_name", values: "longitude"},
			},
			values: longitudes,
		},
		{
			name:   "latitude",
			dimIDs: []int32{0},
			attrs: []ncAttribute{
				{name: "long_name", values: "纬度"},
				{name: "units", values: "degrees_north"},
				{name: "standard_name", values: "latitude"},
			},
			values: latitudes,
		},
		{
			name:   "forecast_1h",
			dimIDs: []int32{0},
			attrs: []ncAttribute{
				{name: "long_name", values: "1h预测值"},
				{name: "units", values: "mm"},
				{name: "forecast_period", values: "1 hour"},
			},
			values: forecast1H,
		},
		{
			name:   "forecast_12h",
			dimIDs: []int32{0},
			attrs: []ncAttribute{
				{name: "long_name", values: "12h预测值"},
				{name: "units", values: "mm"},
				{name: "forecast_period", values: "12 hours"},
			},
			values: forecast12H,
		},
		{
			name:   "forecast_24h",
			dimIDs: []int32{0},
			attrs: []ncAttribute{
				{name: "long_name", values: "24h预测值"},
				{name: "units", values: "mm"},
				{name: "forecast_period", values: "24 hours"},
			},
			values: forecast24H,
		},
	}

	return encodeNetCDFClassic(
		[]ncDimension{{name: "point", size: int32(len(points))}},
		globalAttrs,
		variables,
	)
}

func encodeNetCDFClassic(dimensions []ncDimension, globalAttrs []ncAttribute, variables []ncVariable) ([]byte, error) {
	for i := range variables {
		data, err := encodeNCDoubleValues(variables[i].values)
		if err != nil {
			return nil, err
		}
		variables[i].vsize = uint32(len(data))
	}

	header := buildNetCDFClassicHeader(dimensions, globalAttrs, variables)
	begin := uint32(len(header))
	for i := range variables {
		variables[i].begin = begin
		begin += variables[i].vsize
	}

	header = buildNetCDFClassicHeader(dimensions, globalAttrs, variables)
	out := bytes.NewBuffer(make([]byte, 0, int(begin)))
	out.Write(header)
	for _, variable := range variables {
		data, err := encodeNCDoubleValues(variable.values)
		if err != nil {
			return nil, err
		}
		out.Write(data)
	}
	return out.Bytes(), nil
}

func buildNetCDFClassicHeader(dimensions []ncDimension, globalAttrs []ncAttribute, variables []ncVariable) []byte {
	var buffer bytes.Buffer
	buffer.Write([]byte{'C', 'D', 'F', 1})
	writeNCInt32(&buffer, 0)

	writeNCListTag(&buffer, ncTagDimension, len(dimensions))
	for _, dimension := range dimensions {
		writeNCString(&buffer, dimension.name)
		writeNCInt32(&buffer, dimension.size)
	}

	writeNCAttributes(&buffer, globalAttrs)

	writeNCListTag(&buffer, ncTagVariable, len(variables))
	for _, variable := range variables {
		writeNCString(&buffer, variable.name)
		writeNCInt32(&buffer, int32(len(variable.dimIDs)))
		for _, dimID := range variable.dimIDs {
			writeNCInt32(&buffer, dimID)
		}
		writeNCAttributes(&buffer, variable.attrs)
		writeNCInt32(&buffer, ncDouble)
		writeNCUint32(&buffer, variable.vsize)
		writeNCUint32(&buffer, variable.begin)
	}
	return buffer.Bytes()
}

func writeNCListTag(buffer *bytes.Buffer, tag int32, count int) {
	if count == 0 {
		writeNCInt32(buffer, 0)
		return
	}
	writeNCInt32(buffer, tag)
	writeNCInt32(buffer, int32(count))
}

func writeNCAttributes(buffer *bytes.Buffer, attrs []ncAttribute) {
	writeNCListTag(buffer, ncTagAttribute, len(attrs))
	for _, attr := range attrs {
		writeNCString(buffer, attr.name)
		writeNCInt32(buffer, ncChar)
		writeNCInt32(buffer, int32(len([]byte(attr.values))))
		buffer.WriteString(attr.values)
		writeNCPadding(buffer, len([]byte(attr.values)))
	}
}

func encodeNCDoubleValues(values []float64) ([]byte, error) {
	var buffer bytes.Buffer
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("invalid netcdf double value: %v", value)
		}
		if err := binary.Write(&buffer, binary.BigEndian, value); err != nil {
			return nil, err
		}
	}
	writeNCPadding(&buffer, buffer.Len())
	return buffer.Bytes(), nil
}

func writeNCString(buffer *bytes.Buffer, value string) {
	data := []byte(value)
	writeNCInt32(buffer, int32(len(data)))
	buffer.Write(data)
	writeNCPadding(buffer, len(data))
}

func writeNCInt32(buffer *bytes.Buffer, value int32) {
	_ = binary.Write(buffer, binary.BigEndian, value)
}

func writeNCUint32(buffer *bytes.Buffer, value uint32) {
	_ = binary.Write(buffer, binary.BigEndian, value)
}

func writeNCPadding(buffer *bytes.Buffer, size int) {
	padding := (4 - (size % 4)) % 4
	for i := 0; i < padding; i++ {
		buffer.WriteByte(0)
	}
}
