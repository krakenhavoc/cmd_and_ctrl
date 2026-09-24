package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// copy_retarget_departed_source_test.go is the card-level proof for
// #1449: when Strionic Resonator copies a triggered ability whose
// source has left the battlefield, the copy's "you may choose new
// targets" prompt (CR 707.10c) judges the new targets against the
// source AS IT LAST EXISTED there (CR 608.2h), not against its card in
// the graveyard.
//
// The line is #1429's (departed_ability_source_target_test.go): a
// black-red Murderous Redcap's enter trigger targets an opponent's
// creature; in response Cerulean Wisps turns the Redcap blue and it
// dies, so its graveyard card is black-red again. Then the Redcap's
// controller copies the trigger with Strionic Resonator.

// resonatorCopiesTheRedcapTrigger activates Strionic Resonator on the
// Redcap's enter trigger and passes priority until the copy's
// re-target prompt opens. Returns the prompt.
func resonatorCopiesTheRedcapTrigger(t *testing.T, g *game.Game, me *game.Player, redcap uuid.UUID) *game.PendingChoice {
	t.Helper()
	res := pushCatalogPermanent(g, me.ID, "Strionic Resonator", "Artifact", strionicResonatorOracle, false)
	trig := triggerOnStack(g, redcap)
	if trig == nil {
		t.Fatal("setup: the Redcap's enter trigger should be on the stack")
	}
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, res, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: trig.ID}},
	}); err != nil {
		t.Fatalf("activate Strionic Resonator: %v", err)
	}
	for i := 0; i < 8 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority over the Resonator: %v", err)
		}
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the Resonator's copy opened no re-target prompt")
	}
	return p
}

// protectedCreatures seats an opponent's plain creature and one with
// protection from each of red and blue.
func protectedCreatures(g *game.Game, opp *game.Player) (plain, proRed, proBlue uuid.UUID) {
	seat := func(name string, keywords ...string) uuid.UUID {
		return pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Giant",
			Power: 3, Toughness: 5, Colors: []string{"W"}, Keywords: keywords,
			Owner: opp.ID, Controller: opp.ID,
		})
	}
	return seat("Plain Giant"), seat("Pro-Red Giant", "protection from red"), seat("Pro-Blue Giant", "protection from blue")
}

// The departed Redcap was BLUE. The copy may be pointed at the pro-red
// creature and not at the pro-blue one — the reverse of what its
// black-red graveyard card would allow — and the copy resolves there.
func TestResonatorCopyOfADepartedBlueRedcapRetargetsByItsLastKnownColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	plain, proRed, proBlue := protectedCreatures(g, opp)
	redcap := redcapTurnedBlueAndKilled(t, g, game.TargetRef{Kind: game.TargetCard, ID: plain})

	p := resonatorCopiesTheRedcapTrigger(t, g, me, redcap)
	if !slices.Contains(p.PickTargetCards, plain) {
		t.Fatalf("setup: the plain creature must be offered: %v", p.PickTargetCards)
	}
	if !slices.Contains(p.PickTargetCards, proRed) {
		t.Error("the copy's prompt does not offer the pro-red creature: the Redcap was BLUE as it " +
			"last existed (CR 707.10c / 608.2h) — it read the black-red graveyard card (#1449)")
	}
	if slices.Contains(p.PickTargetCards, proBlue) {
		t.Error("the copy's prompt offers the pro-blue creature: the Redcap was BLUE as it last " +
			"existed (CR 707.10c / 608.2h) — it read the black-red graveyard card (#1449)")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: proBlue}}); err == nil {
		t.Fatal("the answer accepted the pro-blue creature as the copy's new target (#1449)")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: proRed}}); err != nil {
		t.Fatalf("the answer refused the pro-red creature: %v (#1449)", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, proRed); c.DamageMarked != 2 {
		t.Errorf("the pro-red creature has %d damage, want 2 from the copy of the departed blue "+
			"Redcap's trigger", c.DamageMarked)
	}
	if c, _ := battlefieldCard(g, plain); c.DamageMarked != 2 {
		t.Errorf("the plain creature has %d damage, want 2 from the original trigger", c.DamageMarked)
	}
}

// A LIVE Redcap is judged as it is now — black-red — exactly as before:
// the pro-red creature is not offered and the pro-blue one is.
func TestResonatorCopyOfALiveRedcapTriggerRetargetsAsBefore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	plain, proRed, proBlue := protectedCreatures(g, opp)

	redcap := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: redcap, Name: "Murderous Redcap", TypeLine: "Creature — Goblin Assassin",
		OracleID: b40MurderousRedcapOracle, Power: 2, Toughness: 2, Colors: []string{"B", "R"},
		Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, redcap, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Redcap: %v", err)
	}
	passUntilOnBattlefield(t, g, redcap)
	b04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(latestPickTarget(g, me.ID).ID, me.ID,
		game.TargetRef{Kind: game.TargetCard, ID: plain}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}

	p := resonatorCopiesTheRedcapTrigger(t, g, me, redcap)
	if slices.Contains(p.PickTargetCards, proRed) {
		t.Error("a live black-red Redcap's copy may not be pointed at a pro-red creature")
	}
	if !slices.Contains(p.PickTargetCards, proBlue) {
		t.Error("a live black-red Redcap's copy may be pointed at a pro-blue creature")
	}
}
