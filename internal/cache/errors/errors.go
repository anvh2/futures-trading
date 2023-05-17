package errors

import "errors"

var (
	ErrorChartNotFound   = errors.New("chart: not found")
	ErrorCandlesNotFound = errors.New("candles: not found")
	ErrorSymbolNotFound  = errors.New("symbol: not found")
	ErrorFilterNotFound  = errors.New("filter: not found")
)
