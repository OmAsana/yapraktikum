package pkg

import (
	"math"
	"net/http"
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
