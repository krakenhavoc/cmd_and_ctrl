package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b687_test.go — Mercadian Masques creatures, batch 687 (#5619). Every card is the
// real catalog entry played through the engine.

const (
	c687Pangosaur  = "0b350f23-9150-40c7-8ac7-86adfdb8ba52"
	c687Inspector  = "25450b06-3065-4b4f-975e-586d045e5815"
	c687Crawler    = "302c0400-043d-42fa-abbb-488e520116ba"
	c687Grifters   = "399af8ea-7a6f-4f43-bd39-dc29a6e75e06"
	c687Guard      = "3d8aae86-266f-4bfd-a058-0fe12b5ea354"
	c687Bruiser    = "3db407ec-da4b-4b25-b613-e45c164ffe86"
	c687Fugitive   = "4178ae56-2449-4617-aff8-f59d8fb96c66"
	c687TypeLine   = "Creature — Test"
	c687Wall       = "Creature — Wall"
	c687HumanTypes = "Creature — Human"
)

func TestS82OldB687EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		c687Pangosaur: "Pangosaur", c687Inspector: "Port Inspector", c687Crawler: "Rampart Crawler",
		c687Grifters: "Alley Grifters", c687Guard: "Ceremonial Guard", c687Bruiser: "Cho-Arrim Bruiser",
		c687Fugitive: "Skulking Fugitive",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82OldB687PangosaurReturnsToHandWhenAnyPlayerPlaysALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	saur := effectstest.B12Push(g, me.ID, "Pangosaur", "Creature — Dinosaur", c687Pangosaur, 6, 6)
	// A land put onto the battlefield without being played does not trigger it.
	effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(saur) {
		t.Fatal("a land that was not played returned the Pangosaur")
	}
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(saur) {
		t.Fatal("the Pangosaur stayed on the battlefield after a land was played")
	}
	if !me.Hand.Contains(saur) {
		t.Error("the Pangosaur is not in its owner's hand")
	}
}

func TestS82OldB687RampartCrawlerCantBeBlockedByWalls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crawler := effectstest.B12Push(g, me.ID, "Rampart Crawler", "Creature — Lizard Mercenary", c687Crawler, 1, 1)
	wall := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Wall", TypeLine: c687Wall, Power: 0, Toughness: 4})
	bear := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.DeclareAttack(t, g, opp.ID, crawler)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, crawler); err == nil {
		t.Error("a Wall blocked the Rampart Crawler")
	}
	if err := g.DeclareBlocker(bear, crawler); err != nil {
		t.Errorf("a non-Wall could not block it: %v", err)
	}
}

func TestS82OldB687AlleyGriftersMakesTheDefenderDiscardWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for opp.Hand.Size() > 0 {
		_, _ = opp.Hand.PopTop()
	}
	keep := effectstest.EoHand(opp, "Keep", "Instant")
	pitch := effectstest.EoHand(opp, "Pitch", "Instant")
	grifters := effectstest.B12Push(g, me.ID, "Alley Grifters", "Creature — Human Mercenary", c687Grifters, 2, 2)
	wall := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Wall", TypeLine: c687Wall, Power: 0, Toughness: 4})
	effectstest.DeclareAttack(t, g, opp.ID, grifters)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, grifters); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, pitch)
	if opp.Hand.Contains(pitch) || !opp.Hand.Contains(keep) {
		t.Errorf("the defender did not discard the card they chose (hand %d)", opp.Hand.Size())
	}
}

func TestS82OldB687PortInspectorShowsTheDefendersHandToItsControllerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for opp.Hand.Size() > 0 {
		_, _ = opp.Hand.PopTop()
	}
	secret := effectstest.EoHand(opp, "Secret", "Instant")
	inspector := effectstest.B12Push(g, me.ID, "Port Inspector", c687HumanTypes, c687Inspector, 1, 2)
	wall := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Wall", TypeLine: c687Wall, Power: 0, Toughness: 4})
	effectstest.DeclareAttack(t, g, opp.ID, inspector)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, inspector); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	known := func(viewer uuid.UUID) bool {
		for i := range opp.Hand.Cards {
			if opp.Hand.Cards[i].InstanceID == secret {
				return opp.Hand.Cards[i].IsKnownTo(viewer)
			}
		}
		return false
	}
	if !known(me.ID) {
		t.Error("the Inspector's controller did not see the hand")
	}
	if known(g.Seats[2].ID) {
		t.Error("a third player learned the hand")
	}
}

