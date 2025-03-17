package pkg

import (
	"math"
	"strconv"
	"strings"
)

// Redondea a 2 decimales
func RoundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}

// Convierte un string "$13.00" a float64
func ParsePrice(priceStr string) (float64, error) {
	cleanStr := strings.ReplaceAll(priceStr, "$", "")
	cleanStr = strings.ReplaceAll(cleanStr, ",", "")
	return strconv.ParseFloat(cleanStr, 64)
}
