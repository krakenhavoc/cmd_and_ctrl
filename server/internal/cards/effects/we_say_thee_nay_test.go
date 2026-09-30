package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const weSayTheeNayOracle = "1abe8246-d2f9-407b-8c16-83da0a6b7de3"

// wstnCastVictim seeds a harmless sorcery in opp's hand and casts it,
// returning its stack ID — the same shape as TestB38ManaSculptCountersTheTargetSpell.
func wstnCastVictim(t *testing.T, g *game.Game, opp *game.Player) uuid.UUID {
	t.Helper()
	id := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}",
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent casts Divination: %v", err)
	}
	return id
}

func wstnCastCounter(g *game.Game, me *game.Player, victim uuid.UUID, optional []int, teamwork []uuid.UUID) error {
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "We Say Thee Nay!", TypeLine: "Instant — Arcane",
		OracleID: weSayTheeNayOracle, Owner: me.ID, Controller: me.ID,
	})
	return g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
		OptionalCosts: optional,
		TeamworkIDs:   teamwork,
	})
}

func TestWeSayTheeNayWithTeamworkTaxesFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 1)
	victim := wstnCastVictim(t, g, opp)
	a := pushVanillaCreature(g, me.ID, "Elf A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "Elf B", 1, 1)
	if err := wstnCastCounter(g, me, victim, []int{0}, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("cast We Say Thee Nay!: %v", err)
	}
	passPriorityAroundTable(t, g)
	before := countEvents(g, game.EventCounterSpell)
	ask := answerPayUnless(t, g, opp.ID, false)
	if ask.PayCost != "{4}" {
		t.Errorf("tax: %q, want {4} when cast using teamwork", ask.PayCost)
	}
	if got := countEvents(g, game.EventCounterSpell); got != before+1 {
		t.Error("declining the payment must counter the spell")
	}
}

func TestWeSayTheeNayWithoutTeamworkTaxesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Fund the {2} tax: two Islands in play for the auto-tapper.
	for i := 0; i < 2; i++ {
		g.Battlefield.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Island",
			TypeLine: "Basic Land — Island", Owner: opp.ID, Controller: opp.ID,
		})
	}
	advanceToMainOf(t, g, 1)
	victim := wstnCastVictim(t, g, opp)
	if err := wstnCastCounter(g, me, victim, nil, nil); err != nil {
		t.Fatalf("cast We Say Thee Nay!: %v", err)
	}
	passPriorityAroundTable(t, g)
	before := countEvents(g, game.EventCounterSpell)
	ask := answerPayUnless(t, g, opp.ID, true)
	if ask.PayCost != "{2}" {
		t.Errorf("tax: %q, want {2} without teamwork", ask.PayCost)
	}
	passPriorityAroundTable(t, g)
	if got := countEvents(g, game.EventCounterSpell); got != before {
		t.Error("paying the {2} must let the spell resolve, not be countered")
	}
}

func TestWeSayTheeNayRefusesTeamworkWithTooLittlePower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 1)
	victim := wstnCastVictim(t, g, opp)
	weak := pushVanillaCreature(g, me.ID, "Elf", 1, 1)
	if err := wstnCastCounter(g, me, victim, []int{0}, []uuid.UUID{weak}); !errors.Is(err, game.ErrInsufficientTeamwork) {
		t.Fatalf("err = %v, want ErrInsufficientTeamwork", err)
	}
}
