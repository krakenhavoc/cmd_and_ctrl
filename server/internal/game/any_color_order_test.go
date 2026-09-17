package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// Owner decision (2026-09-17): "any color" offers all five colours with
// the commander's identity listed first; only NarrowToCommanderIdentity
// (Command Tower, Arcane Signet, ...) intersects.

func TestManaPickOptions(t *testing.T) {
	five := []string{"W", "U", "B", "R", "G"}
	cases := []struct {
		name     string
		options  []string
		identity []string
		narrow   bool
		want     []string
	}{
		{"mono-green, any color", five, []string{"G"}, false, []string{"G", "W", "U", "B", "R"}},
		{"golgari, any color", five, []string{"B", "G"}, false, []string{"B", "G", "W", "U", "R"}},
		{"five-colour identity keeps printed order", five, five, false, five},
		{"no identity keeps printed order", five, nil, false, five},
		{"no overlap keeps printed order", []string{"W", "U"}, []string{"G"}, false, []string{"W", "U"}},
		{"dual, second colour in identity", []string{"W", "B"}, []string{"B"}, false, []string{"B", "W"}},
		{"narrow: mono-green tower", five, []string{"G"}, true, []string{"G"}},
		{"narrow: no identity keeps all", five, nil, true, five},
		{"narrow: no overlap falls back", []string{"W", "U"}, []string{"G"}, true, []string{"W", "U"}},
		{"single option untouched", []string{"C"}, []string{"G"}, true, []string{"C"}},
	}
	for _, tc := range cases {
		if got := manaPickOptions(tc.options, tc.identity, tc.narrow); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
	// The input slice is never reordered in place: it may be the
	// catalog's parsed option list.
	in := []string{"W", "U", "B", "R", "G"}
	_ = manaPickOptions(in, []string{"G"}, false)
	if !reflect.DeepEqual(in, five) {
		t.Errorf("manaPickOptions reordered its input: %v", in)
	}
}

// The auto-tapper books Birds of Paradise at its printed width (so an
// off-identity pip is payable, as the card prints) but mints an
// identity colour when the slot only pays generic mana.
func TestAutoTapAnyColorPrefersCommanderIdentity(t *testing.T) {
	withCatalogHook(t, birdsHook)

	t.Run("generic slot mints the identity colour", func(t *testing.T) {
		g := newActiveGame(t)
		advanceTo(t, g, StepPrecombatMain)
		p := g.Seats[0]
		setCommanderCostForTest(t, p, "{2}{G}")
		birds := pushBattlefieldForTest(g, p.ID, "Birds of Paradise", "Creature — Bird", "d3a0b660-358c-41bd-9cd2-41fbf3491b1a")
		g.WithWriteLock(func() {
			g.materializePlanLocked(p, []uuid.UUID{birds}, costFor(t, "{1}"))
		})
		if len(p.ManaPool) != 1 || p.ManaPool[0].Color != "G" {
			t.Errorf("pool = %+v, want one {G}", p.ManaPool)
		}
	})

	t.Run("an off-identity pip is payable", func(t *testing.T) {
		g := newActiveGame(t)
		advanceTo(t, g, StepPrecombatMain)
		p := g.Seats[0]
		setCommanderCostForTest(t, p, "{2}{G}")
		pushBattlefieldForTest(g, p.ID, "Birds of Paradise", "Creature — Bird", "d3a0b660-358c-41bd-9cd2-41fbf3491b1a")
		id := pushTypedCardToHandWithCost(p, "Swords to Plowshares", "Instant", "{W}")
		if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
			t.Fatalf("auto-tap {W} off Birds under a mono-green commander: %v", err)
		}
		if len(p.ManaPool) != 0 {
			t.Errorf("post-spend pool = %+v, want empty", p.ManaPool)
		}
	})
}
