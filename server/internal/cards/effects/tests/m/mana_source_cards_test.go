package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Hired Hexblade: the SOURCE, read by a permanent ----------------

// Cast off a cracked Treasure and a Swamp, the Hexblade draws a card
// and loses a life. Cast off two Swamps it is a 2/2 and nothing else.
//
// The Treasure is in a graveyard — in fact it has ceased to exist
// (CR 111.7) — before the spell is even announced, and the spell has
// finished resolving before the trigger does. Both facts are recovered
// from the snapshot, which is the whole of #1212.
func TestHiredHexbladeDrawsOnlyOffTreasureMana(t *testing.T) {
	for _, tc := range []struct {
		name     string
		treasure bool
		draws    int
	}{
		{"a Treasure paid for it", true, 1},
		{"two Swamps paid for it", false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			if tc.treasure {
				effectstest.CrackATreasureFor(t, g, me, "B")
				effectstest.FloatForTest(g, me, "B")
			} else {
				effectstest.FloatForTest(g, me, "BB")
			}
			handBefore, lifeBefore := me.Hand.Size(), me.Life

			effectstest.CastFromHandForTest(t, g, me, "Hired Hexblade", "Creature — Elf Warlock",
				"{1}{B}", effectstest.HiredHexbladeOracle, game.CastSpellParams{Strict: true})
			effectstest.PassPriorityAroundTable(t, g)

			if got := me.Hand.Size() - handBefore; got != tc.draws {
				t.Errorf("drew %d cards, want %d", got, tc.draws)
			}
			if got := lifeBefore - me.Life; got != tc.draws {
				t.Errorf("lost %d life, want %d — the draw and the loss are one clause", got, tc.draws)
			}
		})
	}
}

// CR 603.4: an intervening if gates whether the ability triggers at
// all, so a Hexblade paid for without Treasure mana never puts a
// trigger on the stack and nobody gets a window to respond to one.
func TestHiredHexbladeWithoutTreasureManaTriggersNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "BB")

	effectstest.CastFromHandForTest(t, g, me, "Hired Hexblade", "Creature — Elf Warlock",
		"{1}{B}", effectstest.HiredHexbladeOracle, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)

	for _, ev := range g.Events {
		if ev.Kind == game.EventTrigger {
			t.Errorf("a trigger was announced for a Hexblade nothing triggered: %+v", ev)
		}
	}
}

// --- Gruul Scrapper: a COLOUR, read by a permanent ------------------

// {R} paid into the generic half of {3}{G} makes the Scrapper hasty;
// paying the same four mana without a red one does not.
func TestGruulScrapperGainsHasteOnlyWhenRedWasSpent(t *testing.T) {
	for _, tc := range []struct {
		name  string
		pool  string
		haste bool
	}{
		{"a Mountain paid part of the generic half", "GRCC", true},
		{"no red mana at all", "GCCC", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			effectstest.FloatForTest(g, me, tc.pool)

			id := effectstest.CastFromHandForTest(t, g, me, "Gruul Scrapper", "Creature — Human Berserker",
				"{3}{G}", effectstest.GruulScrapperOracle, game.CastSpellParams{Strict: true})
			effectstest.PassPriorityAroundTable(t, g)

			card := effectstest.B12Card(t, g, id)
			var got bool
			g.WithWriteLock(func() { got = game.HasKeyword(&card, "haste") })
			if got != tc.haste {
				t.Errorf("haste = %v, want %v", got, tc.haste)
			}
		})
	}
}

// --- Ribbons of Night: a COLOUR, read by the spell ------------------

// The damage and the life happen either way; only the draw is the
// rider, and it reads the record off the stack item while the spell is
// still resolving.
func TestRibbonsOfNightDrawsOnlyWhenBlueWasSpent(t *testing.T) {
	for _, tc := range []struct {
		name  string
		pool  string
		draws int
	}{
		{"an Island paid part of the generic half", "BUCCC", 1},
		{"no blue mana at all", "BCCCC", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			effectstest.AdvanceToMain(t, g)
			effectstest.FloatForTest(g, me, tc.pool)

			victim := uuid.New()
			g.WithWriteLock(func() {
				g.Battlefield.PushTop(game.Card{
					InstanceID: victim, Name: "Bear", TypeLine: "Creature — Bear",
					Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
				})
			})
			handBefore, lifeBefore := me.Hand.Size(), me.Life

			effectstest.CastFromHandForTest(t, g, me, "Ribbons of Night", "Sorcery",
				"{4}{B}", effectstest.RibbonsOfNightOracl, game.CastSpellParams{
					Strict:  true,
					Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
				})
			effectstest.PassPriorityAroundTable(t, g)

			if me.Life-lifeBefore != 4 {
				t.Errorf("gained %d life, want 4 — the life is not conditional", me.Life-lifeBefore)
			}
			if g.Battlefield.Contains(victim) {
				t.Error("the Bear survived 4 damage")
			}
			if got := me.Hand.Size() - handBefore; got != tc.draws {
				t.Errorf("drew %d cards, want %d", got, tc.draws)
			}
		})
	}
}

// --- the registry ---------------------------------------------------

// Every card this PR converted is registered under the oracle ID the
// Scryfall dump gives it, and the two that declare a source wish
// declare the right one.
func TestManaSourceCardsAreRegistered(t *testing.T) {
	for _, tc := range []struct {
		oracle string
		name   string
		wish   game.ManaSourceKinds
	}{
		{effectstest.HiredHexbladeOracle, "Hired Hexblade", game.ManaSourceTreasure},
		{effectstest.GruulScrapperOracle, "Gruul Scrapper", 0},
		{effectstest.RibbonsOfNightOracl, "Ribbons of Night", 0},
	} {
		spec, ok := effects.Lookup(tc.oracle)
		if !ok {
			t.Errorf("%s is not registered under %s", tc.name, tc.oracle)
			continue
		}
		if spec.Name != tc.name {
			t.Errorf("%s registered as %q", tc.oracle, spec.Name)
		}
		if spec.WantsManaFrom != tc.wish {
			t.Errorf("%s declares WantsManaFrom %b, want %b", tc.name, spec.WantsManaFrom, tc.wish)
		}
	}
}
