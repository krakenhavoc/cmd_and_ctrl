package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bestowCard puts a bestow card in the active seat's hand and walks to
// a main phase.
func BestowCard(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id
}

// castBestowed casts the card bestowed onto host.
func CastBestowed(t *testing.T, g *game.Game, id, host uuid.UUID) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		AlternativeCost: game.BestowKey,
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: host}},
	}); err != nil {
		t.Fatalf("cast bestowed: %v", err)
	}
}

// cardState reads one card's zone flags and type predicates under the
// read lock.
type bestowState struct {
	OnStack, OnBattlefield   bool
	Creature, Aura, Bestowed bool
	AttachedTo               uuid.UUID
	Horror                   bool
}

func ReadBestowState(g *game.Game, id uuid.UUID) bestowState {
	var s bestowState
	g.ReadSnapshot(func() {
		read := func(c game.Card) {
			s.Creature = c.IsCreature()
			s.Aura = c.IsAura()
			s.Bestowed = c.Bestowed
			s.Horror = c.HasSubtype("Horror")
			if c.AttachedTo.Kind == game.TargetCard {
				s.AttachedTo = c.AttachedTo.ID
			}
		}
		for _, c := range g.Stack.Cards {
			if c.InstanceID == id {
				s.OnStack = true
				read(c)
			}
		}
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				s.OnBattlefield = true
				read(c)
			}
		}
	})
	return s
}
