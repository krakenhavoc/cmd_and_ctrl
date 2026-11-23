package testsl

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lightning_storm_test.go — #1865, ADR 0106 §1 amendment 2026-10-16:
// an activated ability of a spell on the stack, which any player may
// activate.

const lightningStormOracle = "2c0fceb6-d92d-4f4a-85a2-0af5be5bba81"

func castLightningStorm(t *testing.T, g *game.Game, at uuid.UUID) uuid.UUID {
	t.Helper()
	return effectstest.CastCatalogSpell(t, g, "Lightning Storm", "Instant", lightningStormOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: at}})
}

func stormCharge(g *game.Game, id uuid.UUID) int {
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return -1
	}
	return c.Counters[game.CounterCharge]
}

// Unpumped, it is a Lightning Bolt for 3.
func TestLightningStormDealsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[1].ID
	castLightningStorm(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, victim); got != 37 {
		t.Errorf("life = %d, want 37", got)
	}
}

// The caster pumps their own storm: the land is discarded as the cost,
// two charge counters go on the spell, the caster may keep the target,
// and the storm deals 3 + 2.
func TestLightningStormCasterAddsChargeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1].ID
	storm := castLightningStorm(t, g, victim)
	land := effectstest.HandCard(me, "Test Mountain", "Basic Land — Mountain")
	if err := g.ActivateCatalogAbility(me.ID, storm, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.InGraveyard(me, land) {
		t.Fatal("the land card was not discarded to pay the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := stormCharge(g, storm); got != 2 {
		t.Fatalf("charge counters on the storm = %d, want 2", got)
	}
	prompt := effectstest.LatestRetarget(g, me.ID)
	if prompt == nil {
		t.Fatal(`"you may choose a new target" opened no prompt for the activator`)
	}
	if prompt.PickTargetMin != 0 {
		t.Errorf("the new target is optional: min %d, want 0", prompt.PickTargetMin)
	}
	if err := g.ResolveRetarget(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("keep the target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, victim); got != 35 {
		t.Errorf("life = %d, want 35 (3 plus two charge counters)", got)
	}
}

// Any player may activate it, and "you" is the activator (CR 109.5):
// the targeted player discards their own land and chooses the new
// target, which keeps the counters.
func TestLightningStormAnyPlayerRedirects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, victim, other := g.Seats[0].ID, g.Seats[1], g.Seats[2].ID
	storm := castLightningStorm(t, g, victim.ID)
	effectstest.PassTo(t, g, victim.ID)
	land := effectstest.HandCard(victim, "Test Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(victim.ID, storm, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("a non-controller activates: %v", err)
	}
	if !effectstest.InGraveyard(victim, land) {
		t.Fatal("the activator's land card was not discarded (CR 602.1a)")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestRetarget(g, caster) != nil {
		t.Error("the spell's controller was asked; the activator chooses (CR 109.5)")
	}
	prompt := effectstest.LatestRetarget(g, victim.ID)
	if prompt == nil {
		t.Fatal("the activator got no retarget prompt")
	}
	if err := g.ResolveRetarget(prompt.ID, victim.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: other}}); err != nil {
		t.Fatalf("redirect: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, victim.ID); got != 40 {
		t.Errorf("the original target's life = %d, want 40", got)
	}
	if got := effectstest.LifeOf(g, other); got != 35 {
		t.Errorf("the new target's life = %d, want 35", got)
	}
}

// A card in hand that is not a land does not pay "Discard a land card".
func TestLightningStormNeedsALandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1].ID
	storm := castLightningStorm(t, g, victim)
	spell := effectstest.HandCard(me, "Test Bear", "Creature — Bear")
	err := g.ActivateCatalogAbility(me.ID, storm, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{spell}})
	if err == nil {
		t.Fatal("a creature card paid a discard-a-land-card cost")
	}
	if effectstest.InGraveyard(me, spell) {
		t.Error("the refused activation discarded the card")
	}
}

// The ability functions only on the stack (CR 113.6): a Lightning Storm
// in hand offers nothing.
func TestLightningStormAbilityDoesNotWorkFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	storm := effectstest.HandCardForTest(me, "Lightning Storm", "Instant", lightningStormOracle)
	land := effectstest.HandCard(me, "Test Mountain", "Basic Land — Mountain")
	err := g.ActivateCatalogAbility(me.ID, storm, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land}})
	if !errors.Is(err, game.ErrActivationZoneNotAllowed) {
		t.Fatalf("activated from hand: err = %v, want ErrActivationZoneNotAllowed", err)
	}
}

// Countered in response, the storm is gone: the ability puts no counter
// on the card in the graveyard and asks nothing (CR 400.7, 608.2b).
func TestLightningStormAbilityDoesNothingOnceTheSpellIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1].ID
	storm := castLightningStorm(t, g, victim)
	land := effectstest.HandCard(me, "Test Mountain", "Basic Land — Mountain")
	if err := g.ActivateCatalogAbility(me.ID, storm, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := g.CounterSpell(storm, nil); err != nil {
		t.Fatalf("counter the storm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := stormCharge(g, storm); got > 0 {
		t.Errorf("charge counters on the countered card = %d, want none", got)
	}
	if effectstest.LatestRetarget(g, me.ID) != nil {
		t.Error("a retarget prompt opened for a spell that has left the stack")
	}
	if got := effectstest.LifeOf(g, victim); got != 40 {
		t.Errorf("life = %d, want 40", got)
	}
}

// A table with the ability waiting above the storm is a restore point:
// the item names its catalog row (ADR 0041 P9), the counters ride the
// spell card, and the restored game resolves both.
func TestLightningStormRestoresWithTheAbilityOnTheStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1].ID
	storm := castLightningStorm(t, g, victim)
	land := effectstest.HandCard(me, "Test Mountain", "Basic Land — Mountain")
	if err := g.ActivateCatalogAbility(me.ID, storm, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	r := effectstest.RestoreThroughJSON(t, g)
	effectstest.PassPriorityAroundTable(t, r)
	if got := stormCharge(r, storm); got != 2 {
		t.Fatalf("restored: charge counters = %d, want 2", got)
	}
	prompt := effectstest.LatestRetarget(r, me.ID)
	if prompt == nil {
		t.Fatal("restored: no retarget prompt")
	}
	if err := r.ResolveRetarget(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("keep the target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, r)
	if got := effectstest.LifeOf(r, victim); got != 35 {
		t.Errorf("restored: life = %d, want 35", got)
	}
}
