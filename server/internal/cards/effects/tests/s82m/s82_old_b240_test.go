package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB240Survivor = "e6dfac7c-72ed-4db8-97b0-371ed9108b00"
	oldB240Strangle = "c429716b-bd3a-4060-b8a3-82e3d6aecd88"
	oldB240Wall     = "7ab15fcd-8371-4b3c-b534-e9ef44249213"
	oldB240Hour     = "b9ea0d4c-cd71-4817-8213-96899e1a14bb"
	oldB240Tactics  = "95ee46ec-4b60-4fea-a09b-d15c083916b3"
	oldB240Unesh    = "7492ea06-24d5-4a1f-a6ac-c33743e91aa7"
	oldB240Naga     = "92247856-35a4-4c1d-90bc-fa0d28524ecd"
	oldB240Crook    = "afc8df6d-7291-400e-b768-0c0c3e3fdc32"
	oldB240Hydra    = "a2dadaf5-eaa1-4c43-9e21-bb2334ceb918"
	oldB240Dagger   = "459d7385-11b9-4f81-a6c1-0a77d163a0e5"
	oldB240JaceDef  = "8ed8e27b-2137-4e19-99bb-5e0d6d32e3af"
	oldB240Camel    = "525d1562-4fb6-4df3-ab6d-ab6a94c866e0"
)

func TestOldB240CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		oldB240Survivor: "Cunning Survivor", oldB240Strangle: "Sand Strangler", oldB240Wall: "Wall of Forgotten Pharaohs",
		oldB240Hour: "Hour of Devastation", oldB240Tactics: "Unconventional Tactics", oldB240Unesh: "Unesh, Criosphinx Sovereign",
		oldB240Naga: "Sidewinder Naga", oldB240Crook: "Crook of Condemnation", oldB240Hydra: "Ramunap Hydra",
		oldB240Dagger: "Dagger of the Worthy", oldB240JaceDef: "Jace's Defeat", oldB240Camel: "Wretched Camel",
	}
	for oracle, name := range want {
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

// oldB240DesertInYard moves a Desert card into the graveyard the way the game
// does, so the layers hear about it.
func oldB240DesertInYard(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Desert", TypeLine: "Land — Desert", Owner: me.ID, Controller: me.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, id); err != nil {
		t.Fatalf("move the Desert card: %v", err)
	}
}

func oldB240Desert(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Test Desert", "Land — Desert", "", 0, 0)
}

func TestOldB240CunningSurvivorGrowsAndEvadesOnADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cs := effectstest.B12Push(g, me.ID, "Cunning Survivor", "Creature — Human Warrior", oldB240Survivor, 1, 3)
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: me.ID, Source: cs, N: 1, Question: "Discard a card"})
	})
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.B38PT(t, g, cs); p != 2 {
		t.Errorf("power %d, want 2", p)
	}
	if effectstest.AuraRestrictions(t, g, cs)&game.CantBeBlocked == 0 {
		t.Error("it can still be blocked")
	}
}

func TestOldB240SandStranglerNeedsADesertAndMayShoot(t *testing.T) {
	for _, desert := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 3)
		if desert {
			oldB240DesertInYard(t, g, me)
		}
		effectstest.AdvanceToMain(t, g)
		effectstest.EnterFromHand(t, g, me.ID, "Sand Strangler", "Creature — Beast", oldB240Strangle)
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			effectstest.PickCard(t, g, me.ID, victim)
		}
		if desert {
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		}
		effectstest.PassPriorityAroundTable(t, g)
		dead := !g.Battlefield.Contains(victim)
		if dead != desert {
			t.Errorf("desert=%v: the 2/3 victim died = %v", desert, dead)
		}
	}
}

func TestOldB240WallOfForgottenPharaohsPingsOnlyWithADesert(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wall := effectstest.B12Push(g, me.ID, "Wall of Forgotten Pharaohs", "Artifact Creature — Wall", oldB240Wall, 0, 4)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, wall, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err == nil {
		t.Fatal("activated with no Desert")
	}
	oldB240DesertInYard(t, g, me)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, wall, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
}

