package effectstest

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// game_end_cards_test.go — ADR 0057's first-wave cards (#749), each
// proved through the engine: a state-based check that does or does
// not take a player out, a game that does or does not end, and the
// outcome it records. One test per card at least; Herald of Eternal
// Dawn, the Aang deck's (#1306) card, gets the full set.

const (
	HeraldOracle      = "4080f7e3-06d3-4d3d-9929-4e826cb66713"
	PlatinumAngelOrc  = "b148578c-c0bf-4785-b97c-4b6f83028008"
	PersecutorOracle  = "7282f643-191b-41be-9f6f-82360c915d6d"
	FelidarOracle     = "5ae3cbb9-9f0c-4077-ae7c-fba660d7fb4b"
	LabManiacOracle   = "aa286dd5-aa19-446d-9003-684d81eb57ca"
	thassasOracleOrc  = "1de1b591-a73f-4974-b507-8c63e07a0868"
	AngelsGraceOracle = "66ca8a60-e028-4a5f-8177-860b888cb9d1"
)

// newTwoSeatCatalogGame is newCatalogGame at a two-seat table, where a
// single departure ends the game.
func NewTwoSeatCatalogGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

// checkState runs the state-based check the way play does: the table
// walks one step, and the step boundary checks.
func CheckState(t *testing.T, g *game.Game) {
	t.Helper()
	if g.State != game.StateActive {
		return
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
}

func EventsOfKind(g *game.Game, kind game.EventKind) []game.Event {
	var out []game.Event
	for _, ev := range g.Events {
		if ev.Kind == kind {
			out = append(out, ev)
		}
	}
	return out
}

func Destroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
}

func EmptyLibrary(p *game.Player) {
	for p.Library.Size() > 0 {
		_, _ = p.Library.PopTop()
	}
}

// --- Thassa's Oracle --------------------------------------------------

// castThassasOracle puts a blue permanent with {U}{U} out for devotion,
// casts the Oracle, and passes until its ETB asks the pick. Returns the
// Oracle's instance ID.
func CastThassasOracle(t *testing.T, g *game.Game, devotion string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Blue Devotion", ManaCost: devotion,
		TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID,
	})
	id := CastCatalogSpell(t, g, "Thassa's Oracle", "Creature — Merfolk Wizard", thassasOracleOrc, nil)
	PassPriorityAroundTable(t, g)
	return id
}
