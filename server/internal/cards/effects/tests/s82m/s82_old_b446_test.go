package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b446_test.go — Innistrad activated abilities, S82 OLD batch 446
// (#5572), played through the real catalog entries.

const (
	oldB446Elder      = "0431b67e-ce09-4009-ba5b-9fd41e6f1df4"
	oldB446Apprentice = "14bf3049-faa3-4c1f-a613-603e977a50a4"
	oldB446Mystics    = "4418ff27-57c0-4c94-8aee-6151e4f8b0f9"
	oldB446Gargoyle   = "edb6c7ea-b4e1-4cb7-b5a9-8c5644a2aad9"
	oldB446Shrieker   = "f638ceaf-6b44-4d99-95f3-0aa97b1738fe"
	oldB446Phantasm   = "fd26127d-6807-41e3-9dcf-20ef257d71d3"
)

var oldB446MysticsCard = effectstest.IwCard{Oracle: oldB446Mystics, Front: "Ulvenwald Mystics", FrontType: "Creature — Human Shaman Werewolf", Cost: "{2}{G}{G}",
	FrontText: effectstest.IwFrontText, Fp: "3", Ft: "3",
	Back: "Ulvenwald Primordials", BackText: "{G}: Regenerate this creature.\n" + effectstest.IwBackText, Bp: "5", Bt: "5", Colors: []string{"G"}}

func TestOldB446EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB446Elder: "Elder of Laurels", oldB446Apprentice: "Stitcher's Apprentice",
		oldB446Mystics: "Ulvenwald Mystics", oldB446Mystics + "#1": "Ulvenwald Primordials",
		oldB446Gargoyle: "Manor Gargoyle", oldB446Shrieker: "Mindshrieker", oldB446Phantasm: "Mirror-Mad Phantasm",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func TestOldB446ElderOfLaurelsPumpsByCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elder := effectstest.PushCatalogPermanent(g, me.ID, "Elder of Laurels", "Creature — Human Advisor", oldB446Elder, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "G", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, elder, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without a target")
	}
	if err := g.ActivateCatalogAbility(me.ID, elder, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4 (X = my two creatures, not theirs)", p, tt)
	}
}

func TestOldB446ElderOfLaurelsCostsFourMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elder := effectstest.PushCatalogPermanent(g, me.ID, "Elder of Laurels", "Creature — Human Advisor", oldB446Elder, false)
	effectstest.B06AddMana(me, "G", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, elder, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(elder)}); err == nil {
		t.Fatal("activated with three mana")
	}
}

func TestOldB446StitchersApprenticeMakesAHomunculusThenSacrificesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	app := effectstest.PushCatalogPermanent(g, me.ID, "Stitcher's Apprentice", "Creature — Homunculus", oldB446Apprentice, false)
	effectstest.B06AddMana(me, "U", "C")
	if err := g.ActivateCatalogAbility(me.ID, app, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Homunculus"); got != 1 {
		t.Fatalf("%d Homunculus tokens, want 1", got)
	}
	var token uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Homunculus" && c.Controller == me.ID {
			token = c.InstanceID
		}
	}
	if effectstest.SacrificeChoiceFor(g, me.ID) == nil {
		t.Fatal("the controller was not asked which creature to sacrifice")
	}
	effectstest.AnswerSacrifice(t, g, me.ID, token)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Homunculus"); got != 0 {
		t.Errorf("%d Homunculus tokens after sacrificing the token", got)
	}
	if !g.Battlefield.Contains(app) {
		t.Error("the Apprentice was sacrificed")
	}
}

func TestOldB446StitchersApprenticeCanSacrificeItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	app := effectstest.PushCatalogPermanent(g, me.ID, "Stitcher's Apprentice", "Creature — Homunculus", oldB446Apprentice, false)
	effectstest.B06AddMana(me, "U", "C")
	if err := g.ActivateCatalogAbility(me.ID, app, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, app)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(app) {
		t.Error("the Apprentice is still on the battlefield")
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Homunculus"); got != 1 {
		t.Errorf("%d Homunculus tokens, want the new one to stay", got)
	}
}

