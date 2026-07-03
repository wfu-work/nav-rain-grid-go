package domains

type GridDiffResult struct {
	GridGuid         string                `json:"gridGuid"`
	GridName         string                `json:"gridName"`
	GridIdentifier   string                `json:"gridIdentifier"`
	CoordinateSystem string                `json:"coordinateSystem"`
	BaseTime         int64                 `json:"baseTime"`
	Resolution       float64               `json:"resolution"`
	Points           []GridDiffPointResult `json:"points"`
}

type GridDiffPointResult struct {
	CenterLng   float64           `json:"centerLng"`
	CenterLat   float64           `json:"centerLat"`
	Forecast1H  *GridDiffForecast `json:"forecast1h,omitempty"`
	Forecast12H *GridDiffForecast `json:"forecast12h,omitempty"`
	Forecast24H *GridDiffForecast `json:"forecast24h,omitempty"`
}

type GridDiffForecast struct {
	ForecastHour     int                    `json:"forecastHour"`
	Time             int64                  `json:"time"`
	PredictRain      float64                `json:"predictRain"`
	PredictRainLevel int                    `json:"predictRainLevel"`
	Devices          []GridDiffDeviceWeight `json:"devices"`
}

type GridDiffDeviceWeight struct {
	Sncode      string  `json:"sncode"`
	Distance    float64 `json:"distance"`
	Weight      float64 `json:"weight"`
	PredictRain float64 `json:"predictRain"`
}
