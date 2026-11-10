package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const glenElendrasAnswerOracle = "11ed8209-2184-4416-a8f0-72de00599574"

func TestGlenElendrasAnswerIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(glenElendrasAnswerOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
	if !spec.CantBeCountered {
		t.Error("the Answer can be countered")
	}
}

// glenElendraCastFor puts a catalog instant into p's hand and casts it
// (unchecked, as an opponent responding would).
func glenElendraCastFor(t *testing.T, g *game.Game, p *game.Player, name, oracle string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle, Owner: p.ID, Controller: p.ID})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("%s casts %s: %v", p.Name, name, err)
	}
	return id
}

// The Answer counters every spell and ability its opponents control and
// makes a Faerie for each one countered: an opponent's spell that can't be
// countered stays and resolves, makes no Faerie, and your own spell is
// untouched.
func TestGlenElendrasAnswerCountersOpponentsSpellsAndAbilities(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp1 := g.Seats[(seat+1)%len(g.Seats)]
	opp2 := g.Seats[(seat+2)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	nuisance := b27Push(g, opp2.ID, "Flitterwing Nuisance", "Creature — Faerie Rogue", flitterwingNuisanceOracle, "{U}", 2, 2, "U")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(nuisance, game.CounterMinusOne, 1) })
	theirBear := b12Creature(g, opp2.ID, "Their Bear", "Creature — Bear", 2, 2)
	myLife, opp1Life := me.Life, opp1.Life

	mine := glenElendraCastFor(t, g, me, "Shock", b5ShockOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp1.ID}})
	theirShock := glenElendraCastFor(t, g, opp1, "Shock", b5ShockOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if err := g.ActivateCatalogAbility(opp2.ID, nuisance, 0, game.ActivateAbilityParams{CounterKind: game.CounterMinusOne}); err != nil {
		t.Fatalf("activate the opponent's Nuisance: %v", err)
	}
	decay := glenElendraCastFor(t, g, opp1, "Abrupt Decay", abruptDecayOracle, cardRefs(bear))
	glenElendraCastFor(t, g, me, "Glen Elendra's Answer", glenElendrasAnswerOracle, nil)
	passPriorityAroundTable(t, g)

	if !opp1.Graveyard.Contains(theirShock) {
		t.Error("the opponent's Shock was not countered")
	}
	if me.Life != myLife {
		t.Errorf("the countered Shock resolved: life %d → %d", myLife, me.Life)
	}
	if opp1.Life != opp1Life-2 || !me.Graveyard.Contains(mine) {
		t.Errorf("your own Shock did not resolve: opponent life %d → %d", opp1Life, opp1.Life)
	}
	if g.Battlefield.Contains(bear) || !opp1.Graveyard.Contains(decay) {
		t.Error("the uncounterable Abrupt Decay did not resolve")
	}
	n, _ := countTokensNamed(g, me.ID, "Faerie")
	if n != 2 {
		t.Fatalf("%d Faeries, want 2 (the Shock and the ability)", n)
	}
	tok := layeredCard(t, g, findBattlefieldByName(g, "Faerie"))
	if tok.CurrentPower() != 1 || tok.CurrentToughness() != 1 || !tok.HasColor("U") || !tok.HasColor("B") || !game.HasKeyword(&tok, "flying") {
		t.Errorf("the token is not a 1/1 blue and black Faerie with flying: %+v", tok)
	}

	// The countered ability made no delayed draw for its controller.
	hand := opp2.Hand.Size()
	dealCombatDamageToPlayer(g, theirBear, me.ID, 2)
	passPriorityAroundTable(t, g)
	if opp2.Hand.Size() != hand {
		t.Errorf("the countered Nuisance ability still drew: hand %d → %d", hand, opp2.Hand.Size())
	}
}

// With nothing of an opponent's on the stack the Answer makes nothing.
func TestGlenElendrasAnswerWithNothingToCounterMakesNoFaerie(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	glenElendraCastFor(t, g, me, "Shock", b5ShockOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID}})
	glenElendraCastFor(t, g, me, "Glen Elendra's Answer", glenElendrasAnswerOracle, nil)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Faerie"); n != 0 {
		t.Errorf("%d Faeries with nothing countered", n)
	}
}

// "Countered this way" is counted: under your Guile the opponent's spell is
// exiled instead of countered and makes no Faerie, while the ability, which
// Guile does not touch, is countered and makes one.
func TestGlenElendrasAnswerCountsOnlyWhatItCountered(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp1 := g.Seats[(seat+1)%len(g.Seats)]
	opp2 := g.Seats[(seat+2)%len(g.Seats)]
	pushGuile(g, me.ID)
	nuisance := b27Push(g, opp2.ID, "Flitterwing Nuisance", "Creature — Faerie Rogue", flitterwingNuisanceOracle, "{U}", 2, 2, "U")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(nuisance, game.CounterMinusOne, 1) })

	theirShock := glenElendraCastFor(t, g, opp1, "Shock", b5ShockOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if err := g.ActivateCatalogAbility(opp2.ID, nuisance, 0, game.ActivateAbilityParams{CounterKind: game.CounterMinusOne}); err != nil {
		t.Fatalf("activate the opponent's Nuisance: %v", err)
	}
	glenElendraCastFor(t, g, me, "Glen Elendra's Answer", glenElendrasAnswerOracle, nil)
	passPriorityAroundTable(t, g)
	if resolutionCastOfferFor(g, me.ID) != nil {
		declineResolutionCast(t, g, me.ID)
	}
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(theirShock) {
		t.Fatal("Guile did not exile the Shock instead")
	}
	if n, _ := countTokensNamed(g, me.ID, "Faerie"); n != 1 {
		t.Errorf("%d Faeries, want 1 (the ability only)", n)
	}
}
