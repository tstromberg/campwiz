package parse

import "testing"

func TestHTMLText(t *testing.T) {
	var tests = []struct {
		in  string
		out string
	}{
		{`&#8220;Lions.&#8221;`, `“Lions.”`},
		{`winter <a id="page_499"></a>weekends.`, "winter weekends."},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := htmlText(tt.in)
			if got != tt.out {
				t.Errorf("got %q, want %q", got, tt.out)
			}
		})
	}
}
