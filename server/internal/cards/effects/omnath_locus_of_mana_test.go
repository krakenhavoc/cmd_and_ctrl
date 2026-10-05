package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #2363: Omnath's +1/+1 reads the controller's pool, which is not a zone
// move or a counter, so the layer cache has to drop on a pool change.

const omnathManaOracle = "9da896e1-2256-425b-b801-1ae6f0470559"

func omnathSetup(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Omnath, Locus of Mana",
		OracleID: omnathManaOracle, TypeLine: "Legendary Creature — Elemental",
		ManaCost: "{2}{G}", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	return g, me, id
}

func omnathWantPT(t *testing.T, g *game.Game, _ *game.Player, id uuid.UUID, want int, when string) {
	t.Helper()
	if p, tt := effectivePower(t, g, id), effectiveToughness(t, g, id); p != want || tt != want {
		t.Errorf("%s: effective %d/%d, want %d/%d", when, p, tt, want, want)
	}
	// The wire view must agree, not just the engine.
	view := protocol.ViewOfGame(g)
	for _, c := range view.Battlefield.Cards {
		if c.InstanceID == id.String() {
			if c.Power != want || c.Toughness != want {
				t.Errorf("%s: view shows %d/%d, want %d/%d", when, c.Power, c.Toughness, want, want)
			}
			return
		}
	}
}

func TestOmnathGrowsWithGreenManaAndShrinksWhenItIsSpent(t *testing.T) {
	g, me, id := omnathSetup(t)
	omnathWantPT(t, g, me, id, 1, "empty pool")

	mkAdd(t, g, me, "{G}{G}{R}", game.AddManaOptions{})
	omnathWantPT(t, g, me, id, 3, "two green and a red in the pool")

	// A real payment: strict-mode CastSpell takes {G} out of the pool.
	spell := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: spell, Name: "Plain Sorcery", TypeLine: "Sorcery", ManaCost: "{G}",
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	omnathWantPT(t, g, me, id, 2, "after spending one green")
}

func TestOmnathKeepsItsSizeAcrossAStepBoundary(t *testing.T) {
	g, me, id := omnathSetup(t)
	mkAdd(t, g, me, "{G}{G}{G}{R}", game.AddManaOptions{})
	omnathWantPT(t, g, me, id, 4, "three green")
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{"G": 3})
	omnathWantPT(t, g, me, id, 4, "after the step ended (green kept)")
}

func TestOmnathCountsOnlyItsControllersGreen(t *testing.T) {
	g, me, id := omnathSetup(t)
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mkAdd(t, g, other, "{G}{G}", game.AddManaOptions{})
	omnathWantPT(t, g, me, id, 1, "an opponent's green is not mine")
	mkAdd(t, g, me, "{G}", game.AddManaOptions{})
	omnathWantPT(t, g, me, id, 2, "my green counts")
}

func TestOmnathSizeSurvivesASnapshotRoundTrip(t *testing.T) {
	g, me, id := omnathSetup(t)
	mkAdd(t, g, me, "{G}{G}", game.AddManaOptions{})
	omnathWantPT(t, g, me, id, 3, "before the round trip")

	restored := restoreRoundTrip(t, g, false)
	rme := restored.Seats[g.Turn.ActiveSeat]
	omnathWantPT(t, restored, rme, id, 3, "after the round trip")

	// And the restored game keeps tracking.
	mkAdd(t, restored, rme, "{G}", game.AddManaOptions{})
	omnathWantPT(t, restored, rme, id, 4, "restored game, one more green")
}
