package testscm

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The departed Redcap was BLUE. The copy may be pointed at the pro-red
// creature and not at the pro-blue one — the reverse of what its
// black-red graveyard card would allow — and the copy resolves there.
func TestResonatorCopyOfADepartedBlueRedcapRetargetsByItsLastKnownColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	plain, proRed, proBlue := effectstest.ProtectedCreatures(g, opp)
	redcap := effectstest.RedcapTurnedBlueAndKilled(t, g, game.TargetRef{Kind: game.TargetCard, ID: plain})

	p := effectstest.ResonatorCopiesTheRedcapTrigger(t, g, me, redcap)
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
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, proRed); c.DamageMarked != 2 {
		t.Errorf("the pro-red creature has %d damage, want 2 from the copy of the departed blue "+
			"Redcap's trigger", c.DamageMarked)
	}
	if c, _ := effectstest.BattlefieldCard(g, plain); c.DamageMarked != 2 {
		t.Errorf("the plain creature has %d damage, want 2 from the original trigger", c.DamageMarked)
	}
}

// A LIVE Redcap is judged as it is now — black-red — exactly as before:
// the pro-red creature is not offered and the pro-blue one is.
func TestResonatorCopyOfALiveRedcapTriggerRetargetsAsBefore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	plain, proRed, proBlue := effectstest.ProtectedCreatures(g, opp)

	redcap := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: redcap, Name: "Murderous Redcap", TypeLine: "Creature — Goblin Assassin",
		OracleID: effectstest.B40MurderousRedcapOracle, Power: 2, Toughness: 2, Colors: []string{"B", "R"},
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, redcap, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Redcap: %v", err)
	}
	effectstest.PassUntilOnBattlefield(t, g, redcap)
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID,
		game.TargetRef{Kind: game.TargetCard, ID: plain}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}

	p := effectstest.ResonatorCopiesTheRedcapTrigger(t, g, me, redcap)
	if slices.Contains(p.PickTargetCards, proRed) {
		t.Error("a live black-red Redcap's copy may not be pointed at a pro-red creature")
	}
	if !slices.Contains(p.PickTargetCards, proBlue) {
		t.Error("a live black-red Redcap's copy may be pointed at a pro-blue creature")
	}
}
