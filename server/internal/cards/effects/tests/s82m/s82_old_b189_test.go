package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b189_test.go — Guilds of Ravnica, older-sets batch 189 (#5525).
// Every test plays the real catalog entry through the engine.

const (
	oldB189SkylineScout      = "77dfca02-3ca3-4a9c-b3e1-479a951cbfaa"
	oldB189GarrisonSergeant  = "cdf8b698-57d6-4d17-8168-87680060c0d6"
	oldB189GoblinLocksmith   = "d692adcf-8ab5-4208-85d1-c75285a85c59"
	oldB189SmeltWardMinotaur = "ee25680a-d01a-4ddb-908b-27a9bea0ceae"
	oldB189BeamsplitterMage  = "fa09e18c-e7cf-4f08-9cc4-324e36594063"
	oldB189JoinShields       = "1102524d-7287-496f-a18a-5fdae6d4c271"
	oldB189SelectiveSnare    = "85442273-9596-47bf-9b40-6d48fd7183a6"
	oldB189NeverHappened     = "ec4a1a8d-4071-40e9-812e-42c746180d99"
	oldB189GruesomeMenagerie = "f95fda8c-0737-447f-9212-1c459062b9e7"
)

func TestOldB189EveryCardIsRegisteredFull(t *testing.T) {
	effectstest.NewCatalogGame(t)
	for _, id := range []string{oldB189SkylineScout, oldB189GarrisonSergeant, oldB189GoblinLocksmith, oldB189SmeltWardMinotaur,
		oldB189BeamsplitterMage, oldB189JoinShields, oldB189SelectiveSnare, oldB189NeverHappened, oldB189GruesomeMenagerie} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not in the catalog", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB189SkylineScoutGainsFlyingOnlyWhenYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		c := effectstest.B12Push(g, me.ID, "Skyline Scout", "Creature — Human Scout", oldB189SkylineScout, 2, 1)
		effectstest.DeclareAttack(t, g, opp.ID, c)
		effectstest.B06AddMana(me, "W", "W")
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.HasEffectiveKeyword(t, g, c, "flying"); got != pay {
			t.Errorf("pay=%v: flying=%v", pay, got)
		}
	}
	// A pool that cannot pay {1}{W} grants nothing.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := effectstest.B12Push(g, me.ID, "Skyline Scout", "Creature — Human Scout", oldB189SkylineScout, 2, 1)
	effectstest.DeclareAttack(t, g, opp.ID, c)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, c, "flying") {
		t.Error("flying without paying")
	}
}

func TestOldB189GarrisonSergeantHasDoubleStrikeOnlyWhileYouControlAGate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sergeant := effectstest.B12Push(g, me.ID, "Garrison Sergeant", "Creature — Lizard Soldier", oldB189GarrisonSergeant, 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.CheckState(t, g)
	if effectstest.HasEffectiveKeyword(t, g, sergeant, "double strike") {
		t.Fatal("double strike without a Gate")
	}
	effectstest.B12Permanent(g, opp.ID, "Opposing Gate", "Land — Gate")
	effectstest.CheckState(t, g)
	if effectstest.HasEffectiveKeyword(t, g, sergeant, "double strike") {
		t.Fatal("an opponent's Gate gave double strike")
	}
	gate := effectstest.B12Permanent(g, me.ID, "Selesnya Guildgate", "Land — Gate")
	effectstest.CheckState(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, sergeant, "double strike") {
		t.Error("no double strike with a Gate")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(gate) })
	effectstest.CheckState(t, g)
	if effectstest.HasEffectiveKeyword(t, g, sergeant, "double strike") {
		t.Error("double strike after the Gate left")
	}
}

func TestOldB189GoblinLocksmithStopsEveryCreatureWithDefenderFromBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	locksmith := effectstest.B12Push(g, me.ID, "Goblin Locksmith", "Creature — Goblin Rogue", oldB189GoblinLocksmith, 2, 1)
	wall := effectstest.B12Push(g, opp.ID, "Wall of Test", "Creature — Wall", "", 0, 4)
	ownWall := effectstest.B12Push(g, me.ID, "Own Wall", "Creature — Wall", "", 0, 4)
	bear := effectstest.B12Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	for _, id := range []uuid.UUID{wall, ownWall} {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "defender")
				}
			}
		})
	}
	effectstest.DeclareAttack(t, g, opp.ID, locksmith)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"opponent's wall": wall, "own wall": ownWall} {
		if effectstest.AuraRestrictions(t, g, id)&game.CantBlock == 0 {
			t.Errorf("%s can still block", name)
		}
	}
	if effectstest.AuraRestrictions(t, g, bear)&game.CantBlock != 0 {
		t.Error("a creature without defender can't block")
	}
	// The set stays live: a creature with defender that arrives later is bound too.
	late := effectstest.B12Push(g, opp.ID, "Late Wall", "Creature — Wall", "", 0, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == late {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "defender")
			}
		}
	})
	if effectstest.AuraRestrictions(t, g, late)&game.CantBlock == 0 {
		t.Error("a defender that arrived later can block")
	}
}