func TestOldB240HourOfDevastationKillsIndestructibleCreaturesAndSparesBolas(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tough := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Immortal", TypeLine: "Creature — Golem",
		Power: 3, Toughness: 3, Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID})
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 6, 6)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 5}, Owner: opp.ID, Controller: opp.ID})
	bolas := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bolas Walker", TypeLine: "Legendary Planeswalker — Bolas",
		Counters: map[string]int{game.CounterLoyalty: 5}, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Hour of Devastation", "Sorcery", oldB240Hour, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{tough, small, walker} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", effectstest.B12Card(t, g, id).Name)
		}
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the 6/6 died to 5 damage")
	}
	if !g.Battlefield.Contains(bolas) {
		t.Error("the Bolas planeswalker was hit")
	}
}

func TestOldB240UnconventionalTacticsPumpsAndComesBackForAZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Unconventional Tactics", "Sorcery", oldB240Tactics, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.B38PT(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("%d/%d, want 5/5", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("no flying")
	}
	var card uuid.UUID
	for _, c := range me.Graveyard.Cards {
		if c.OracleID == oldB240Tactics {
			card = c.InstanceID
		}
	}
	if card == uuid.Nil {
		t.Fatal("the spell is not in the graveyard")
	}
	// A non-Zombie entering does nothing; a Zombie asks for {W}.
	effectstest.EnterFromHand(t, g, me.ID, "Another Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a non-Zombie asked for the payment")
	}
	effectstest.EnterFromHand(t, g, me.ID, "Test Zombie", "Creature — Zombie", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "W")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) || me.Graveyard.Contains(card) {
		t.Error("paying {W} did not return the card to hand")
	}
}

func TestOldB240UneshRevealsFourWhenASphinxEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lib := effectstest.DftMPLibraryTop(me, effectstest.C2b15Bear(), effectstest.C2b15Forest(), effectstest.C2b15Spell(), effectstest.C2b15Bear(), effectstest.C2b15Forest())
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.EnterFromHand(t, g, me.ID, "Unesh, Criosphinx Sovereign", "Legendary Creature — Sphinx", oldB240Unesh)
	effectstest.PassPriorityAroundTable(t, g)
	oldB081Piles(t, g, me, opp, []uuid.UUID{lib[0], lib[1]}, 0)
	effectstest.PassPriorityAroundTable(t, g)
	inHand, inYard := 0, 0
	for _, id := range lib[:4] {
		switch {
		case me.Hand.Contains(id):
			inHand++
		case me.Graveyard.Contains(id):
			inYard++
		}
	}
	if inHand+inYard != 4 || inHand == 0 || inYard == 0 {
		t.Errorf("hand %d, graveyard %d of the four revealed", inHand, inYard)
	}
	if me.Hand.Size() != hand+inHand || !me.Library.Contains(lib[4]) {
		t.Errorf("hand %d, want %d; the fifth card stayed = %v", me.Hand.Size(), hand+inHand, me.Library.Contains(lib[4]))
	}
	// Another Sphinx entering triggers Unesh again.
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "More", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	effectstest.EnterFromHand(t, g, me.ID, "Other Sphinx", "Creature — Sphinx", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) == nil && effectstest.LatestRevealPickFor(g, opp.ID) == nil {
		t.Errorf("another Sphinx entering did not trigger Unesh: %+v", g.PendingChoices)
	}
}

