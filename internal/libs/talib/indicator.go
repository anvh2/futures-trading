package talib

import "github.com/cinar/indicator/container/bst"

// RSIPeriod allows to calculate the RSI indicator with a non-standard period.
func RSIPeriod(period int, closing []float64) ([]float64, []float64) {
	gains := make([]float64, len(closing))
	losses := make([]float64, len(closing))

	for i := 1; i < len(closing); i++ {
		difference := closing[i] - closing[i-1]

		if difference > 0 {
			gains[i] = difference
			losses[i] = 0
		} else {
			losses[i] = -difference
			gains[i] = 0
		}
	}

	meanGains := Rma(period, gains)
	meanLosses := Rma(period, losses)

	rsi := make([]float64, len(closing))
	rs := make([]float64, len(closing))

	for i := 0; i < len(rsi); i++ {
		rs[i] = meanGains[i] / meanLosses[i]
		rsi[i] = 100 - (100 / (1 + rs[i]))
	}

	return rs, rsi
}

func KDJ(rPeriod, kPeriod, dPeriod int, high, low, closing []float64) ([]float64, []float64, []float64) {
	highest := Max(rPeriod, high)
	lowest := Min(rPeriod, low)

	rsv := multiplyBy(divide(subtract(closing, lowest), subtract(highest, lowest)), 100)

	k := Rma(kPeriod, rsv)
	d := Rma(dPeriod, k)
	j := subtract(multiplyBy(k, 3), multiplyBy(d, 2))

	return k, d, j
}

// Moving max for the given period.
func Max(period int, values []float64) []float64 {
	result := make([]float64, len(values))

	buffer := make([]float64, period)
	bst := bst.New()

	for i := 0; i < len(values); i++ {
		bst.Insert(values[i])

		if i >= period {
			bst.Remove(buffer[i%period])
		}

		buffer[i%period] = values[i]
		result[i] = bst.Max().(float64)
	}

	return result
}

// Moving min for the given period.
func Min(period int, values []float64) []float64 {
	result := make([]float64, len(values))

	buffer := make([]float64, period)
	bst := bst.New()

	for i := 0; i < len(values); i++ {
		bst.Insert(values[i])

		if i >= period {
			bst.Remove(buffer[i%period])
		}

		buffer[i%period] = values[i]
		result[i] = bst.Min().(float64)
	}

	return result
}

// Rolling Moving Average (RMA).
//
// R[0] to R[p-1] is SMA(values)
// R[p] and after is R[i] = ((R[i-1]*(p-1)) + v[i]) / p
//
// Returns r.
func Rma(period int, values []float64) []float64 {
	result := make([]float64, len(values))
	sum := float64(0)

	for i, value := range values {
		count := i + 1

		if i < period {
			sum += value
		} else {
			sum = (result[i-1] * float64(period-1)) + value
			count = period
		}

		result[i] = sum / float64(count)
	}

	return result
}

// Check values same size.
func checkSameSize(values ...[]float64) {
	if len(values) < 2 {
		return
	}

	n := len(values[0])

	for i := 1; i < len(values); i++ {
		if len(values[i]) != n {
			panic("not all same size")
		}
	}
}

// Multiply values by multipler.
func multiplyBy(values []float64, multiplier float64) []float64 {
	result := make([]float64, len(values))

	for i, value := range values {
		result[i] = value * multiplier
	}

	return result
}

// Divide values1 by values2.
func divide(values1, values2 []float64) []float64 {
	checkSameSize(values1, values2)

	result := make([]float64, len(values1))

	for i := 0; i < len(result); i++ {
		result[i] = values1[i] / values2[i]
	}

	return result
}

// subtract values2 from values1.
func subtract(values1, values2 []float64) []float64 {
	subtract := multiplyBy(values2, float64(-1))
	return add(values1, subtract)
}

// Add values1 and values2.
func add(values1, values2 []float64) []float64 {
	checkSameSize(values1, values2)

	result := make([]float64, len(values1))
	for i := 0; i < len(result); i++ {
		result[i] = values1[i] + values2[i]
	}

	return result
}
