package testsbr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cav_f_test.go — #3172 slice cav-f: the caveats that named an optional
// draw, a mode, a clause per target or a binding to an opponent. Each
// test proves the behaviour the old caveat said was missing.

// Refute's caveat was wrong: the card says "target spell", so an
// ability on the stack is not a legal pick as printed.
func TestRefuteTargetsSpellsOnlyAsPrinted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := g.Seats[1]
	src := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	ability := effectstest.TaAnnounce(t, g, opp.ID, src, "Their Rock — do a thing", nil, nil)
	err := effectstest.CastCatalogSpellErr(t, g, "Refute", "Instant", effectstest.B17RefuteOracle, effectstest.B16TargetCard(ability))
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("an ability is not 'target spell': %v", err)
	}
	if spec, _ := effects.Lookup(effectstest.B17RefuteOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the printed clause is spells only: the card is complete")
	}
}

// Siren Stormtamer counters an ABILITY that targets you or your creature.
func TestSirenStormtamerCountersAnAbilityAimedAtYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	siren := effectstest.PushCatalogPermanent(g, me.ID, "Siren Stormtamer", "Creature — Siren Pirate Wizard", effectstest.B10SirenStormtamerOracle, false)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	src := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	fired, firedElsewhere := false, false
	aimed := effectstest.TaAnnounce(t, g, opp.ID, src, "Their Rock — aimed", effectstest.B16TargetCard(bear), &fired)
	stray := effectstest.TaAnnounce(t, g, opp.ID, src, "Their Rock — stray", nil, &firedElsewhere)

	effectstest.B10AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, siren, 0, game.ActivateAbilityParams{
		Targets: effectstest.B16TargetCard(stray),
	}); err == nil {
		t.Fatal("an ability that targets nothing of yours was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, siren, 0, game.ActivateAbilityParams{
		Targets: effectstest.B16TargetCard(aimed),
	}); err != nil {
		t.Fatalf("activate against the ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if fired {
		t.Error("the countered ability resolved")
	}
	if !firedElsewhere {
		t.Error("the other ability was not touched")
	}
	if spec, _ := effects.Lookup(effectstest.B10SirenStormtamerOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("spell or ability: nothing is declared")
	}
}

// Windgrace's Judgment: a permanent that changes hands in response is
// an illegal target for the opponent it was chosen under, even when it
// moves to ANOTHER opponent.
func TestWindgracesJudgmentPickIsBoundToItsOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a, b := g.Seats[1], g.Seats[2]
	aBear := effectstest.B12Creature(g, a.ID, "A's Bear", "Creature — Bear", 2, 2)
	bAura := effectstest.B12Permanent(g, b.ID, "B's Enchantment", "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Windgrace's Judgment", "Instant", effectstest.B30WindgracesJudgmentOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: aBear}, {Kind: game.TargetCard, ID: bAura},
	})
	// In response, seat 2 takes A's Bear for good.
	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), aBear, b.ID, game.Duration{Kind: game.Indefinite}, "test — steal") {
			t.Fatal("could not move the bear")
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aBear) {
		t.Error("the bear moved to a different opponent: it is no longer a legal target for the first opponent's clause")
	}
	if g.Battlefield.Contains(bAura) {
		t.Error("the untouched pick is destroyed")
	}
}

// Return to Dust in your own main phase takes two; the second slot is
// 'up to one other'.
func TestReturnToDustSecondTargetMustDiffer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Permanent(g, opp.ID, "Rock A", "Artifact")
	err := effectstest.CastCatalogSpellErr(t, g, "Return to Dust", "Instant", effectstest.B14ReturnToDustOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: a},
	})
	if err == nil {
		t.Error("'one other target' cannot be the same permanent")
	}
}