// oldB189Settle answers the targeted trigger's pick, when one is open, and lets the stack empty.
func oldB189Settle(t *testing.T, g *game.Game, me *game.Player, pick uuid.UUID) {
	t.Helper()
	for j := 0; j < 12; j++ {
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, effectstest.PbCard(pick)); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestOldB189SmeltWardMinotaurStopsAnOpponentsCreatureFromBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Smelt-Ward Minotaur", "Creature — Minotaur Warrior", oldB189SmeltWardMinotaur, 2, 3)
	victim := effectstest.B12Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	other := effectstest.B12Push(g, opp.ID, "Other Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	// A creature spell does not trigger it.
	cr := effectstest.HandCardFull(me, "Test Bear", "Creature — Bear", "{R}", "", []string{"R"})
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.CastSpell(me.ID, cr, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast creature: %v", err)
	}
	oldB189Settle(t, g, me, victim)
	if effectstest.AuraRestrictions(t, g, victim)&game.CantBlock != 0 {
		t.Fatal("a creature spell triggered the Minotaur")
	}
	spell := effectstest.HandCardFull(me, "Spell", "Instant", "{R}", "", []string{"R"})
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB189Settle(t, g, me, victim)
	if effectstest.AuraRestrictions(t, g, victim)&game.CantBlock == 0 {
		t.Error("the chosen creature can still block")
	}
	if effectstest.AuraRestrictions(t, g, other)&game.CantBlock != 0 {
		t.Error("an unchosen creature can't block")
	}
}

func TestOldB189BeamsplitterMageCopiesTheSpellForTheOnlyOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.B12Push(g, me.ID, "Beamsplitter Mage", "Creature — Vedalken Wizard", oldB189BeamsplitterMage, 2, 2)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastFromSeat(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardTargets(mage))
	if n := effectstest.TriggersOnStackFrom(g, mage); n != 1 {
		t.Errorf("%d triggers on the stack, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, mage) {
		t.Error("the Mage survived the Bolt")
	}
	if effects.OnBattlefield(g, bear) {
		t.Error("the copy did not hit the other creature")
	}
	if !effects.OnBattlefield(g, theirs) {
		t.Error("an opponent's creature was hit")
	}
}

func TestOldB189BeamsplitterMageAsksWhichCreatureWhenThereAreSeveral(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mage := effectstest.B12Push(g, me.ID, "Beamsplitter Mage", "Creature — Vedalken Wizard", oldB189BeamsplitterMage, 2, 2)
	one := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	two := effectstest.B12Push(g, me.ID, "Other Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastFromSeat(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardTargets(mage))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, two)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, two) {
		t.Error("the chosen creature survived the copy")
	}
	if !effects.OnBattlefield(g, one) {
		t.Error("the creature that was not chosen died")
	}
}

func TestOldB189BeamsplitterMageDoesNothingWithNoOtherCreatureOfYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.B12Push(g, me.ID, "Beamsplitter Mage", "Creature — Vedalken Wizard", oldB189BeamsplitterMage, 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastFromSeat(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardTargets(mage))
	if n := effectstest.TriggersOnStackFrom(g, mage); n != 0 {
		t.Errorf("%d triggers on the stack, want none: no other creature of yours could be the target", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, theirs) {
		t.Error("an opponent's creature was hit")
	}
}

func TestOldB189BeamsplitterMageIgnoresASpellWithAnotherTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mage := effectstest.B12Push(g, me.ID, "Beamsplitter Mage", "Creature — Vedalken Wizard", oldB189BeamsplitterMage, 2, 2)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastFromSeat(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardTargets(bear))
	if n := effectstest.TriggersOnStackFrom(g, mage); n != 0 {
		t.Errorf("%d triggers on the stack, want none: the spell does not target the Mage", n)
	}
}

func TestOldB189JoinShieldsUntapsAndProtectsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(mine); _ = g.TapTargetForEffect(theirs) })
	effectstest.CastCatalogSpell(t, g, "Join Shields", "Instant", oldB189JoinShields, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HmlTapped(t, g, mine) {
		t.Error("my creature stayed tapped")
	}
	if !effectstest.HmlTapped(t, g, theirs) {
		t.Error("an opponent's creature was untapped")
	}
	for _, kw := range []string{"hexproof", "indestructible"} {
		if !effectstest.HasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("my creature has no %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("an opponent's creature has %s", kw)
		}
	}
	late := effectstest.B12Push(g, me.ID, "Late Bear", "Creature — Bear", "", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, late, "hexproof") {
		t.Error("a creature that arrived after resolution got hexproof")
	}
}

