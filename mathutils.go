package main

import (
	"fmt"
)

func min(f1 float64, f2 float64) float64{
    if (f1 < f2) {
        return f1
    }
    return f2
}

func max(f1 float64, f2 float64) float64{
    if (f1 > f2) {
        return f1
    }
    return f2
}

func float64ToString(f float64) string {
	return fmt.Sprintf("%f", f)
}