func TestOldB446UlvenwaldWerewolvesTurnOverAndTheBackRegenerates(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.IwPut(t, g, oldB446MysticsCard, me)
	effectstest.IwNextUpkeep(t, g, map[int]int{})
	if effectstest.IwFace(g, id) != 1 {
		t.Fatalf("Ulvenwald Mystics did not turn into Ulvenwald Primordials, face %d", effectstest.IwFace(g, id))
	}
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	if _, ok := effectstest.BattlefieldCard(g, id); !ok {
		t.Fatal("the regeneration shield did not save Ulvenwald Primordials")
	}
	effectstest.IwNextUpkeep(t, g, map[int]int{1: 2})
	if effectstest.IwFace(g, id) != 0 {
		t.Errorf("Ulvenwald Primordials did not turn back after a player cast two spells, face %d", effectstest.IwFace(g, id))
	}
}

func TestOldB446ManorGargoyleLosesIndestructibleWithDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	garg := effectstest.PushCatalogPermanent(g, me.ID, "Manor Gargoyle", "Artifact Creature — Gargoyle", oldB446Gargoyle, false)
	if !effectstest.HasEffectiveKeyword(t, g, garg, "indestructible") || !effectstest.HasEffectiveKeyword(t, g, garg, "defender") {
		t.Fatal("the Gargoyle lacks defender or indestructible")
	}
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, garg, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, garg, "defender") {
		t.Error("the Gargoyle still has defender")
	}
	if !effectstest.HasEffectiveKeyword(t, g, garg, "flying") {
		t.Error("the Gargoyle did not gain flying")
	}
	if effectstest.HasEffectiveKeyword(t, g, garg, "indestructible") {
		t.Error("the Gargoyle without defender is still indestructible")
	}
}

func TestOldB446MindshriekerGetsTheMilledCardsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bird := effectstest.PushCatalogPermanent(g, me.ID, "Mindshrieker", "Creature — Spirit Bird", oldB446Shrieker, false)
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{3}{R}", Owner: me.ID, Controller: me.ID})
	before := len(me.Library.Cards)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, bird, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Library.Cards) != before-1 {
		t.Errorf("library went from %d to %d cards, want one milled", before, len(me.Library.Cards))
	}
	if p, tt := effectstest.EffectivePower(t, g, bird), effectstest.EffectiveToughness(t, g, bird); p != 5 || tt != 5 {
		t.Errorf("Mindshrieker is %d/%d, want 5/5 (1/1 plus the milled mana value 4)", p, tt)
	}
}

func TestOldB446MindshriekerWithNoMilledCardGainsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := effectstest.PushCatalogPermanent(g, me.ID, "Mindshrieker", "Creature — Spirit Bird", oldB446Shrieker, false)
	g.WithWriteLock(func() { opp.Library.Cards = nil })
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, bird, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bird); p != 1 {
		t.Errorf("Mindshrieker is %d power after milling nothing, want 1", p)
	}
}

func TestOldB446MirrorMadPhantasmShufflesInAndRevealsToAPhantasm(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ph := effectstest.PushCatalogPermanent(g, me.ID, "Mirror-Mad Phantasm", "Creature — Spirit", oldB446Phantasm, false)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	for i := 0; i < 6; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	effectstest.B06AddMana(me, "U", "C")
	if err := g.ActivateCatalogAbility(me.ID, ph, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Mirror-Mad Phantasm"); got != 1 {
		t.Fatalf("%d Mirror-Mad Phantasm on the battlefield, want 1", got)
	}
	// Every card revealed before it went to the graveyard; the rest stay.
	if got := len(me.Graveyard.Cards) + len(me.Library.Cards); got != 6 {
		t.Errorf("graveyard+library hold %d cards, want the 6 fillers", got)
	}
	for _, c := range me.Graveyard.Cards {
		if c.Name != "Filler" {
			t.Errorf("%s reached the graveyard", c.Name)
		}
	}
}

func TestOldB446MirrorMadPhantasmDoesNothingWithoutMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ph := effectstest.PushCatalogPermanent(g, me.ID, "Mirror-Mad Phantasm", "Creature — Spirit", oldB446Phantasm, false)
	effectstest.B06AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, ph, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with one mana")
	}
}