func TestOldB189SelectiveSnareReturnsXCreaturesOfTheNamedType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	elf1 := effectstest.B12Push(g, opp.ID, "Elf One", "Creature — Elf", "", 1, 1)
	elf2 := effectstest.B12Push(g, opp.ID, "Elf Two", "Creature — Elf", "", 1, 1)
	goblin := effectstest.B12Push(g, opp.ID, "Goblin", "Creature — Goblin", "", 1, 1)
	snare := effectstest.HandCardFull(me, "Selective Snare", "Sorcery", "{X}{U}", oldB189SelectiveSnare, []string{"U"})
	effectstest.FillPoolColored(me, "U", 3)
	err := g.CastSpell(me.ID, snare, game.CastSpellParams{Strict: true, XValue: 2, ChosenCreatureType: "Elf",
		Targets: effectstest.CardTargets(elf1, goblin)})
	if err == nil {
		t.Fatal("a Goblin was accepted as a target of the type Elf")
	}
	if err := g.CastSpell(me.ID, snare, game.CastSpellParams{Strict: true, XValue: 2, ChosenCreatureType: "Elf",
		Targets: effectstest.CardTargets(elf1, elf2)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{elf1, elf2} {
		if effects.OnBattlefield(g, id) {
			t.Error("an Elf was not returned")
		}
		if !opp.Hand.Contains(id) {
			t.Error("the Elf is not in its owner's hand")
		}
	}
	if !effects.OnBattlefield(g, goblin) {
		t.Error("the Goblin was returned")
	}
}

func TestOldB189NeverHappenedExilesANonlandCardFromHandOrGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	spell, grave, land := uuid.New(), uuid.New(), uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: spell, Name: "Hand Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	opp.Hand.PushTop(game.Card{InstanceID: land, Name: "Hand Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: grave, Name: "Grave Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Never Happened", "Sorcery", oldB189NeverHappened, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	var c *game.PendingChoice
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoiceRevealedHandPick {
			c = pc
		}
	}
	if c == nil {
		t.Fatalf("no revealed-hand pick is open: %+v", g.PendingChoices)
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.DiscardOptions {
		offered[id] = true
	}
	if offered[land] || !offered[spell] || !offered[grave] {
		t.Errorf("offered %v, want the nonland hand card and the graveyard card only", c.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(c.ID, me.ID, []uuid.UUID{spell}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(spell) {
		t.Error("the chosen card is still in hand")
	}
	if !opp.Graveyard.Contains(grave) || !opp.Hand.Contains(land) {
		t.Error("an unchosen card was touched")
	}
	if z := g.FindCardZoneForEffect(spell); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the chosen card is in %v, want exile", z)
	}
}

func TestOldB189GruesomeMenagerieReturnsOneCreatureOfEachValueOneTwoAndThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	mk := func(name, cost string, p *game.Player) uuid.UUID {
		id := uuid.New()
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", ManaCost: cost,
			Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
		return id
	}
	one, oneB := mk("One", "{G}", me), mk("One B", "{G}", me)
	two, three := mk("Two", "{1}{G}", me), mk("Three", "{2}{G}", me)
	four := mk("Four", "{3}{G}", me)
	theirs := mk("Theirs", "{G}", opp)
	effectstest.CastCatalogSpell(t, g, "Gruesome Menagerie", "Sorcery", oldB189GruesomeMenagerie, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, oneB)
	effectstest.AnswerChooseCards(t, g, me.ID, two)
	effectstest.AnswerChooseCards(t, g, me.ID, three)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"One B": oneB, "Two": two, "Three": three} {
		if !effects.OnBattlefield(g, id) {
			t.Errorf("%s did not return", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"One": one, "Four": four} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("%s left the graveyard", name)
		}
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's graveyard card returned")
	}
}

func TestOldB189GruesomeMenagerieSkipsAValueWithNoCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Three", TypeLine: "Creature — Bear", ManaCost: "{2}{G}",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Gruesome Menagerie", "Sorcery", oldB189GruesomeMenagerie, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, id) {
		t.Error("the only legal choice did not return")
	}
}
