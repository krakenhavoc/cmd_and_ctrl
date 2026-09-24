package game

import (
	"testing"

	"github.com/google/uuid"
)

// mana_spend_rider_test.go — the engine contract of #1547's spend
// riders, without the catalog. The printed cards are pinned end to end
// in cards/effects/mana_spend_rider_test.go.

// The filter vocabulary #1547 added to the restriction matcher: a
// colour tag, and `|` alternation within one key.
func TestRiderFilterVocabulary(t *testing.T) {
	redInstant := ManaSpendForCast(Card{TypeLine: "Instant", ManaCost: "{R}"})
	redSorcery := ManaSpendForCast(Card{TypeLine: "Sorcery", ManaCost: "{1}{R}"})
	blueInstant := ManaSpendForCast(Card{TypeLine: "Instant", ManaCost: "{U}"})
	redCreature := ManaSpendForCast(Card{TypeLine: "Creature — Goblin", ManaCost: "{R}"})
	goggles := []string{ManaRestrictCast, ManaRestrictColor("R"), ManaRestrictAnyType("Instant", "Sorcery")}

	for _, tc := range []struct {
		name string
		ctx  ManaSpendContext
		want bool
	}{
		{"a red instant", redInstant, true},
		{"a red sorcery", redSorcery, true},
		{"a blue instant", blueInstant, false},
		{"a red creature", redCreature, false},
		{"the zero context", ManaSpendContext{}, false},
	} {
		if got := tc.ctx.allows(goggles); got != tc.want {
			t.Errorf("%s: allows = %v, want %v", tc.name, got, tc.want)
		}
	}
	// An empty alternative is a card-file bug and matches nothing.
	if redInstant.allows([]string{"type:|"}) {
		t.Error(`"type:|" admitted a spell`)
	}
}

// One production is one "that mana": two tokens from one activation (a
// doubler turned one {R} into two) paying for one spell put ONE trigger
// on the queue, while two separate productions put two.
func TestRiderTriggerFiresOncePerProduction(t *testing.T) {
	const key = "test-1547-once-per-production"
	// Guarded: the registry is process-lifetime, and -count=N reruns
	// this test in the same process.
	if _, ok := ManaSpendTriggerFor(key); !ok {
		RegisterManaSpendTrigger(key, ManaSpendTrigger{
			Label:  "test rider",
			Effect: func(*Game, *StackItem) error { return nil },
		})
	}
	rider := ManaSpendRider{Kind: ManaRiderTrigger, Trigger: key, When: []string{ManaRestrictCast}}
	one, two := uuid.New(), uuid.New()
	tok := func(production uuid.UUID) ManaToken {
		r := rider
		r.Production = production
		return ManaToken{Color: "R", Riders: []ManaSpendRider{r}}
	}

	for _, tc := range []struct {
		name string
		mana []ManaToken
		want int
	}{
		{"one production, two tokens", []ManaToken{tok(one), tok(one)}, 1},
		{"two productions", []ManaToken{tok(one), tok(two)}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newFourPlayerActiveGame(t)
			item := &StackItem{ID: uuid.New(), Kind: StackItemSpell, Controller: g.Seats[0].ID,
				Paid: PaidCost{Mana: tc.mana}}
			g.WithWriteLock(func() {
				g.applyManaSpendRidersLocked(item, ManaSpendForCast(Card{TypeLine: "Instant", ManaCost: "{R}"}), Card{})
			})
			if got := len(g.PendingTriggers); got != tc.want {
				t.Errorf("queued %d triggers, want %d", got, tc.want)
			}
			for _, t2 := range item.Paid.Mana {
				if !t2.Riders[0].Applied {
					t.Error("a fired rider was not stamped Applied on the record")
				}
			}
		})
	}
}

// Undo must not alias: a token's riders are game state, and a clone
// that shared the backing array would let the restored game's record
// be rewritten by the live one (the Applied stamp is written in place).
func TestManaTokenCloneDoesNotAliasRiders(t *testing.T) {
	orig := ManaToken{Color: "G", Riders: []ManaSpendRider{{Kind: ManaRiderCantBeCountered, When: []string{ManaRestrictCast}}}}
	c := orig.clone()
	c.Riders[0].Applied = true
	c.Riders[0].When[0] = "mutated"
	if orig.Riders[0].Applied || orig.Riders[0].When[0] != ManaRestrictCast {
		t.Errorf("the clone aliases the original's riders: %+v", orig.Riders[0])
	}

	g := newFourPlayerActiveGame(t)
	g.WithWriteLock(func() { g.Seats[0].ManaPool.AddMana(orig) })
	before := g.Clone()
	g.WithWriteLock(func() { g.Seats[0].ManaPool[0].Riders[0].Applied = true })
	if before.Seats[0].ManaPool[0].Riders[0].Applied {
		t.Error("an undo snapshot shares its pool's riders with the live game")
	}
}
