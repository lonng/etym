package review

import "net/http"

func Review(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "statics/views/review.html")
}
