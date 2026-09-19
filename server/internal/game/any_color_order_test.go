package game

import (
	"reflect"
	"testing"
)

// Owner decision (2026-09-17): "any color" offers all five colours with
// the commander's identity listed first; only NarrowToCommanderIdentity
// (Command Tower, Arcane Signet, ...) intersects.
//
// CR 903.4f (#844): that intersection may come back EMPTY, and an empty
// option list is "this ability adds no mana". The three identities are
// told apart by identityState, never by len(Colors).

// known / colourless / noCommander / unknown are the four identities
// this table needs, spelled out so each case says which one it means.
func known(colors ...string) commanderIdentity {
	return commanderIdentity{State: identityKnown, Colors: colors}
}

func TestManaPickOptions(t *testing.T) {
	five := []string{"W", "U", "B", "R", "G"}
	colourless := commanderIdentity{State: identityKnown} // Kozilek, Karn
	noCommander := commanderIdentity{State: identityNoCommander}
	unknown := commanderIdentity{State: identityUnknown} // placeholder, no card data
	cases := []struct {
		name     string
		options  []string
		identity commanderIdentity
		narrow   bool
		want     []string
	}{
		{"mono-green, any color", five, known("G"), false, []string{"G", "W", "U", "B", "R"}},
		{"golgari, any color", five, known("B", "G"), false, []string{"B", "G", "W", "U", "R"}},
		{"five-colour identity keeps printed order", five, known(five...), false, five},
		{"no commander keeps printed order", five, noCommander, false, five},
		{"colourless commander keeps printed order", five, colourless, false, five},
		{"missing colour data keeps printed order", five, unknown, false, five},
		{"no overlap keeps printed order", []string{"W", "U"}, known("G"), false, []string{"W", "U"}},
		{"dual, second colour in identity", []string{"W", "B"}, known("B"), false, []string{"B", "W"}},
		{"narrow: mono-green tower", five, known("G"), true, []string{"G"}},
		// CR 903.4f: "that quality is undefined if that player doesn't
		// have a commander. That part of the ability won't do
		// anything." A colourless commander (Kozilek) has an identity,
		// and it names no colours, so there is equally nothing to add.
		{"narrow: no commander adds nothing", five, noCommander, true, nil},
		{"narrow: colourless commander adds nothing", five, colourless, true, nil},
		// The one fallback left: a data gap is not a rules state, so a
		// commander with no colour data anywhere leaves the printed
		// card alone rather than switching it off.
		{"narrow: missing colour data keeps the printed set", five, unknown, true, five},
		// The old no-overlap fallback is gone: "any color in your
		// commander's color identity" can only add a colour the
		// identity names. No catalog card reaches either case — the
		// four narrowing cards all print five colours.
		{"narrow: no overlap adds nothing", []string{"W", "U"}, known("G"), true, nil},
		{"narrow: single off-identity option adds nothing", []string{"C"}, known("G"), true, nil},
		{"narrow: single in-identity option stands", []string{"G"}, known("G"), true, []string{"G"}},
	}
	for _, tc := range cases {
		if got := manaPickOptions(tc.options, tc.identity, tc.narrow); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
	// The input slice is never reordered in place: it may be the
	// catalog's parsed option list.
	in := []string{"W", "U", "B", "R", "G"}
	_ = manaPickOptions(in, known("G"), false)
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
			g.materializePlanLocked(p, tapPlan{{CardID: birds}}, costFor(t, "{1}"))
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
