package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cav_f_test.go — #3172 slice cav-f: the caveats that named an optional
// draw, a mode, a clause per target or a binding to an opponent. Each
// test proves the behaviour the old caveat said was missing.

// Refute's caveat was wrong: the card says "target spell", so an
// ability on the stack is not a legal pick as printed.
func TestRefuteTargetsSpellsOnlyAsPrinted(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := g.Seats[1]
	src := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	ability := taAnnounce(t, g, opp.ID, src, "Their Rock — do a thing", nil, nil)
	err := castCatalogSpellErr(t, g, "Refute", "Instant", b17RefuteOracle, b16TargetCard(ability))
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("an ability is not 'target spell': %v", err)
	}
	if spec, _ := Lookup(b17RefuteOracle); spec.Completeness != CompletenessFull {
		t.Error("the printed clause is spells only: the card is complete")
	}
}

// Siren Stormtamer counters an ABILITY that targets you or your creature.
func TestSirenStormtamerCountersAnAbilityAimedAtYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	siren := pushCatalogPermanent(g, me.ID, "Siren Stormtamer", "Creature — Siren Pirate Wizard", b10SirenStormtamerOracle, false)
	bear := seedCreature(g, "Bear", me.ID)
	src := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	fired, firedElsewhere := false, false
	aimed := taAnnounce(t, g, opp.ID, src, "Their Rock — aimed", b16TargetCard(bear), &fired)
	stray := taAnnounce(t, g, opp.ID, src, "Their Rock — stray", nil, &firedElsewhere)

	b10AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, siren, 0, game.ActivateAbilityParams{
		Targets: b16TargetCard(stray),
	}); err == nil {
		t.Fatal("an ability that targets nothing of yours was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, siren, 0, game.ActivateAbilityParams{
		Targets: b16TargetCard(aimed),
	}); err != nil {
		t.Fatalf("activate against the ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if fired {
		t.Error("the countered ability resolved")
	}
	if !firedElsewhere {
		t.Error("the other ability was not touched")
	}
	if spec, _ := Lookup(b10SirenStormtamerOracle); spec.Completeness != CompletenessFull {
		t.Error("spell or ability: nothing is declared")
	}
}

// Windgrace's Judgment: a permanent that changes hands in response is
// an illegal target for the opponent it was chosen under, even when it
// moves to ANOTHER opponent.
func TestWindgracesJudgmentPickIsBoundToItsOpponent(t *testing.T) {
	g := newCatalogGame(t)
	a, b := g.Seats[1], g.Seats[2]
	aBear := b12Creature(g, a.ID, "A's Bear", "Creature — Bear", 2, 2)
	bAura := b12Permanent(g, b.ID, "B's Enchantment", "Enchantment")
	castCatalogSpell(t, g, "Windgrace's Judgment", "Instant", b30WindgracesJudgmentOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: aBear}, {Kind: game.TargetCard, ID: bAura},
	})
	// In response, seat 2 takes A's Bear for good.
	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), aBear, b.ID, game.Duration{Kind: game.Indefinite}, "test — steal") {
			t.Fatal("could not move the bear")
		}
	})
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Permanent(g, opp.ID, "Rock A", "Artifact")
	err := castCatalogSpellErr(t, g, "Return to Dust", "Instant", b14ReturnToDustOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: a},
	})
	if err == nil {
		t.Error("'one other target' cannot be the same permanent")
	}
}
