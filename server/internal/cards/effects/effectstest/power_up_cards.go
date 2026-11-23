package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// power_up_cards_test.go — #4527, the power-up proof cards (CR 702.193):
// Brave Brawler (the keyword alone, coloured reduction), Hulk, Gamma
// Goliath (a board discount for OTHER creatures' power-ups) and
// Advancing the Spirit (the {0} alternative for the first power-up of
// each of your turns). The engine half is game/power_up_test.go.

const (
	BraveBrawlerOracle       = "a1a791be-71e6-4c1f-8717-08ca62d40de2"
	HulkGammaGoliathOracle   = "3508fd7d-ca02-47f9-8b0a-525f7bde4f57"
	AdvancingTheSpiritOracle = "c9949a5a-13d4-41ca-80fe-902b0ae71204"
)

func PowerUpMain(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; g.Turn.Step != game.StepPrecombatMain; i++ {
		if i >= 12 {
			t.Fatalf("never reached the precombat main phase (stuck at %v)", g.Turn.Step)
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

// powerUpPassTo passes turns until `seat` is active, then walks to its
// precombat main phase.
func PowerUpPassTo(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; ; i++ {
		if i >= 2*len(g.Seats)+2 {
			t.Fatalf("seat %d never became active", seat)
		}
		if err := g.PassTurn(); err != nil {
			t.Fatalf("PassTurn: %v", err)
		}
		if g.Turn.ActiveSeat == seat {
			break
		}
	}
	PowerUpMain(t, g)
}

// castPowerUpCreature casts a creature from `p`'s hand with its mana
// cost floated and resolves it, so it ENTERED this turn.
func CastPowerUpCreature(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, pay, oracle string, pt int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: p.ID, Controller: p.ID, Power: pt, Toughness: pt,
		})
	})
	FloatMana(t, g, p, pay)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	ResolveTopOfStack(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatalf("%s did not resolve onto the battlefield", name)
	}
	return id
}

// pushPowerUpPermanent puts a permanent onto the battlefield without an
// entry this turn, with its printed mana cost.
func PushPowerUpPermanent(g *game.Game, owner uuid.UUID, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
	return id
}

func ChargedPowerUp(t *testing.T, g *game.Game, who, src uuid.UUID) string {
	t.Helper()
	p, err := g.PriceActivation(who, src, 0, nil)
	if err != nil {
		t.Fatalf("PriceActivation: %v", err)
	}
	return p.Total.String()
}