func TestOldB240NagaAndHydraReadTheDesert(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	naga := effectstest.B12Push(g, me.ID, "Sidewinder Naga", "Creature — Snake Warrior", oldB240Naga, 3, 2)
	hydra := effectstest.B12Push(g, me.ID, "Ramunap Hydra", "Creature — Snake Hydra", oldB240Hydra, 3, 3)
	effectstest.AdvanceToMain(t, g)
	if p, _ := effectstest.B38PT(t, g, naga); p != 3 || effectstest.HasEffectiveKeyword(t, g, naga, "trample") {
		t.Errorf("Naga without a Desert: power %d", p)
	}
	if p, tough := effectstest.B38PT(t, g, hydra); p != 3 || tough != 3 {
		t.Errorf("Hydra without a Desert: %d/%d", p, tough)
	}
	oldB240DesertInYard(t, g, me)
	if p, _ := effectstest.B38PT(t, g, naga); p != 4 || !effectstest.HasEffectiveKeyword(t, g, naga, "trample") {
		t.Errorf("Naga with a Desert card: power %d", p)
	}
	if p, tough := effectstest.B38PT(t, g, hydra); p != 4 || tough != 4 {
		t.Errorf("Hydra with a Desert card: %d/%d, want 4/4", p, tough)
	}
	oldB240Desert(g, me.ID)
	if p, tough := effectstest.B38PT(t, g, hydra); p != 5 || tough != 5 {
		t.Errorf("Hydra with both: %d/%d, want 5/5", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, hydra, "reach") || !effectstest.HasEffectiveKeyword(t, g, hydra, "vigilance") {
		t.Error("the Hydra lost a printed keyword")
	}
}

func TestOldB240CrookOfCondemnationExilesOneCardOrEveryGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	crook := effectstest.PushCatalogPermanent(g, me.ID, "Crook of Condemnation", "Artifact", oldB240Crook, false)
	a := uuid.New()
	b := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: a, Name: "A", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: b, Name: "B", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, crook, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(a)}); err != nil {
		t.Fatalf("exile one: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || !me.Graveyard.Contains(b) {
		t.Error("the first ability did not exile only the target")
	}
	effectstest.ApaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, crook, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("exile all: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(b) || !g.Exile.Contains(b) {
		t.Error("the second ability did not exile every graveyard")
	}
	if g.Battlefield.Contains(crook) {
		t.Error("the Crook was not exiled as the cost")
	}
}

func TestOldB240DaggerOfTheWorthyGivesPlusTwoAndAfflict(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 5)
	dagger := effectstest.PushCatalogPermanent(g, me.ID, "Dagger of the Worthy", "Artifact — Equipment", oldB240Dagger, false)
	effectstest.B64ReadyToAttack(g, bear)
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, dagger, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.B38PT(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	effectstest.LciB1Attack(t, g, bear, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	life := opp.Life
	if err := g.DeclareBlocker(blocker, bear); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("defender life %d, want %d from afflict 1", opp.Life, life-1)
	}
}

func TestOldB240JacesDefeatCountersOnlyBlueSpellsAndScriesForAJaceWalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	red := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: red, Name: "Red Spell", TypeLine: "Sorcery", ManaCost: "{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, red, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast red: %v", err)
	}
	defeat := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: defeat, Name: "Jace's Defeat", TypeLine: "Instant", ManaCost: "{1}{U}", OracleID: oldB240JaceDef, Owner: opp.ID, Controller: opp.ID})
	effectstest.ApaMana(opp, "C", "U")
	if err := g.CastSpell(opp.ID, defeat, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(red)}); err == nil {
		t.Fatal("Jace's Defeat targeted a red spell")
	}
	effectstest.PassPriorityAroundTable(t, g)

	jace := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: jace, Name: "Jace, Test", TypeLine: "Legendary Planeswalker — Jace", ManaCost: "{2}{U}", Colors: []string{"U"},
		Counters: map[string]int{game.CounterLoyalty: 3}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, jace, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Jace: %v", err)
	}
	effectstest.ApaMana(opp, "C", "U")
	if err := g.CastSpell(opp.ID, defeat, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(jace)}); err != nil {
		t.Fatalf("cast Jace's Defeat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(jace) {
		t.Error("the Jace spell resolved")
	}
	if effectstest.ScryChoiceFor(g, opp.ID) == nil {
		t.Error("no scry 2 prompt after countering a Jace planeswalker spell")
	}
}

func TestOldB240WretchedCamelMakesAPlayerDiscardOnlyWithADesert(t *testing.T) {
	for _, desert := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		camel := effectstest.B12Push(g, me.ID, "Wretched Camel", "Creature — Zombie Camel", oldB240Camel, 2, 1)
		if desert {
			oldB240DesertInYard(t, g, me)
		}
		effectstest.AdvanceToMain(t, g)
		hand := opp.Hand.Size()
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(camel, game.DestroyOptions{}) })
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, effectstest.PlayerTarget(opp.ID)); err != nil {
				t.Fatalf("pick the player: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		got := effectstest.DiscardChoiceFor(g, opp.ID) != nil
		if got != desert {
			t.Errorf("desert=%v: the opponent was asked to discard = %v (hand %d)", desert, got, hand)
		}
	}
}
