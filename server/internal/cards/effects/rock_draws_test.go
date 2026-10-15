package effects

import "testing"

// rock_draws_test.go — #2939 gap 5: a rock's sacrifice-to-draw row
// declares the cards it draws, so the bot prices it as a draw (ADR 0126
// §6) and not only as a flat activation.
func TestRockDrawRowsDeclareDraws(t *testing.T) {
	for _, c := range []struct {
		name, oracle string
		draws        int
	}{
		{"Mind Stone", "c97361b5-af16-4a7b-af85-a429dbaf4ad2", 1},
		{"Hedron Archive", "32263baa-d3f0-463f-92b3-4e9938476add", 2},
		{"Commander's Sphere", "0b67c4e2-f88b-4e01-85a1-9d5f5b8db13b", 1},
		{"Dreamstone Hedron", "0e2575be-c596-4c8c-bf07-7941ca065721", 3},
	} {
		spec, ok := Lookup(c.oracle)
		if !ok || len(spec.Activated) != 1 {
			t.Fatalf("%s: spec found=%v with %d activated rows, want one", c.name, ok, len(spec.Activated))
		}
		if got := spec.Activated[0].Purpose.Draws; got != c.draws {
			t.Errorf("%s: draw row declares Draws %d, want %d", c.name, got, c.draws)
		}
	}
}
