package httpkit

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestToMap(t *testing.T) {
	t.Run("HeaderToMap", func(t *testing.T) {
		h := http.Header{}
		h.Set("k1", "v1")
		h.Set("k2", "v2")
		h.Add("k2", "v22")
		m := ToMapAny(h)
		diff := cmp.Diff(map[string]any{"K1": "v1", "K2": []string{"v2", "v22"}}, m)
		if diff != "" {
			t.Fatal(diff)
		}
	})
}
