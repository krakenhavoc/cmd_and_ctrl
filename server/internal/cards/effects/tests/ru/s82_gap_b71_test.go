package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b71_test.go — Ravnica Remastered creatures, batch 71 (#5948). Each
// test plays the real catalog entry through the engine.

const (
	gapB71FirewheelerOracle = "f81327c2-a0e4-475d-bb89-8a42b1ddc93c"
	gapB71GuildmageOracle   = "650949ca-609b-46c5-bc6b-6f782cdcb6d2"
	gapB71MindleechOracle   = "facc2d6f-3f28-4b87-b643-44fdf939a2b5"
	gapB71AristocratOracle  = "9603e553-f72d-40bf-94ab-596afb7d6db3"
)

func TestGapB71EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB71FirewheelerOracle: "Rakdos Firewheeler",
		gapB71GuildmageOracle:   "Rakdos Guildmage",
		gapB71MindleechOracle:   "Mindleech Mass",
		gapB71AristocratOracle:  "Cartel Aristocrat",
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

func TestGapB71FirewheelerDamagesAnOpponentAndACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	life := opp.Life
	effectstest.AdvanceToMain(t, g)
	effectstest.EnterFromHand(t, g, me.ID, "Rakdos Firewheeler", "Creature — Human Rogue", gapB71FirewheelerOracle)
	effectstest.PickTargetRefs(t, g, me.ID, nil, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	effectstest.PickTargetRefs(t, g, me.ID, nil, game.TargetRef{Kind: game.TargetCard, ID: bear, Slot: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-2)
	}
	if d := effectstest.DamageMarkedOn(g, bear); d != 2 {
		t.Errorf("the creature has %d damage, want 2", d)
	}
}

func TestGapB71FirewheelerHitsAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Walker", TypeLine: "Planeswalker — Test", Counters: map[string]int{"loyalty": 5}})
	effectstest.AdvanceToMain(t, g)
	effectstest.EnterFromHand(t, g, me.ID, "Rakdos Firewheeler", "Creature — Human Rogue", gapB71FirewheelerOracle)
	effectstest.PickTargetRefs(t, g, me.ID, nil, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	effectstest.PickTargetRefs(t, g, me.ID, nil, game.TargetRef{Kind: game.TargetCard, ID: walker, Slot: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, walker, "loyalty"); n != 3 {
		t.Errorf("the planeswalker has %d loyalty, want 3", n)
	}
}

func TestGapB71GuildmageShrinksATargetByDiscarding(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.B21Push(g, me.ID, "Rakdos Guildmage", "Creature — Zombie Shaman", gapB71GuildmageOracle, 2, 2, "B", "R")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	discard := effectstest.HandCardForTest(me, "Discarded", "Sorcery", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "B", 4)
	target := game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, target); err == nil {
		t.Fatal("activated without paying the discard")
	}
	target.DiscardIDs = []uuid.UUID{discard}
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, target); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 1 || tg != 1 {
		t.Errorf("the bear is %d/%d, want 1/1", p, tg)
	}
	if !me.Graveyard.Contains(discard) {
		t.Error("the discarded card is not in the graveyard")
	}
}

func TestGapB71GuildmageMakesAGoblinThatIsExiledAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mage := effectstest.B21Push(g, me.ID, "Rakdos Guildmage", "Creature — Zombie Shaman", gapB71GuildmageOracle, 2, 2, "B", "R")
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "R", 4)
	if err := g.ActivateCatalogAbility(me.ID, mage, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var goblin uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" && c.IsToken() {
			goblin = c.InstanceID
		}
	}
	if goblin == uuid.Nil {
		t.Fatal("no Goblin token")
	}
	if p, tg := effectstest.EffectivePower(t, g, goblin), effectstest.EffectiveToughness(t, g, goblin); p != 2 || tg != 1 {
		t.Errorf("the Goblin is %d/%d, want 2/1", p, tg)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, goblin, "haste") {
		t.Error("the Goblin has no haste")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, goblin) != nil {
		t.Error("the Goblin was not exiled at the beginning of the next end step")
	}
}

func TestGapB71CartelAristocratGainsProtectionFromAChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aristocrat := effectstest.B21Push(g, me.ID, "Cartel Aristocrat", "Creature — Human Advisor", gapB71AristocratOracle, 2, 2, "W", "B")
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, aristocrat, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{aristocrat}}); err == nil {
		t.Fatal("the Aristocrat was allowed to sacrifice itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, aristocrat, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.FindBattlefieldCardForTest(g, fodder) != nil {
		t.Error("the fodder was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	if !effectstest.HasKeywordOnBattlefield(t, g, aristocrat, game.ProtectionFromColor("R")) {
		t.Error("the Aristocrat has no protection from red")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, aristocrat, game.ProtectionFromColor("G")) {
		t.Error("the Aristocrat has protection from a colour that was not chosen")
	}
}

func gapB71MindleechHits(t *testing.T, g *game.Game) (me, opp *game.Player, sorcery uuid.UUID) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	sorcery = effectstest.PushHandCardWithManaCost(opp, "Their Big Sorcery", "Sorcery", "{5}{U}{U}")
	mass := effectstest.B43Catalog(g, me.ID, "Mindleech Mass", "Creature — Horror", gapB71MindleechOracle, 6, 6)
	effectstest.AttackWith(t, g, opp.ID, mass)
	effectstest.PassPriorityAroundTable(t, g)
	return me, opp, sorcery
}

func TestGapB71MindleechMassLooksAndCastsFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, sorcery := gapB71MindleechHits(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, sorcery)
	if !effectstest.Knows(g, opp, sorcery, me.ID) {
		t.Error("the Mass's controller did not look at the hand")
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, sorcery, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	effectstest.CastInsideTheResolution(t, g, sorcery)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(sorcery) {
		t.Error("the resolved sorcery should be in its owner's graveyard")
	}
}

func TestGapB71MindleechMassMayDeclineToLook(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, sorcery := gapB71MindleechHits(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("an offer was made though the player did not look")
	}
	if effectstest.Knows(g, opp, sorcery, me.ID) {
		t.Error("the player learned the hand without looking")
	}
	if !opp.Hand.Contains(sorcery) {
		t.Error("the card left the hand")
	}
}
