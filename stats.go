package flowmeter

import "math"

// MinMaxMeanStd computes min, max, mean, and sample standard deviation of values.
// For 0 or 1 value, Std is 0. For 1 value, Min and Max are that value.
func MinMaxMeanStd(values []float64) (min, max, mean, std float64) {
	n := float64(len(values))
	if n == 0 {
		return 0, 0, 0, 0
	}
	if n == 1 {
		return values[0], values[0], values[0], 0
	}
	min = values[0]
	max = values[0]
	sum := 0.0
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}
	mean = sum / n
	// Sample variance: sum((x - mean)^2) / (n-1)
	var sqDiffSum float64
	for _, v := range values {
		d := v - mean
		sqDiffSum += d * d
	}
	variance := sqDiffSum / (n - 1)
	std = math.Sqrt(variance)
	return min, max, mean, std
}

// Variance returns sample variance of values. Returns 0 for len < 2.
func Variance(values []float64) float64 {
	n := float64(len(values))
	if n < 2 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / n
	var sqDiffSum float64
	for _, v := range values {
		d := v - mean
		sqDiffSum += d * d
	}
	return sqDiffSum / (n - 1)
}

// StatsFromValues fills a Stats struct from a slice of float64.
func StatsFromValues(values []float64) Stats {
	min, max, mean, std := MinMaxMeanStd(values)
	return Stats{Min: min, Max: max, Mean: mean, Std: std}
}
