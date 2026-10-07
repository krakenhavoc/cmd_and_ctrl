package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// return_this_cost_test.go — #2028: the catalog cards whose activation
// cost returns the permanent itself to its owner's hand (ReturnThis).
// The engine half is game/return_self_cost_test.go.

const (
	rtcGossamerChains = "af3936ab-0156-438d-9fef-a924d2fcffc3"
	rtcShigeki        = "cbad3570-5417-42e6-b3ef-f42194098314"
	rtcBrokenFall     = "79791c7f-dd33-406a-8c2e-6b722abd5161"
	rtcMoltingSkin    = "8262dbb3-7abb-4de8-9a1d-edee71906a10"
)

// Gossamer Chains: the enchantment is in its owner's hand as soon as the
// ability is activated, and the unblocked attacker's combat damage is
// prevented.
func TestGossamerChainsReturnsItselfAndStopsTheUnblockedAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	att := pr7Creature(g, me.ID, "Attacker", 3, "R")
	chains := pushCatalogPermanent(g, opp.ID, "Gossamer Chains", "Enchantment", rtcGossamerChains, false)
	pr7bAttack(t, g, opp.ID, att)
	pr7bBlock(t, g, nil)

	if err := g.ActivateCatalogAbility(opp.ID, chains, 0, game.ActivateAbilityParams{Targets: pr7bTargets(att)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !opp.Hand.Contains(chains) || g.Battlefield.Contains(chains) {
		t.Fatal("Gossamer Chains is not in its owner's hand once the ability is activated")
	}
	passPriorityAroundTable(t, g)
	life := opp.Life
	pr7bDamageStep(t, g)
	if opp.Life != life {
		t.Fatalf("defender life %d, want %d: the unblocked attacker's combat damage is prevented", opp.Life, life)
	}
}

// Before blockers are declared nothing is unblocked, so the target is
// illegal and nothing is paid: the enchantment stays where it is.
func TestGossamerChainsNeedsAnUnblockedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	att := pr7Creature(g, me.ID, "Attacker", 3, "R")
	chains := pushCatalogPermanent(g, me.ID, "Gossamer Chains", "Enchantment", rtcGossamerChains, false)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(att, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}

	err := g.ActivateCatalogAbility(me.ID, chains, 0, game.ActivateAbilityParams{Targets: pr7bTargets(att)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("activate before blockers: err = %v, want ErrIllegalTarget", err)
	}
	if !g.Battlefield.Contains(chains) {
		t.Fatal("a refused activation returned the enchantment")
	}
}

// Shigeki: the creature goes back to its owner's hand as the cost, the
// land the controller picks enters tapped, and the rest go to the
// graveyard.
func TestShigekiReturnsItselfAndPutsALandOntoTheBattlefieldTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	shigeki := apaPush(g, me.ID, me.ID, game.Card{Name: "Shigeki, Jukai Visionary",
		TypeLine: "Legendary Enchantment Creature — Snake Druid", OracleID: rtcShigeki, Power: 1, Toughness: 3})
	ids := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Ponder", TypeLine: "Sorcery"},
	)
	forest := ids[0]

	if err := g.ActivateCatalogAbility(me.ID, shigeki, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Hand.Contains(shigeki) || g.Battlefield.Contains(shigeki) {
		t.Fatal("Shigeki is not in its owner's hand once the ability is activated")
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, forest)
	land := findBattlefieldCardForTest(g, forest)
	if land == nil || !land.Tapped {
		t.Fatalf("the Forest is %+v, want it on the battlefield tapped", land)
	}
	for _, id := range ids[1:] {
		if !me.Graveyard.Contains(id) {
			t.Errorf("revealed card %v is not in the graveyard", id)
		}
	}
}

// A Shigeki gained from an opponent goes back to that opponent's hand,
// and the dig is still the activator's.
func TestAStolenShigekiReturnsToItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	shigeki := apaPush(g, opp.ID, me.ID, game.Card{Name: "Shigeki, Jukai Visionary",
		TypeLine: "Legendary Enchantment Creature — Snake Druid", OracleID: rtcShigeki, Power: 1, Toughness: 3})
	seedSearchLibrary(me, game.Card{Name: "Shock", TypeLine: "Instant"})

	if err := g.ActivateCatalogAbility(me.ID, shigeki, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !opp.Hand.Contains(shigeki) {
		t.Fatal("the stolen Shigeki is not in its owner's hand")
	}
	if me.Hand.Contains(shigeki) {
		t.Fatal("the stolen Shigeki went to the activator's hand")
	}
}

// Shigeki's return ability works only on the battlefield; from the hand
// the card offers its channel ability and nothing else.
func TestShigekisReturnAbilityIsNotActivatedFromTheHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	shigeki := handCardForTest(me, "Shigeki, Jukai Visionary", "Legendary Enchantment Creature — Snake Druid", rtcShigeki)

	err := g.ActivateCatalogAbility(me.ID, shigeki, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrActivationZoneNotAllowed) {
		t.Fatalf("activate from hand: err = %v, want ErrActivationZoneNotAllowed", err)
	}
	if !me.Hand.Contains(shigeki) {
		t.Fatal("the refused activation moved the card")
	}
}

// Shigeki's channel: X target nonlegendary cards from your graveyard
// return to your hand.
func TestShigekiChannelReturnsXNonlegendaryCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	shigeki := handCardForTest(me, "Shigeki, Jukai Visionary", "Legendary Enchantment Creature — Snake Druid", rtcShigeki)
	a := rtcGraveyardCard(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	b := rtcGraveyardCard(me, game.Card{Name: "Shock", TypeLine: "Instant"})
	legend := rtcGraveyardCard(me, game.Card{Name: "Isamaru", TypeLine: "Legendary Creature — Dog"})

	if err := g.ActivateCatalogAbility(me.ID, shigeki, 1, game.ActivateAbilityParams{
		XValue: 1, Targets: pr7bTargets(legend)}); err == nil {
		t.Fatal("channel targeted a legendary card")
	}
	if err := g.ActivateCatalogAbility(me.ID, shigeki, 1, game.ActivateAbilityParams{
		XValue: 2, Targets: pr7bTargets(a, b)}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Fatal("the two targeted cards are not in hand")
	}
	if !me.Graveyard.Contains(shigeki) || !me.Graveyard.Contains(legend) {
		t.Fatal("want Shigeki discarded and the legendary card left in the graveyard")
	}
}

// Broken Fall and Molting Skin: the enchantment goes back to hand and
// the target gets a regeneration shield.
func TestReturnThisRegenerateRows(t *testing.T) {
	for _, c := range []struct{ name, oracle string }{
		{"Broken Fall", rtcBrokenFall},
		{"Molting Skin", rtcMoltingSkin},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceTo(t, g, game.StepPrecombatMain)
			ench := pushCatalogPermanent(g, me.ID, c.name, "Enchantment", c.oracle, false)
			bear := pr7Creature(g, me.ID, "Bear", 2)
			pr7Activate(t, g, me.ID, ench, 0, game.ActivateAbilityParams{Targets: pr7bTargets(bear)})
			if !me.Hand.Contains(ench) {
				t.Fatalf("%s is not in its owner's hand", c.name)
			}
			if n := regenShieldsOn(t, g, bear); n != 1 {
				t.Fatalf("regeneration shields = %d, want 1", n)
			}
		})
	}
}

// rtcGraveyardCard puts a card into p's graveyard, owned by p.
func rtcGraveyardCard(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Graveyard.PushTop(c)
	return c.InstanceID
}
