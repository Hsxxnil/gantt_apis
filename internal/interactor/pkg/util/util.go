package util

import (
	"github.com/bytedance/sonic"
	"hta/internal/interactor/pkg/util/log"
	"math"
	"math/rand"
	"time"
)

func PointerString(s string) *string     { return &s }
func PointerInt64(i int64) *int64        { return &i }
func PointerBool(b bool) *bool           { return &b }
func PointerTime(t time.Time) *time.Time { return &t }

func GenerateRangeNum(min, max int) int {
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(max-min) + min
	return randNum
}

func Round(x float64) int64 {
	return int64(math.Floor(x + 0.5))
}

// RemoveString is a generic function to remove a string from a slice of strings.
func RemoveString(s []string, item string) []string {
	for i, v := range s {
		if v == item {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// DecodeJSONToSlice is a generic function to decode a JSON string into a slice.
func DecodeJSONToSlice(jsonStr string, targetSlice any) error {
	if jsonStr != "" {
		err := sonic.Unmarshal([]byte(jsonStr), targetSlice)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}
