package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 430 (Magic 2013): Fungal Sprouting, Index, Revive, Veilborn
// Ghoul, Intrepid Hero, Worldfire, Chandra, the Firebrand, Phyrexian Hulk,
// Hellion Crucible.

const (
	oldB430Sprouting = "303d4533-0c6b-4f93-80e8-b1f5926f4cd4"
	oldB430Index     = "613b824e-39d1-4c6e-a19c-c915abdc860b"
	oldB430Revive    = "51991d28-137e-4690-a674-d852b904bd1a"
	oldB430Ghoul     = "8381b04a-666b-4e33-b277-6db39fdfe905"
	oldB430Hero      = "63c1dfcc-9d47-4dee-b4e6-1f50e716cf9e"
	oldB430Worldfire = "ae0b8c13-0a71-4a60-bf9f-6e2da9503e9c"
	oldB430Chandra   = "536c1112-694f-42f8-a7b1-604389e54a95"
	oldB430Hulk      = "a16509da-ad46-4ee3-86f3-06c521c23481"
	oldB430Crucible  = "c238ef51-4b46-43d5-a70b-40270a96a1fd"
)

func TestOldB430EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB430Sprouting: "Fungal Sprouting", oldB430Index: "Index", oldB430Revive: "Revive",
		oldB430Ghoul: "Veilborn Ghoul", oldB430Hero: "Intrepid Hero", oldB430Worldfire: "Worldfire",
		oldB430Chandra: "Chandra, the Firebrand", oldB430Hulk: "Phyrexian Hulk", oldB430Crucible: "Hellion Crucible",
	} {
		effectstest.CmdrB21AssertFull(t, oracle, name)
	}
}

// --- Fungal Sprouting -------------------------------------------------------

func TestOldB430SproutingMakesSaprolingsEqualToTheGreatestPowerYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 4, 4)
	effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 9, 9)
	effectstest.B12PlayFromHand(t, g, "Fungal Sprouting", "Sorcery", oldB430Sprouting, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Saproling", me.ID); n != 4 {
		t.Errorf("%d Saprolings, want 4 (the greatest power you control)", n)
	}
}

func TestOldB430SproutingWithNoCreatureMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12PlayFromHand(t, g, "Fungal Sprouting", "Sorcery", oldB430Sprouting, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Saproling", me.ID); n != 0 {
		t.Errorf("%d Saprolings with no creature, want 0", n)
	}
}

// --- Index ------------------------------------------------------------------

func TestOldB430IndexLooksAtTheTopFiveAndPutsThemBackInAnyOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	effectstest.B12PlayFromHand(t, g, "Index", "Sorcery", oldB430Index, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LookAtTopChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 5 {
		t.Fatalf("look-at prompt = %+v, want five cards", c)
	}
	rev := []uuid.UUID{c.ScryCards[4], c.ScryCards[3], c.ScryCards[2], c.ScryCards[1], c.ScryCards[0]}
	if err := g.ResolveLookAtTop(c.ID, me.ID, rev); err != nil {
		t.Fatalf("ResolveLookAtTop: %v", err)
	}
	got := effectstest.LibraryTopNames(me, 6)
	if got[0] != "E" || got[4] != "A" || got[5] != "F" {
		t.Errorf("library top is %v, want E D C B A F", got)
	}
}

// --- Revive -----------------------------------------------------------------

func oldB430GraveCard(p *game.Player, name, typeLine string, colors ...string) uuid.UUID {
	id := effectstest.PushGraveyardCardTyped(p, name, typeLine)
	for i := range p.Graveyard.Cards {
		if p.Graveyard.Cards[i].InstanceID == id {
			p.Graveyard.Cards[i].Colors = colors
		}
	}
	return id
}

func TestOldB430ReviveReturnsAGreenCardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	green := oldB430GraveCard(me, "Green Bear", "Creature — Bear", "G")
	red := oldB430GraveCard(me, "Red Imp", "Creature — Imp", "R")
	if err := g.CastSpell(me.ID, effectstest.PushHandCardWithManaCost(me, "Revive", "Sorcery", ""), game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(red)}); err == nil {
		t.Fatal("Revive targeted a red card")
	}
	effectstest.B12PlayFromHand(t, g, "Revive", "Sorcery", oldB430Revive, game.CastSpellParams{Targets: effectstest.CardRefs(green)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, green) {
		t.Error("the green card did not return to hand")
	}
	if !me.Graveyard.Contains(red) {
		t.Error("the red card left the graveyard")
	}
}

func TestOldB430ReviveCannotTargetAnOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := oldB430GraveCard(opp, "Their Elf", "Creature — Elf", "G")
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushHandCardWithManaCost(me, "Revive", "Sorcery", "")
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == id {
			me.Hand.Cards[i].OracleID = oldB430Revive
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Error("Revive targeted a card in an opponent's graveyard")
	}
}

// --- Veilborn Ghoul ---------------------------------------------------------

// oldB430Land puts a land onto the battlefield under p's control and emits its
// enters event, as a land play would.
func oldB430Land(g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: p.ID})
	})
	return id
}

func TestOldB430GhoulCannotBlockAndReturnsWhenASwampEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ghoul := effectstest.SeedGraveyardCard(t, g, "Veilborn Ghoul", "Creature — Zombie", oldB430Ghoul)
	oldB430Land(g, me, "Forest", "Basic Land — Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(ghoul) {
		t.Fatal("a Forest returned the Ghoul")
	}
	oldB430Land(g, me, "Swamp", "Basic Land — Swamp")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, ghoul) {
		t.Errorf("the Ghoul did not return to hand: %v", effectstest.B12ZoneOf(g, ghoul))
	}
}

func TestOldB430GhoulMayDeclineAndIgnoresAnOpponentsSwamp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghoul := effectstest.SeedGraveyardCard(t, g, "Veilborn Ghoul", "Creature — Zombie", oldB430Ghoul)
	oldB430Land(g, opp, "Swamp", "Basic Land — Swamp")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(ghoul) {
		t.Fatal("an opponent's Swamp returned the Ghoul")
	}
	oldB430Land(g, me, "Swamp", "Basic Land — Swamp")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(ghoul) {
		t.Error("the Ghoul returned after declining")
	}
}

func TestOldB430GhoulOnTheBattlefieldCannotBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghoul := effectstest.PushCatalogPermanent(g, me.ID, "Veilborn Ghoul", "Creature — Zombie", oldB430Ghoul, false)
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.DeclareAttack(t, g, me.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(ghoul, attacker); err == nil {
		t.Error("Veilborn Ghoul blocked")
	}
}

// --- Intrepid Hero ----------------------------------------------------------

func TestOldB430HeroDestroysAPowerFourCreatureButNotASmallerOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hero := effectstest.PushCatalogPermanent(g, me.ID, "Intrepid Hero", "Creature — Human Soldier", oldB430Hero, false)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 4, 1)
	if err := g.ActivateCatalogAbility(me.ID, hero, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(small)}); err == nil {
		t.Fatal("targeted a creature with power 3")
	}
	if err := g.ActivateCatalogAbility(me.ID, hero, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(big)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the power-4 creature survived")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the small creature died")
	}
}

func TestOldB430HeroNeedsToTapSoNotWhileSummoningSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hero := effectstest.PushCatalogPermanent(g, me.ID, "Intrepid Hero", "Creature — Human Soldier", oldB430Hero, true)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 5, 5)
	if err := g.ActivateCatalogAbility(me.ID, hero, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(big)}); err == nil {
		t.Error("a summoning-sick Hero tapped")
	}
}

// --- Worldfire --------------------------------------------------------------

func TestOldB430WorldfireExilesEverythingAndSetsEveryLifeToOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	myGrave := effectstest.PushGraveyardCardTyped(me, "Dead Mine", "Creature — Elf")
	theirGrave := effectstest.PushGraveyardCardTyped(opp, "Dead Theirs", "Creature — Elf")
	theirHand := effectstest.PushHandCardWithManaCost(opp, "Their Card", "Instant", "{U}")
	g.WithWriteLock(func() { opp.Life = 3 })
	effectstest.B12PlayFromHand(t, g, "Worldfire", "Sorcery", oldB430Worldfire, game.CastSpellParams{})
	myHandBefore := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, land, theirs} {
		if g.Battlefield.Contains(id) {
			t.Errorf("permanent %s was not exiled", id)
		}
	}
	for _, id := range []uuid.UUID{myGrave, theirGrave, theirHand} {
		if got := effectstest.B12ZoneOf(g, id); got != game.ZoneExile {
			t.Errorf("card %s is in %q, want exile", id, got)
		}
	}
	if len(me.Hand.Cards) != 0 || myHandBefore < 0 {
		t.Errorf("my hand has %d cards, want 0", len(me.Hand.Cards))
	}
	for _, p := range g.Seats {
		if p.Life != 1 {
			t.Errorf("%s has %d life, want 1", p.Name, p.Life)
		}
	}
}

// --- Chandra, the Firebrand -------------------------------------------------

