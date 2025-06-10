package helpers

import "math"

type Step struct {
	Minimum int64
	Maximum int64
}

func GetStepSize(max int64, stepLimit int8) int64 {
	recordCount := float64(max) / float64(stepLimit)
	recordCount = math.Ceil(recordCount)

	return int64(recordCount)
}

func GetSteps(min, max, stepSize int64) []Step {
	var steps []Step
	var i int64
	for i = min; i <= max; i += stepSize {
		steps = append(steps, Step{
			Minimum: i,
			Maximum: i + stepSize,
		})
	}

	return steps
}

func GetStepChunks(slices []Step, chunk int) [][]Step {
	var divided [][]Step

	chunkSize := (len(slices) + chunk - 1) / chunk

	for i := 0; i < len(slices); i += chunkSize {
		end := i + chunkSize

		if end > len(slices) {
			end = len(slices)
		}

		divided = append(divided, slices[i:end])
	}

	return divided
}