func TestS82OldB687CeremonialGuardIsDestroyedAtEndOfCombatWhenItAttacksOrBlocks(t *testing.T) {
	// Attacking.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := effectstest.B12Push(g, me.ID, "Ceremonial Guard", "Creature — Human Soldier", c687Guard, 3, 4)
	effectstest.DeclareAttack(t, g, opp.ID, guard)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(guard) {
		t.Error("an attacking Ceremonial Guard survived the end of combat")
	}
	// Idle: no attack, no block, no destruction.
	g2 := effectstest.NewCatalogGame(t)
	idle := effectstest.B12Push(g2, g2.Seats[0].ID, "Ceremonial Guard", "Creature — Human Soldier", c687Guard, 3, 4)
	effectstest.AdvanceTo(t, g2, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g2)
	if !g2.Battlefield.Contains(idle) {
		t.Error("a Guard that did not attack was destroyed")
	}
	// Blocking.
	g3 := effectstest.NewCatalogGame(t)
	attacker := effectstest.SeedBear(g3, g3.Seats[0].ID)
	blocker := effectstest.B12Push(g3, g3.Seats[1].ID, "Ceremonial Guard", "Creature — Human Soldier", c687Guard, 3, 4)
	effectstest.DeclareAttack(t, g3, g3.Seats[1].ID, attacker)
	effectstest.AdvanceTo(t, g3, game.StepDeclareBlockers)
	if err := g3.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g3)
	effectstest.AdvanceTo(t, g3, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g3)
	if g3.Battlefield.Contains(blocker) {
		t.Error("a blocking Ceremonial Guard survived the end of combat")
	}
}

func TestS82OldB687ChoArrimBruiserTapsUpToTwoCreaturesWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bruiser := effectstest.B12Push(g, me.ID, "Cho-Arrim Bruiser", "Creature — Ogre Rebel", c687Bruiser, 3, 4)
	a := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "A", TypeLine: c687TypeLine, Power: 1, Toughness: 1})
	b := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "B", TypeLine: c687TypeLine, Power: 1, Toughness: 1})
	c := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "C", TypeLine: c687TypeLine, Power: 1, Toughness: 1})
	effectstest.DeclareAttack(t, g, opp.ID, bruiser)
	effectstest.PickTargetRefs(t, g, me.ID, nil,
		game.TargetRef{Kind: game.TargetCard, ID: a}, game.TargetRef{Kind: game.TargetCard, ID: b})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if tapped, _ := effectstest.BattlefieldCardTapped(g, id); !tapped {
			t.Error("a chosen creature was not tapped")
		}
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, c); tapped {
		t.Error("a creature that was not chosen was tapped")
	}
}

func TestS82OldB687SkulkingFugitiveIsSacrificedWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	fugitive := effectstest.B12Push(g, opp.ID, "Skulking Fugitive", "Creature — Horror Mercenary", c687Fugitive, 3, 4)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: fugitive}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fugitive) {
		t.Error("the Fugitive survived being targeted")
	}
	// A creature the spell does not target is untouched; an untargeted Fugitive stays.
	g2 := effectstest.NewCatalogGame(t)
	calm := effectstest.B12Push(g2, g2.Seats[1].ID, "Skulking Fugitive", "Creature — Horror Mercenary", c687Fugitive, 3, 4)
	effectstest.PassPriorityAroundTable(t, g2)
	if !g2.Battlefield.Contains(calm) {
		t.Error("an untargeted Fugitive left the battlefield")
	}
}
