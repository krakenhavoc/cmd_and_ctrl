package effects

import (
	"errors"
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
	braveBrawlerOracle       = "a1a791be-71e6-4c1f-8717-08ca62d40de2"
	hulkGammaGoliathOracle   = "3508fd7d-ca02-47f9-8b0a-525f7bde4f57"
	advancingTheSpiritOracle = "c9949a5a-13d4-41ca-80fe-902b0ae71204"
)

func powerUpMain(t *testing.T, g *game.Game) {
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
func powerUpPassTo(t *testing.T, g *game.Game, seat int) {
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
	powerUpMain(t, g)
}

// castPowerUpCreature casts a creature from `p`'s hand with its mana
// cost floated and resolves it, so it ENTERED this turn.
func castPowerUpCreature(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, pay, oracle string, pt int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: p.ID, Controller: p.ID, Power: pt, Toughness: pt,
		})
	})
	floatMana(t, g, p, pay)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	resolveTopOfStack(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatalf("%s did not resolve onto the battlefield", name)
	}
	return id
}

// pushPowerUpPermanent puts a permanent onto the battlefield without an
// entry this turn, with its printed mana cost.
func pushPowerUpPermanent(g *game.Game, owner uuid.UUID, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
	return id
}

func chargedPowerUp(t *testing.T, g *game.Game, who, src uuid.UUID) string {
	t.Helper()
	p, err := g.PriceActivation(who, src, 0, nil)
	if err != nil {
		t.Fatalf("PriceActivation: %v", err)
	}
	return p.Total.String()
}

func TestPowerUpCardsAreFull(t *testing.T) {
	for _, oracle := range []string{braveBrawlerOracle, hulkGammaGoliathOracle, advancingTheSpiritOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", oracle)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v, want Full and none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// Brave Brawler cast this turn: {4}{W} less its {1}{W} is {3}, paid with
// three colourless mana; it puts two +1/+1 counters on it and cannot be
// activated again.
func TestPowerUpBraveBrawlerCostsLessTheTurnItEnteredAndOnlyOnce(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	brawler := castPowerUpCreature(t, g, me, "Brave Brawler", "Creature — Human Warrior Hero", "{1}{W}", "{C}{W}", braveBrawlerOracle, 2)

	if got := chargedPowerUp(t, g, me.ID, brawler); got != "{3}" {
		t.Fatalf("the turn it entered: %q, want {3}", got)
	}
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, brawler, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up for {3}: %v", err)
	}
	resolveTopOfStack(t, g)
	if got := plusOnes(g, brawler); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, brawler, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second activation: err = %v, want ErrPowerUpUsed", err)
	}
}

// The turn after, the same Brawler pays the printed {4}{W}.
func TestPowerUpBraveBrawlerPaysFullPriceTheNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	brawler := castPowerUpCreature(t, g, me, "Brave Brawler", "Creature — Human Warrior Hero", "{1}{W}", "{C}{W}", braveBrawlerOracle, 2)
	powerUpPassTo(t, g, 0)
	if got := chargedPowerUp(t, g, me.ID, brawler); got != "{4}{W}" {
		t.Errorf("the next turn: %q, want {4}{W}", got)
	}
}

// Hulk's static discounts another creature's power-up by {3} and not his
// own, and only for creatures his controller controls. His own costs
// {3} the turn he entered ({6}{R}{G} less {3}{R}{G}).
func TestPowerUpHulkDiscountsOtherCreaturesPowerUps(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushPowerUpPermanent(g, me.ID, "Brave Brawler", "Creature — Human Warrior Hero", "{1}{W}", braveBrawlerOracle)
	theirs := pushPowerUpPermanent(g, opp.ID, "Brave Brawler", "Creature — Human Warrior Hero", "{1}{W}", braveBrawlerOracle)
	if got := chargedPowerUp(t, g, me.ID, mine); got != "{4}{W}" {
		t.Fatalf("before Hulk: %q, want {4}{W}", got)
	}
	hulk := castPowerUpCreature(t, g, me, "Hulk, Gamma Goliath", "Legendary Creature — Gamma Berserker Hero", "{3}{R}{G}", "{C}{C}{C}{R}{G}", hulkGammaGoliathOracle, 6)

	if got := chargedPowerUp(t, g, me.ID, mine); got != "{1}{W}" {
		t.Errorf("my Brawler under Hulk: %q, want {1}{W}", got)
	}
	if got := chargedPowerUp(t, g, opp.ID, theirs); got != "{4}{W}" {
		t.Errorf("an opponent's Brawler: %q, want {4}{W} — Hulk discounts creatures you control", got)
	}
	if got := chargedPowerUp(t, g, me.ID, hulk); got != "{3}" {
		t.Errorf("Hulk's own power-up the turn he entered: %q, want {3} — his static is for OTHER creatures", got)
	}
	powerUpPassTo(t, g, 0)
	if got := chargedPowerUp(t, g, me.ID, hulk); got != "{6}{R}{G}" {
		t.Errorf("Hulk's own power-up the next turn: %q, want {6}{R}{G}", got)
	}
	floatMana(t, g, me, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, mine, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate my Brawler for {1}{W}: %v", err)
	}
	resolveTopOfStack(t, g)
	if got := plusOnes(g, mine); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
}

// Advancing the Spirit: the first power-up I activate during my turn is
// priced from {0}; the second that turn pays its cost; on an opponent's
// turn there is no alternative; and my next turn opens it again.
func TestPowerUpAdvancingTheSpiritPaysZeroForTheFirstPowerUpOfEachOfYourTurns(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	pushPowerUpPermanent(g, me.ID, "Advancing the Spirit", "Enchantment", "{2}{G}", advancingTheSpiritOracle)
	a := pushPowerUpPermanent(g, me.ID, "Brave Brawler", "Creature — Human Warrior Hero", "{1}{W}", braveBrawlerOracle)
	b := pushPowerUpPermanent(g, me.ID, "Brave Brawler", "Creature — Human Warrior Hero", "{1}{W}", braveBrawlerOracle)
	c := pushPowerUpPermanent(g, me.ID, "Brave Brawler", "Creature — Human Warrior Hero", "{1}{W}", braveBrawlerOracle)

	if got := chargedPowerUp(t, g, me.ID, a); got != "" {
		t.Fatalf("the first power-up of my turn: %q, want {0}", got)
	}
	if me.ManaPool.Total() != 0 {
		t.Fatalf("setup: pool = %v, want empty", me.ManaPool)
	}
	if err := g.ActivateCatalogAbility(me.ID, a, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up for {0} with an empty pool: %v", err)
	}
	// Announced, not resolved: the alternative is already spent.
	if got := chargedPowerUp(t, g, me.ID, b); got != "{4}{W}" {
		t.Errorf("the second power-up of my turn: %q, want {4}{W}", got)
	}
	resolveTopOfStack(t, g)
	if got := plusOnes(g, a); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}

	powerUpPassTo(t, g, 1)
	if got := chargedPowerUp(t, g, me.ID, c); got != "{4}{W}" {
		t.Errorf("on an opponent's turn: %q, want {4}{W} — the clause is for each of YOUR turns", got)
	}
	powerUpPassTo(t, g, 0)
	if got := chargedPowerUp(t, g, me.ID, c); got != "" {
		t.Errorf("the first power-up of my next turn: %q, want {0}", got)
	}
}
