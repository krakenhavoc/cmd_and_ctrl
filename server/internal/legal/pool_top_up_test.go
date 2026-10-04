package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// pool_top_up_test.go — the enumerator half of ADR 0118 §1. The
// affordability probe (canPayExcluding) asks the question the engine's
// auto-tapper answers: can the floating pool and a plan pay the cost
// TOGETHER? Before the top-up it asked the pool alone and the lands
// alone, so {G} floating plus one Forest was told a {1}{G} creature
// was not castable, by the bot's move list and by the ready ring.

// movesOfKindFrom returns the moves of one kind on one source.
func movesOfKindFrom(moves []legal.Move, kind legal.Kind, src uuid.UUID) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Kind == kind && m.Source == src {
			out = append(out, m)
		}
	}
	return out
}

// dispatchOnClone plays one move on a clone and returns the clone.
func dispatchOnClone(t *testing.T, g *game.Game, seat uuid.UUID, m legal.Move) *game.Game {
	t.Helper()
	clone := g.Clone()
	if err := actions.Dispatch(clone, actions.Action{
		Type:   actions.Type(m.Type),
		Player: m.Player,
		Caller: seat,
		Params: m.Params,
	}); err != nil {
		t.Fatalf("move %q (%s %s) rejected: %v", m.Label, m.Type, string(m.Params), err)
	}
	return clone
}

func tappedOn(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Tapped
		}
	}
	return false
}

// {G} floating plus one Forest: the {1}{G} creature is listed, and the
// move, played, taps the Forest and spends the floating {G}.
func TestEnumeratorTopsUpAFloatingGreen(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	bears := handCard(active, creature("Grizzly Bears", "{1}{G}", 2, 2))
	forest := battlefieldCard(g, active, basic("Forest", "Forest"))
	advanceTo(t, g, game.StepPrecombatMain)

	if got := movesOfKindFrom(legal.EnumerateFor(g, active.ID), legal.KindCast, bears); len(got) != 0 {
		t.Fatalf("one Forest and an empty pool cannot pay {1}{G}, but the cast is listed: %v", labels(got))
	}

	active.ManaPool.AddMana(game.ManaToken{Color: "G"})
	casts := movesOfKindFrom(legal.EnumerateFor(g, active.ID), legal.KindCast, bears)
	if len(casts) == 0 {
		t.Fatal("{G} floating plus a Forest pays {1}{G}, but no cast is listed")
	}
	after := dispatchOnClone(t, g, active.ID, casts[0])
	if !tappedOn(after, forest) {
		t.Error("the listed cast did not tap the Forest")
	}
	seat := after.PlayerByIDForEffect(active.ID)
	if len(seat.ManaPool) != 0 {
		t.Errorf("pool after the listed cast: %+v, want empty", seat.ManaPool)
	}
}

// Two hybrids and one floating {G}: the {G} pays whichever the board
// cannot, so with only an Island the {G/U}{G/W} creature is listed.
func TestEnumeratorTopsUpAcrossHybrids(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	elf := handCard(active, creature("Two Guilds", "{G/U}{G/W}", 2, 2))
	battlefieldCard(g, active, basic("Island", "Island"))
	advanceTo(t, g, game.StepPrecombatMain)
	active.ManaPool.AddMana(game.ManaToken{Color: "G"})

	moves := legal.EnumerateFor(g, active.ID)
	if got := movesOfKindFrom(moves, legal.KindCast, elf); len(got) == 0 {
		t.Fatalf("{G} floating plus an Island pays {G/U}{G/W}, but no cast is listed: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}

// The colour still has to be there: {G} floating and a Mountain do not
// pay {G}{G}, and the enumerator does not pretend they do.
func TestEnumeratorTopUpStillNeedsTheColour(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	twins := handCard(active, creature("Twin Bears", "{G}{G}", 2, 2))
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	advanceTo(t, g, game.StepPrecombatMain)
	active.ManaPool.AddMana(game.ManaToken{Color: "G"})

	if got := movesOfKindFrom(legal.EnumerateFor(g, active.ID), legal.KindCast, twins); len(got) != 0 {
		t.Errorf("{G} floating and a Mountain cannot pay {G}{G}, but the cast is listed: %v", labels(got))
	}
}

// An activation is priced through the same probe, so it tops up too.
func TestEnumeratorTopsUpAnActivation(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	pod := battlefieldCard(g, active, game.Card{
		Name:     "Green Pod",
		TypeLine: "Artifact",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:  "{1}{G}: Mark",
			Cost:   game.AbilityCost{Mana: "{1}{G}"},
			Effect: func(_ *game.Game, _ *game.StackItem) error { return nil },
		}},
	})
	battlefieldCard(g, active, basic("Forest", "Forest"))
	advanceTo(t, g, game.StepPrecombatMain)
	active.ManaPool.AddMana(game.ManaToken{Color: "G"})

	moves := legal.EnumerateFor(g, active.ID)
	if got := activationsOf(moves, pod); len(got) == 0 {
		t.Fatalf("{G} floating plus a Forest pays {1}{G}, but no activation is listed: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}
