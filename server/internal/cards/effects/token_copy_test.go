package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// runStateChecksViaDraw runs the state-based actions from outside the
// game package: a sandbox draw is a special action that runs them
// (CR 117.5). Seat 1 draws, so it is never the active seat's no-op
// draw-step draw.
func runStateChecksViaDraw(t *testing.T, g *game.Game) {
	t.Helper()
	if err := g.DrawCard(g.Seats[1].ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
}

// gainAndLoseACounter puts a +1/+1 counter on `id` and takes it off
// again, the add_counter action both ways, then runs the state checks.
func gainAndLoseACounter(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	if err := g.AddCounter(id, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("add counter: %v", err)
	}
	runStateChecksViaDraw(t, g)
	if err := g.AddCounter(id, game.CounterPlusOne, -1); err != nil {
		t.Fatalf("remove counter: %v", err)
	}
	runStateChecksViaDraw(t, g)
}

// #683: a token copy of a `*` creature copies the importer's 0
// stand-in, so it must copy VariableToughness with it. Without it the
// token is a "printed 0/0" to the toughness state check, and it died
// the first time it lost its last counter, where the original stays.
// The control is a token copy of a real printed 0/0, which dies.
func TestTokenCopyOfAStarCreatureSurvivesLosingItsLastCounter(t *testing.T) {
	for _, tc := range []struct {
		name     string
		variable bool
		survives bool
	}{
		{"Star Lhurgoyf", true, true},
		{"Zero Zero Construct", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			src := uuid.New()
			me.Graveyard.PushTop(game.Card{
				InstanceID:        src,
				Name:              tc.name,
				TypeLine:          "Creature — Test",
				VariableToughness: tc.variable,
				Owner:             me.ID,
				Controller:        me.ID,
			})

			g.WithWriteLock(func() {
				_ = CreateTokenCopy{Controller: me.ID, Copy: src, N: 1}.Apply(NewContext(g, nil))
			})
			token := findBattlefieldByName(g, tc.name)
			if token == uuid.Nil {
				t.Fatal("no token copy was created")
			}
			if c, _ := battlefieldCard(g, token); c.VariableToughness != tc.variable {
				t.Errorf("token VariableToughness = %v, want the source's %v", c.VariableToughness, tc.variable)
			}
			runStateChecksViaDraw(t, g)
			if !g.Battlefield.Contains(token) {
				t.Fatal("a never-countered Toughness 0 token copy died; the placeholder skip regressed")
			}

			gainAndLoseACounter(t, g, token)
			if got := g.Battlefield.Contains(token); got != tc.survives {
				t.Errorf("after gaining and losing a counter: on battlefield = %v, want %v", got, tc.survives)
			}
		})
	}
}

// An exception that sets a numeric toughness ("except it's a 4/4")
// replaces the `*` with a number, so the token is no stand-in.
func TestTokenCopyExceptionWithANumberClearsVariableToughness(t *testing.T) {
	src := game.Card{Name: "Star Lhurgoyf", TypeLine: "Creature — Lhurgoyf", VariableToughness: true}
	hashatonZombieException(&src)
	if src.VariableToughness {
		t.Error("Hashaton's 4/4 exception kept VariableToughness")
	}
}
