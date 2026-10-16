package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lightning_storm_test.go — #1865, ADR 0106 §1 amendment 2026-10-16:
// an activated ability of a spell on the stack, which any player may
// activate.

const lightningStormOracle = "2c0fceb6-d92d-4f4a-85a2-0af5be5bba81"

func castLightningStorm(t *testing.T, g *game.Game, at uuid.UUID) uuid.UUID {
	t.Helper()
	return castCatalogSpell(t, g, "Lightning Storm", "Instant", lightningStormOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: at}})
}

// passTo passes priority until `seat` holds it.
func passTo(t *testing.T, g *game.Game, seat uuid.UUID) {
	t.Helper()
	for i := 0; i < len(g.Seats); i++ {
		if ph := g.Turn.PriorityHolder; ph >= 0 && g.Seats[ph].ID == seat {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatalf("priority never reached %s", seat)
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
	g := newCatalogGame(t)
	victim := g.Seats[1].ID
	castLightningStorm(t, g, victim)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, victim); got != 37 {
		t.Errorf("life = %d, want 37", got)
	}
}

// The caster pumps their own storm: the land is discarded as the cost,
// two charge counters go on the spell, the caster may keep the target,
// and the storm deals 3 + 2.
func TestLightningStormCasterAddsChargeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1].ID
	storm := castLightningStorm(t, g, victim)
	land := handCard(me, "Test Mountain", "Basic Land — Mountain")
	if err := g.ActivateCatalogAbility(me.ID, storm, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !inGraveyard(me, land) {
		t.Fatal("the land card was not discarded to pay the cost")
	}
	passPriorityAroundTable(t, g)
	if got := stormCharge(g, storm); got != 2 {
		t.Fatalf("charge counters on the storm = %d, want 2", got)
	}
	prompt := latestRetarget(g, me.ID)
	if prompt == nil {
		t.Fatal(`"you may choose a new target" opened no prompt for the activator`)
	}
	if prompt.PickTargetMin != 0 {
		t.Errorf("the new target is optional: min %d, want 0", prompt.PickTargetMin)
	}
	if err := g.ResolveRetarget(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("keep the target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, victim); got != 35 {
		t.Errorf("life = %d, want 35 (3 plus two charge counters)", got)
	}
}

// Any player may activate it, and "you" is the activator (CR 109.5):
// the targeted player discards their own land and chooses the new
// target, which keeps the counters.
func TestLightningStormAnyPlayerRedirects(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim, other := g.Seats[0].ID, g.Seats[1], g.Seats[2].ID
	storm := castLightningStorm(t, g, victim.ID)
	passTo(t, g, victim.ID)
	land := handCard(victim, "Test Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(victim.ID, storm, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("a non-controller activates: %v", err)
	}
	if !inGraveyard(victim, land) {
		t.Fatal("the activator's land card was not discarded (CR 602.1a)")
	}
	passPriorityAroundTable(t, g)
	if latestRetarget(g, caster) != nil {
		t.Error("the spell's controller was asked; the activator chooses (CR 109.5)")
	}
	prompt := latestRetarget(g, victim.ID)
	if prompt == nil {
		t.Fatal("the activator got no retarget prompt")
	}
	if err := g.ResolveRetarget(prompt.ID, victim.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: other}}); err != nil {
		t.Fatalf("redirect: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, victim.ID); got != 40 {
		t.Errorf("the original target's life = %d, want 40", got)
	}
	if got := lifeOf(g, other); got != 35 {
		t.Errorf("the new target's life = %d, want 35", got)
	}
}

// A card in hand that is not a land does not pay "Discard a land card".
func TestLightningStormNeedsALandCard(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1].ID
	storm := castLightningStorm(t, g, victim)
	spell := handCard(me, "Test Bear", "Creature — Bear")
	err := g.ActivateCatalogAbility(me.ID, storm, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{spell}})
	if err == nil {
		t.Fatal("a creature card paid a discard-a-land-card cost")
	}
	if inGraveyard(me, spell) {
		t.Error("the refused activation discarded the card")
	}
}

// The ability functions only on the stack (CR 113.6): a Lightning Storm
// in hand offers nothing.
func TestLightningStormAbilityDoesNotWorkFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	storm := handCardForTest(me, "Lightning Storm", "Instant", lightningStormOracle)
	land := handCard(me, "Test Mountain", "Basic Land — Mountain")
	err := g.ActivateCatalogAbility(me.ID, storm, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land}})
	if !errors.Is(err, game.ErrActivationZoneNotAllowed) {
		t.Fatalf("activated from hand: err = %v, want ErrActivationZoneNotAllowed", err)
	}
}

// Countered in response, the storm is gone: the ability puts no counter
// on the card in the graveyard and asks nothing (CR 400.7, 608.2b).
func TestLightningStormAbilityDoesNothingOnceTheSpellIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1].ID
	storm := castLightningStorm(t, g, victim)
	land := handCard(me, "Test Mountain", "Basic Land — Mountain")
	if err := g.ActivateCatalogAbility(me.ID, storm, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := g.CounterSpell(storm, nil); err != nil {
		t.Fatalf("counter the storm: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := stormCharge(g, storm); got > 0 {
		t.Errorf("charge counters on the countered card = %d, want none", got)
	}
	if latestRetarget(g, me.ID) != nil {
		t.Error("a retarget prompt opened for a spell that has left the stack")
	}
	if got := lifeOf(g, victim); got != 40 {
		t.Errorf("life = %d, want 40", got)
	}
}

// A table with the ability waiting above the storm is a restore point:
// the item names its catalog row (ADR 0041 P9), the counters ride the
// spell card, and the restored game resolves both.
func TestLightningStormRestoresWithTheAbilityOnTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1].ID
	storm := castLightningStorm(t, g, victim)
	land := handCard(me, "Test Mountain", "Basic Land — Mountain")
	if err := g.ActivateCatalogAbility(me.ID, storm, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	r := restoreThroughJSON(t, g)
	passPriorityAroundTable(t, r)
	if got := stormCharge(r, storm); got != 2 {
		t.Fatalf("restored: charge counters = %d, want 2", got)
	}
	prompt := latestRetarget(r, me.ID)
	if prompt == nil {
		t.Fatal("restored: no retarget prompt")
	}
	if err := r.ResolveRetarget(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("keep the target: %v", err)
	}
	passPriorityAroundTable(t, r)
	if got := lifeOf(r, victim); got != 35 {
		t.Errorf("restored: life = %d, want 35", got)
	}
}
