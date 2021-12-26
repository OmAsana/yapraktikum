package pkg

import "net/http"

func CheckRequestMethod(next http.Handler, method string) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != method {
			http.Error(writer, "not implemented", http.StatusNotImplemented)
			return
		}

		next.ServeHTTP(writer, request)
	})
}
