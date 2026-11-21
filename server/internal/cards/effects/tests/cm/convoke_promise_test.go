package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestWandOfTheWorldsoulGivesTheNextSpellConvokeOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	elves := effectstest.CgTwoGreenCreatures(g, me.ID)
	soldier := effectstest.PushTapCostPermanent(g, me.ID, "Soldier", "Creature — Soldier", "W")

	// With no promise, a plain spell cannot be convoked.
	plain := effectstest.CgBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(plain) })
	if err := g.CastSpell(me.ID, plain.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err == nil {
		t.Fatal("a spell with no convoke was convoked")
	}

	effectstest.CpActivateWand(t, g, me)

	// The colour clause is convoke's: a white creature cannot pay {G},
	// and two green ones pay {1}{G} whole (CR 702.51a).
	if err := g.CastSpell(me.ID, plain.InstanceID, game.CastSpellParams{Strict: true, TapIDs: []uuid.UUID{elves[0], soldier}}); err != nil {
		t.Fatalf("a green and a white creature should pay {1}{G}: %v", err)
	}
	if !effectstest.CgTapped(g, []uuid.UUID{elves[0], soldier}) || effectstest.CgTapped(g, elves[1:]) {
		t.Error("exactly the named creatures should be tapped")
	}
	if !effectstest.CpHadConvoke(g, plain.InstanceID) {
		t.Error("the spell the promise was spent on has convoke")
	}
	if len(effectstest.NsPromises(me)) != 0 {
		t.Error("the next spell spends the promise")
	}

	// The spell after it has no convoke.
	effectstest.PassPriorityAroundTable(t, g)
	second := effectstest.CgBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(second) })
	if err := g.CastSpell(me.ID, second.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves[1:]}); err == nil {
		t.Error("the promise covered a second spell")
	}
	effectstest.FillPoolColored(me, "G", 2)
	if err := g.CastSpell(me.ID, second.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting the second spell for mana: %v", err)
	}
	if effectstest.CpHadConvoke(g, second.InstanceID) {
		t.Error("the second spell was recorded as having convoke")
	}
}

func TestWandOfTheWorldsoulWhiteCreatureCannotPayGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	soldiers := effectstest.PushTapCostSoldiers(g, me.ID, 2)
	effectstest.CpActivateWand(t, g, me)
	bear := effectstest.CgBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(bear) })
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, TapIDs: soldiers}); err == nil {
		t.Error("two white creatures paid {1}{G}")
	}
	if len(effectstest.NsPromises(me)) != 1 {
		t.Error("a refused cast spent the promise")
	}
}

func TestWandOfTheWorldsoulPromiseIsSpentByASpellPaidWithMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.CpActivateWand(t, g, me)
	bear := effectstest.CgBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(bear) })
	effectstest.FillPoolColored(me, "G", 2)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting for mana: %v", err)
	}
	if !effectstest.CpHadConvoke(g, bear.InstanceID) {
		t.Error("the spell has convoke though no creature was tapped")
	}
	if len(effectstest.NsPromises(me)) != 0 {
		t.Error("the spell did not spend the promise")
	}
}

func TestWandOfTheWorldsoulPromiseLapsesAtEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.CpActivateWand(t, g, me)
	for g.Turn.ActiveSeat == 0 && len(effectstest.NsPromises(me)) > 0 {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if len(effectstest.NsPromises(me)) != 0 {
		t.Error("the convoke promise outlived the turn")
	}
}

func TestWandOfTheWorldsoulEntersTappedAndTapsForWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	wand := game.Card{InstanceID: uuid.New(), Name: "Wand of the Worldsoul", TypeLine: "Artifact",
		OracleID: effectstest.WandOfTheWorldsoulOracle, ManaCost: "{2}{W}", Colors: []string{"W"}, Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(wand) })
	effectstest.FillPoolColored(me, "W", 3)
	if err := g.CastSpell(me.ID, wand.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting the Wand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.CgTapped(g, []uuid.UUID{wand.InstanceID}) {
		t.Fatal("the Wand did not enter tapped")
	}
	// Tapped, it can pay neither {T} cost.
	if err := g.ActivateCatalogAbility(me.ID, wand.InstanceID, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a tapped Wand promised convoke")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(wand.InstanceID) })
	before := len(me.ManaPool)
	if err := g.ActivateManaAbility(me.ID, wand.InstanceID, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tapping the Wand for mana: %v", err)
	}
	if len(me.ManaPool) != before+1 || me.ManaPool[len(me.ManaPool)-1].Color != "W" {
		t.Errorf("the Wand should add {W}: pool %+v", me.ManaPool)
	}
}

// The promise and the cast-time record both survive a capture → JSON →
// restore: the promise still gives convoke, and a spell on the stack is
// still one that had it.
func TestConvokePromiseAndHadConvokeSurviveARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	elves := effectstest.CgTwoGreenCreatures(g, me.ID)
	effectstest.CpActivateWand(t, g, me)

	g2 := effectstest.RestoreRoundTrip(t, g, true)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	if got := effectstest.NsPromises(me2); len(got) != 1 || !got[0].Convoke {
		t.Fatalf("the restored promise: %+v", got)
	}
	bear := effectstest.CgBear(me2.ID)
	g2.WithWriteLock(func() { me2.Hand.PushTop(bear) })
	if err := g2.CastSpell(me2.ID, bear.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err != nil {
		t.Fatalf("convoking under the restored promise: %v", err)
	}
	if !effectstest.CpHadConvoke(g2, bear.InstanceID) {
		t.Fatal("the cast did not record convoke")
	}

	g3 := effectstest.RestoreRoundTrip(t, g2, true)
	if !effectstest.CpHadConvoke(g3, bear.InstanceID) {
		t.Error("the restore lost the spell's convoke record")
	}
}

// A spell with improvise of its own that is also promised convoke has
// both (CR 702.51a, 702.126a): artifacts and creatures each pay.
func TestConvokePromiseOnASpellWithImprovisePaysWithBoth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	arts := effectstest.ImpArtifacts(g, me.ID, 3)
	crs := effectstest.PushTapCostSoldiers(g, me.ID, 2)
	blue := effectstest.PushTapCostPermanent(g, me.ID, "Merfolk", "Creature — Merfolk", "U")
	effectstest.CpActivateWand(t, g, me)
	kappa := effectstest.ImpKappa(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(kappa) })
	taps := append(append([]uuid.UUID{}, arts...), crs...)
	taps = append(taps, blue)
	if err := g.CastSpell(me.ID, kappa.InstanceID, game.CastSpellParams{Strict: true, TapIDs: taps}); err != nil {
		t.Fatalf("three artifacts, two creatures and a blue creature should pay {5}{U}: %v", err)
	}
	if !effectstest.CgTapped(g, taps) {
		t.Error("every named permanent should be tapped")
	}
	if !effectstest.CpHadConvoke(g, kappa.InstanceID) {
		t.Error("the spell has convoke beside its improvise")
	}
}
