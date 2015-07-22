package persist

import (
	"testing"

	"github.com/go-ndn/ndn"
)

func TestCache(t *testing.T) {
	c, err := New("test.db")
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []string{
		"/A/B",
		"/A",
		"/A",
		"/A/B/C",
		"/B",
	} {
		d := &ndn.Data{
			Name: ndn.NewName(test),
		}
		c.Add(d)
	}
	for _, test := range []struct {
		in   string
		want string
	}{
		{"/A", "/A/B/C"},
		{"/A/B", "/A/B/C"},
		{"/C", ""},
	} {
		d := c.Get(&ndn.Interest{
			Name: ndn.NewName(test.in),
			Selectors: ndn.Selectors{
				ChildSelector: 1,
			},
		})
		var got string
		if d != nil {
			got = d.Name.String()
		}
		if got != test.want {
			t.Fatalf("Get(%v) == %v, got %v", test.in, test.want, got)
		}
	}
}
