package util

import (
	"fmt"
	"net/http"
)

// HxRedirect sets the HX-Redirect header and sends an empty response
// to the client, which will trigger a redirect in the browser.
func HxRedirect(rw http.ResponseWriter, path string) {
	rw.Header().Set("HX-Redirect", path)
	fmt.Fprintf(rw, "redirecting to %s", path)
}
