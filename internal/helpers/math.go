package helpers

import (
	"fmt"
	"strconv"
)

func StringToFloat(val string) float64 {
	result, _ := strconv.ParseFloat(val, 64)
	return result
}

func FloatToString(val float64) string {
	return fmt.Sprintf("%.4f", val)
}

func AddFloat(data ...string) float64 {
	result := 0.0
	for _, val := range data {
		result += StringToFloat(val)
	}

	return result
}

func DivFloatToString(fraction, numerator string) string {
	f, _ := strconv.ParseFloat(fraction, 64)
	n, _ := strconv.ParseFloat(numerator, 64)
	return fmt.Sprintf("%.4f", f/n)
}

func MulFloatToString(a, b float64) string {
	return fmt.Sprintf("%.4f", a*b)
}

func MinFloat(a, b string) float64 {
	if a < b {
		return StringToFloat(a)
	}
	return StringToFloat(b)
}
