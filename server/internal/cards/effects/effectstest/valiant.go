package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// valiant_test.go — the valiant seam (#5230), driven through Heartfire
// Hero's counter (an untargeted valiant ability) and Mouse Trapper's tap (a
// targeted one, whose trigger waits on a target prompt before it is
// announced).

const (
	ValiantHeartfireHeroOracle = "49e1e202-37bf-46ed-93d1-1adb13726039"
	ValiantMouseTrapperOracle  = "5b0a4530-6e19-41a2-9ea6-e4c8cf85b6c4"
)

type ValiantTable struct {
	G        *game.Game
	Me, Them *game.Player
	Seat     int
	Hero     uuid.UUID
}

// newValiantTable is a Heartfire Hero under the active player, in their
// precombat main phase.
func NewValiantTable(t *testing.T) ValiantTable {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	tb := ValiantTable{G: g, Seat: seat, Me: g.Seats[seat], Them: g.Seats[(seat+1)%len(g.Seats)]}
	tb.Hero = B12Push(g, tb.Me.ID, "Heartfire Hero", "Creature — Mouse Soldier", ValiantHeartfireHeroOracle, 1, 1)
	return tb
}

// valiantTarget casts an uncatalogued instant for `who` with one target
// slot per id in `targets`, and does not resolve it.
func ValiantTarget(t *testing.T, g *game.Game, who *game.Player, targets ...uuid.UUID) {
	t.Helper()
	spell := game.NewCard("Test Pump", who.ID)
	spell.TypeLine = "Instant"
	spell.ManaCost = "{W}"
	who.Hand.PushTop(spell)
	refs := make([]game.TargetRef, 0, len(targets))
	for _, id := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	if err := g.CastSpell(who.ID, spell.InstanceID, game.CastSpellParams{Targets: refs}); err != nil {
		t.Fatalf("cast the targeting spell: %v", err)
	}
}

// valiantTriggersFrom counts the abilities of `source` waiting to resolve:
// on the stack or still queued.
func ValiantTriggersFrom(g *game.Game, source uuid.UUID) int {
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == source {
			n++
		}
	}
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == source {
			n++
		}
	}
	return n
}

// --- a targeted valiant ability --------------------------------------

func NewMouseTrapperTable(t *testing.T) (ValiantTable, uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	tb := ValiantTable{G: g, Seat: seat, Me: g.Seats[seat], Them: g.Seats[(seat+1)%len(g.Seats)]}
	tb.Hero = B12Push(g, tb.Me.ID, "Mouse Trapper", "Creature — Mouse Soldier", ValiantMouseTrapperOracle, 3, 2)
	return tb, tb.Hero
}

func ValiantPickTargetPrompts(g *game.Game) int {
	n := 0
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoicePickTarget {
			n++
		}
	}
	return n
}
