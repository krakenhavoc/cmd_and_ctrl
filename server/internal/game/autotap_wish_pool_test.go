package game

import (
	"testing"

	"github.com/google/uuid"
)

// The generic half walks colour buckets (colourless, then W U B R G), so
// rider-marked mana has to be ranked ACROSS colours, not only inside one:
// a haste {G} is spent on a creature spell before a plain {W} the default
// order reaches first, and a haste {W} is kept from an instant in favour of
// a plain {G} it would otherwise reach first (#3853).
func TestThePoolRanksRiderManaAcrossColoursOnTheGenericHalf(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, riderColor, plainColor string
		wantRider                              bool
	}{
		{"a creature spends the haste {G} before a plain {W}", "Creature — Bear", "G", "W", true},
		{"an instant keeps the haste {W} and spends a plain {G}", "Instant", "W", "G", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			rider := wishToken(tc.riderColor, hasteForCreatures())
			plain := ManaToken{Color: tc.plainColor, Source: uuid.New()}
			me.ManaPool.AddMana(plain, rider)
			c := NewCard("Spell", me.ID)
			c.TypeLine, c.ManaCost, c.Controller = tc.typeLine, "{1}", me.ID
			me.Hand.PushTop(c)
			if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			paid := g.StackMeta[c.InstanceID].Paid.Mana
			if len(paid) != 1 || (paid[0].Source == rider.Source) != tc.wantRider {
				t.Errorf("spent %+v, want the rider mana = %v", paid, tc.wantRider)
			}
		})
	}
}
