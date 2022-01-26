package pkg

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"testing"
)

func CheckRequestMethod(next http.Handler, method string) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != method {
			http.Error(writer, "not implemented", http.StatusNotImplemented)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

//FloatIsNumber check that f is not an inf or NaN
func FloatIsNumber(f float64) bool {

	if math.IsInf(f, 0) || math.IsNaN(f) {
		return false
	}
	return true
}

func Contains(list []string, value string) bool {
	for _, v := range list {
		return v == value
	}

	return false
}

func PointerFloat(f float64) *float64 {
	return &f
}

func PointerInt(i int64) *int64 {
	return &i
}

type UnsetFunc func()

func SetEnv(t *testing.T, key, value string) (UnsetFunc, error) {
	t.Helper()

	err := os.Setenv(key, value)
	if err != nil {
		return nil, err
	}

	return func() {
		_ = os.Unsetenv(key)

	}, nil

}

func ValidateCounter(value string) (int64, error) {
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return val, fmt.Errorf("value is not int")

	}

	if val < 0 {
		return val, fmt.Errorf("counter can not be negative")
	}
	return val, err
}

func StringNotEmpry(s string) bool {
	return s != ""
}