func TestOldB430ChandraPlusOneDealsOneDamageToAnyTarget(t *testing.T) {
	g, seat, me, opp := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Chandra, the Firebrand", oldB430Chandra, 3)
	effectstest.AdvanceToMainOf(t, g, seat)
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("life %d, want %d", opp.Life, before-1)
	}
	if got := effectstest.LoyaltyOf(g, pw); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
}

func TestOldB430ChandraMinusTwoCopiesTheNextInstantOrSorceryOnly(t *testing.T) {
	g, seat, me, _ := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Chandra, the Firebrand", oldB430Chandra, 3)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, pw); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
	// Fungal Sprouting makes two Saprolings (power 2); its copy makes two more.
	effectstest.B12PlayFromHand(t, g, "Fungal Sprouting", "Sorcery", oldB430Sprouting, game.CastSpellParams{})
	for i := 0; i < 4; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.CountOnBattlefield(g, "Saproling", me.ID); n != 4 {
		t.Errorf("%d Saprolings from a copied Fungal Sprouting, want 4", n)
	}
	// The −2 is spent: a second spell is not copied.
	effectstest.B12PlayFromHand(t, g, "Fungal Sprouting", "Sorcery", oldB430Sprouting, game.CastSpellParams{})
	for i := 0; i < 4; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.CountOnBattlefield(g, "Saproling", me.ID); n != 6 {
		t.Errorf("%d Saprolings after a second, uncopied spell, want 6", n)
	}
}

func TestOldB430ChandraMinusSixHitsUpToSixTargets(t *testing.T) {
	g, seat, me, opp := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Chandra, the Firebrand", oldB430Chandra, 6)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Ogre", 7, 7)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Ogre", 6, 6)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Ogre", 7, 7)
	effectstest.AdvanceToMainOf(t, g, seat)
	before := opp.Life
	refs := append(effectstest.CardRefs(a, b), game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{Strict: true, Targets: refs}); err != nil {
		t.Fatalf("−6: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(b) {
		t.Error("the 6/6 survived 6 damage")
	}
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(c) {
		t.Error("a 7/7 died to 6 damage or an untargeted creature was hit")
	}
	if opp.Life != before-6 {
		t.Errorf("life %d, want %d", opp.Life, before-6)
	}
}

func TestOldB430ChandraMinusSixRefusesSevenTargets(t *testing.T) {
	g, seat, me, opp := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Chandra, the Firebrand", oldB430Chandra, 6)
	var ids []uuid.UUID
	for i := 0; i < 7; i++ {
		ids = append(ids, effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2))
	}
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(ids...)}); err == nil {
		t.Error("seven targets were accepted")
	}
}

// --- Phyrexian Hulk ---------------------------------------------------------

func TestOldB430HulkIsAVanilla(t *testing.T) {
	spec, ok := effectstestLookup(oldB430Hulk)
	if !ok || len(spec.Triggered) != 0 || len(spec.Activated) != 0 || len(spec.Static) != 0 {
		t.Errorf("Phyrexian Hulk is not a bare body: %+v", spec)
	}
}

// --- Hellion Crucible -------------------------------------------------------

func TestOldB430CrucibleCountsUpAndMakesAHastyHellion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Hellion Crucible", "Land", oldB430Crucible, false)
	fill := func() { effectstest.FillPoolColored(me, "R", 2) }
	// Not enough counters: the sacrifice ability is refused.
	fill()
	if err := g.ActivateCatalogAbility(me.ID, land, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("made a Hellion with no pressure counters")
	}
	for i := 0; i < 2; i++ {
		effectstest.Untap(g, land)
		fill()
		if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("put a pressure counter %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.LayeredCard(t, g, land).Counters["pressure"]; got != 2 {
		t.Fatalf("%d pressure counters, want 2", got)
	}
	effectstest.Untap(g, land)
	fill()
	if err := g.ActivateCatalogAbility(me.ID, land, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("make the Hellion: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	var hellion game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Hellion" && c.Controller == me.ID {
			hellion = c
		}
	}
	if hellion.InstanceID == uuid.Nil {
		t.Fatal("no Hellion token")
	}
	if p, th := effectstest.EffectivePower(t, g, hellion.InstanceID), effectstest.EffectiveToughness(t, g, hellion.InstanceID); p != 4 || th != 4 {
		t.Errorf("Hellion is %d/%d, want 4/4", p, th)
	}
	if !effectstest.HasEffectiveKeyword(t, g, hellion.InstanceID, "haste") {
		t.Error("the Hellion lacks haste")
	}
}

func effectstestLookup(oracle string) (effects.Spec, bool) { return effects.Lookup(oracle) }
