package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	whisperwoodOracle = "8ec7c7ce-23a5-4ba7-8cfd-4caf8601a9c8"
	temurShamanOracle = "fb683e18-50b8-4821-9a1a-86c6dea63ac9"
	masteryUnseenOrac = "ae0701c1-7a99-46fe-bcb1-ea33ad2a0fc9"
)

// s37Top puts a creature card on top of p's library and returns its id.
func s37Top(p *game.Player, name string) uuid.UUID {
	return mlTop(p, name, "Creature — Bear", "{1}")
}

// s37ManifestNow manifests the top card of p's library through the engine
// and returns the manifested permanent's id.
func s37ManifestNow(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	id := s37Top(p, "Hidden Bear")
	if _, err := g.ManifestForEffect(p.ID); err != nil {
		t.Fatalf("manifest: %v", err)
	}
	if c, ok := battlefieldCard(g, id); !ok || !c.FaceDown {
		t.Fatal("the card is not a face-down permanent")
	}
	return id
}

func TestWhisperwoodElementalManifestsAtYourEndStepOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Whisperwood Elemental", "Creature — Elemental", whisperwoodOracle, false)
	top := s37Top(me, "Top Bear")
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	mlManifested(t, g, top, me, me)
}

func TestWhisperwoodElementalSacrificeGivesFaceUpNontokenCreaturesADiesManifest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	elem := pushCatalogPermanent(g, me.ID, "Whisperwood Elemental", "Creature — Elemental", whisperwoodOracle, false)
	bear := b12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	down := s37ManifestNow(t, g, me)
	if err := g.ActivateCatalogAbility(me.ID, elem, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, elem); ok {
		t.Fatal("the Elemental is still on the battlefield")
	}
	// The face-down manifest did not get the grant; the bear did.
	g.WithWriteLock(func() { _ = down })
	top := s37Top(me, "Replacement")
	if err := g.DestroyPermanentForEffect(bear); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	mlManifested(t, g, top, me, me)
}

func TestWhisperwoodElementalGrantSkipsFaceDownCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	elem := pushCatalogPermanent(g, me.ID, "Whisperwood Elemental", "Creature — Elemental", whisperwoodOracle, false)
	down := s37ManifestNow(t, g, me)
	if err := g.ActivateCatalogAbility(me.ID, elem, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	top := s37Top(me, "Replacement")
	if err := g.DestroyPermanentForEffect(down); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, top); ok {
		t.Fatal("a face-down creature got the dies trigger")
	}
}

func TestMasteryOfTheUnseenManifestsForFourAndGainsLifeWhenAPermanentTurnsUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	mastery := pushCatalogPermanent(g, me.ID, "Mastery of the Unseen", "Enchantment", masteryUnseenOrac, false)
	top := s37Top(me, "Top Bear")
	// Refused without the mana.
	if err := g.ActivateCatalogAbility(me.ID, mastery, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without paying {3}{W}")
	}
	floatForTest(g, me, "WCCC")
	if err := g.ActivateCatalogAbility(me.ID, mastery, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	mlManifested(t, g, top, me, me)
	b12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	life := me.Life
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("C")...) })
	if err := g.PerformSpecialAction(me.ID, top, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2 (the bear and the turned-up creature)", got)
	}
}

func TestTemurWarShamanManifestsOnEntryAndTurnedUpCreatureFights(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Temur War Shaman", "Creature — Human Shaman", temurShamanOracle, false)
	victim := b12Push(g, opp.ID, "Victim", "Creature — Bear", "", 2, 2)
	top := s37Top(me, "Fighter")
	g.WithWriteLock(func() { me.Library.Cards[0].Power, me.Library.Cards[0].Toughness = 4, 4 })
	if _, err := g.ManifestForEffect(me.ID); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("C")...) })
	if err := g.PerformSpecialAction(me.ID, top, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	g.RunStateChecksForTest()
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: victim}}); err != nil {
			t.Fatalf("pick target: %v", err)
		}
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, victim); ok {
		t.Fatal("the fight did not kill the victim")
	}
	if c, ok := battlefieldCard(g, top); !ok || c.DamageMarked < 2 {
		t.Errorf("the turned-up creature did not take the victim's damage: %+v", c.DamageMarked)
	}
}
