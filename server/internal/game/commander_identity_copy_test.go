package game

import (
	"reflect"
	"testing"
)

func TestCopiedCommanderKeepsOriginalIdentity(t *testing.T) {
	for _, tc := range []struct {
		name             string
		identity, colors []string
		cost             string
		want             []string
	}{
		{"imported identity", []string{"U", "B"}, []string{"U"}, "{3}{U}", []string{"U", "B"}},
		{"printed colors fallback", nil, []string{"U"}, "{3}{U}", []string{"U"}},
		{"printed cost fallback", nil, nil, "{3}{U}", []string{"U"}},
		{"colorless original", nil, nil, "{4}", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			p := g.Seats[0]
			p.Command.Cards = nil
			cmdr := NewCommander("Original Commander", p.ID)
			cmdr.TypeLine = "Legendary Creature"
			cmdr.ColorIdentity, cmdr.Colors, cmdr.ManaCost = tc.identity, tc.colors, tc.cost
			source := NewCard("Green Creature", p.ID)
			source.TypeLine = "Creature"
			source.ColorIdentity, source.Colors = []string{"G"}, []string{"G"}
			source.ManaCost = "{1}{G}"
			cmdr.applyCopy(CopiableValuesOf(source), source)
			g.Battlefield.PushTop(cmdr)
			if got := commanderIdentityFor(g, p).Colors; !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("copied commander's identity = %v, want original %v", got, tc.want)
			}
			got := findBattlefieldCard(g, cmdr.InstanceID)
			if !got.IsCopy() || !reflect.DeepEqual(got.ColorIdentity, []string{"G"}) {
				t.Fatal("reading commander identity changed the copy's characteristics")
			}
		})
	}
}
