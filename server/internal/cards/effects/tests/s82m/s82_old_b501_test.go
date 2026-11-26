package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b501_test.go — Alara Reborn, older-sets batch 501 (#5586). Every test
// plays the real catalog entry through the engine.

const (
	oldB501Esper      = "0c59db70-9cb4-4edf-ba6e-12ffa7c06b7b"
	oldB501Wargate    = "26f39abe-ea0b-400d-b11c-03ee4cf50915"
	oldB501Boon       = "28ebe75a-cddd-469b-be8a-286776b83388"
	oldB501Jund       = "84a3ebc2-5c38-40ef-b5f4-62a92cbca493"
	oldB501Sovereigns = "86e13e19-45df-4d73-8600-51b7caa3e6ab"
	oldB501Hulk       = "a1379c25-6daf-4efa-a517-6e6e25c31875"
	oldB501Revelation = "cc519689-031c-48ba-9c79-3f5a435a0848"
	oldB501Appease    = "d3cb9b08-16d2-4574-afb5-068239dbd152"
	oldB501Sages      = "d6c008c7-6c87-4677-ba0e-6d95f10d10bc"
)

func TestOldB501EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB501Esper: "Esper Sojourners", oldB501Wargate: "Wargate", oldB501Boon: "Stormcaller's Boon",
		oldB501Jund: "Jund Sojourners", oldB501Sovereigns: "Sovereigns of Lost Alara", oldB501Hulk: "Glassdust Hulk",
		oldB501Revelation: "Drastic Revelation", oldB501Appease: "Dragon Appeasement", oldB501Sages: "Sages of the Anima",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

// --- Esper Sojourners ---------------------------------------------------------

func TestOldB501EsperCycledTapsAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	hand := me.Hand.Size()
	effectstest.CycleFromHand(t, g, "Esper Sojourners", "Artifact Creature — Vedalken Wizard", oldB501Esper, "{C}{C}{U}")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("the target was not tapped")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (the cycling draw)", me.Hand.Size(), hand+1)
	}
}

func TestOldB501EsperDiesUntapsAPermanentAndFiresOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	effectstest.TapForTest(t, g, land)
	esper := effectstest.PushCatalogPermanent(g, me.ID, "Esper Sojourners", "Artifact Creature — Vedalken Wizard", oldB501Esper, false)
	effectstest.KillCreature(t, g, me.ID, esper)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	if effectstest.B16Tapped(t, g, land) {
		t.Error("the target was not untapped")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Error("the dies trigger fired twice")
	}
}

// --- Jund Sojourners ----------------------------------------------------------

func TestOldB501JundCycledDealsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	effectstest.CycleFromHand(t, g, "Jund Sojourners", "Creature — Lizard Shaman", oldB501Jund, "{C}{C}{R}")
	pick := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	effectstest.AssistCardSettle(t, g, me.ID, &pick, true)
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
}

func TestOldB501JundDiesDealsOneOrDeclines(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		life := opp.Life
		jund := effectstest.PushCatalogPermanent(g, me.ID, "Jund Sojourners", "Creature — Lizard Shaman", oldB501Jund, false)
		effectstest.KillCreature(t, g, me.ID, jund)
		pick := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
		effectstest.AssistCardSettle(t, g, me.ID, &pick, yes)
		want := life
		if yes {
			want = life - 1
		}
		if opp.Life != want {
			t.Errorf("yes=%v: life %d, want %d", yes, opp.Life, want)
		}
	}
}

// --- Wargate ------------------------------------------------------------------

func TestOldB501WargateFindsAPermanentWithManaValueXOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibraryTop(me, "Big Creature", "Creature — Giant", "{4}{G}")
	effectstest.SeedLibraryTop(me, "Plain Spell", "Sorcery", "{G}")
	small := effectstest.SeedLibraryTop(me, "Small Creature", "Creature — Elf", "{1}{G}")
	effectstest.AssistCardCastX(t, g, "Wargate", "Sorcery", oldB501Wargate, "{X}{G}{W}{U}",
		[]string{"G", "W", "U", "C", "C"}, 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if effectstest.SearchOptionNamed(g, c, "Big Creature") != uuid.Nil {
		t.Error("a mana value 5 card was offered with X=2")
	}
	if effectstest.SearchOptionNamed(g, c, "Plain Spell") != uuid.Nil {
		t.Error("a sorcery was offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Small Creature")
	if effectstest.B12ZoneOf(g, small) != game.ZoneBattlefield {
		t.Errorf("the chosen card is in %q, want the battlefield", effectstest.B12ZoneOf(g, small))
	}
}

// --- Stormcaller's Boon -------------------------------------------------------

func TestOldB501BoonSacrificeGivesYourCreaturesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	boon := effectstest.PushCatalogPermanent(g, me.ID, "Stormcaller's Boon", "Enchantment", oldB501Boon, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, boon, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(boon) {
		t.Error("the enchantment was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, mine, "flying") {
		t.Error("my creature did not gain flying")
	}
	if effectstest.P1g2HasKeyword(g, theirs, "flying") {
		t.Error("an opponent's creature gained flying")
	}
}

func TestOldB501BoonCascades(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cheap := effectstest.SeedLibraryTop(me, "Cheap Elf", "Creature — Elf", "{G}")
	_, err := effectstest.AssistCardCastXErr(t, g, "Stormcaller's Boon", "Enchantment", oldB501Boon, "{2}{W}{U}",
		[]string{"W", "U", "C", "C"}, 0, nil)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if effectstest.B12ZoneOf(g, cheap) == game.ZoneLibrary {
		t.Fatal("cascade exiled nothing")
	}
	if effectstest.B12ZoneOf(g, cheap) == game.ZoneExile && effectstest.GrantedPermissionOn(g, me.ID, cheap, game.ZoneExile) == nil {
		t.Error("the cascaded card was exiled and can't be cast")
	}
}

// --- Sovereigns of Lost Alara -------------------------------------------------

func TestOldB501SovereignsFetchesAnAuraForALoneAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibraryTop(me, "Plain Card", "Instant", "{U}")
	aura := effectstest.SeedLibraryTop(me, "Spectral Gift", "Enchantment — Aura", "{1}{W}")
	effectstest.B12Push(g, me.ID, "Sovereigns of Lost Alara", "Creature — Spirit", oldB501Sovereigns, 4, 5)
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		effectstest.AnswerSearchNamed(t, g, me.ID, "Spectral Gift")
	} else {
		for _, pc := range g.PendingChoices {
			t.Logf("pending: kind=%v chooser=%v", pc.Kind, pc.Chooser)
		}
		t.Fatalf("no search prompt")
	}
	card, ok := effectstest.BattlefieldCard(g, aura)
	if !ok {
		t.Fatalf("the Aura is in %q, want the battlefield", effectstest.B12ZoneOf(g, aura))
	}
	if card.AttachedTo.ID != bear {
		t.Errorf("the Aura is attached to %v, want the attacking creature", card.AttachedTo.ID)
	}
}

func TestOldB501SovereignsDoesNothingWhenTwoCreaturesAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	aura := effectstest.SeedLibraryTop(me, "Spectral Gift", "Enchantment — Aura", "{1}{W}")
	effectstest.B12Push(g, me.ID, "Sovereigns of Lost Alara", "Creature — Spirit", oldB501Sovereigns, 4, 5)
	a, b := effectstest.SeedBear(g, me.ID), effectstest.SeedBear(g, me.ID)
	effectstest.AlaB1Attack(t, g, opp.ID, a, b)
	if effectstest.SearchChoiceFor(g, me.ID) != nil || effectstest.B12ZoneOf(g, aura) != game.ZoneLibrary {
		t.Error("a search happened although two creatures attacked")
	}
}

// --- Glassdust Hulk -----------------------------------------------------------

func TestOldB501HulkTriggersOnAnArtifactEnteringButNotOnItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	hulk := effectstest.CastFromHandForTest(t, g, me, "Glassdust Hulk", "Artifact Creature — Golem", "{3}{W}{U}", oldB501Hulk, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	base := effectstest.EffectivePower(t, g, hulk)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, hulk); p != base {
		t.Errorf("the Hulk triggered on itself entering: power %d, want %d", p, base)
	}
	effectstest.CastFromHandForTest(t, g, me, "Trinket", "Artifact", "{0}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, hulk); p != base+1 {
		t.Errorf("power %d after another artifact entered, want %d", p, base+1)
	}
}

// --- Drastic Revelation -------------------------------------------------------

func TestOldB501RevelationDiscardsHandDrawsSevenDiscardsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "L1", "L2", "L3", "L4", "L5", "L6", "L7", "L8", "L9")
	old := effectstest.HandCardForTest(me, "Old Card", "Sorcery", "")
	effectstest.AssistCardCastX(t, g, "Drastic Revelation", "Sorcery", oldB501Revelation, "{2}{U}{B}{R}",
		[]string{"U", "B", "R", "C", "C"}, 0, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, old) != game.ZoneGraveyard {
		t.Error("the old hand card was not discarded")
	}
	if me.Hand.Size() != 4 {
		t.Errorf("hand %d, want 4 (seven drawn, three discarded)", me.Hand.Size())
	}
}

// --- Dragon Appeasement -------------------------------------------------------

func TestOldB501AppeasementSkipsTheDrawStepAndDrawsOnASacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Dragon Appeasement", "Enchantment", oldB501Appease, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(fodder); err != nil {
			t.Errorf("sacrifice: %v", err)
		}
	})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

// --- Sages of the Anima -------------------------------------------------------

func TestOldB501SagesRevealsThreeTakesCreaturesAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Sages of the Anima", "Creature — Elf Wizard", oldB501Sages, 3, 4)
	deep := effectstest.SeedLibraryTop(me, "Deep Card", "Sorcery", "{1}")
	r3 := effectstest.SeedLibraryTop(me, "Rest B", "Instant", "{1}")
	c2 := effectstest.SeedLibraryTop(me, "Bear Two", "Creature — Bear", "{1}")
	r1 := effectstest.SeedLibraryTop(me, "Rest A", "Sorcery", "{1}")
	c1 := effectstest.SeedLibraryTop(me, "Bear One", "Creature — Bear", "{1}")
	_ = r3
	_ = deep
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(me.ID, 1); err != nil {
			t.Errorf("draw: %v", err)
		}
	})
	_ = c2
	_ = r1
	_ = c1
	// the top three are Bear One, Rest A, Bear Two: two creatures, one other.
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d (two creature cards taken, none drawn)", me.Hand.Size(), hand+2)
	}
	if effectstest.B12ZoneOf(g, deep) != game.ZoneLibrary {
		t.Error("the fourth card left the library")
	}
	if got := effectstest.B12ZoneOf(g, r1); got != game.ZoneLibrary {
		t.Errorf("the non-creature card is in %q, want the library (bottom)", got)
	}
}

func TestOldB501HulkCantBeBlockedAfterAnotherArtifactEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hulk := effectstest.B12Push(g, me.ID, "Glassdust Hulk", "Artifact Creature — Golem", oldB501Hulk, 3, 4)
	blocker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastFromHandForTest(t, g, me, "Trinket", "Artifact", "{0}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, hulk); p != 4 {
		t.Fatalf("power %d, want 4", p)
	}
	effectstest.BrAttack(t, g, hulk)
	if err := g.DeclareBlocker(blocker, hulk); err == nil {
		t.Error("a creature blocked the Hulk after another artifact entered")
	}
}

func TestOldB501SagesOrdersTheRestWhenThereAreSeveral(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Sages of the Anima", "Creature — Elf Wizard", oldB501Sages, 3, 4)
	effectstest.SeedLibraryTop(me, "Rest C", "Instant", "{1}")
	effectstest.SeedLibraryTop(me, "Rest B", "Sorcery", "{1}")
	effectstest.SeedLibraryTop(me, "Rest A", "Sorcery", "{1}")
	hand := me.Hand.Size()
	lib := me.Library.Size()
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(me.ID, 1); err != nil {
			t.Errorf("draw: %v", err)
		}
	})
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: no creature among the three", me.Hand.Size(), hand)
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no order prompt for the rest on the bottom")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, order.ScryCards, nil); err != nil {
		t.Fatal(err)
	}
	if me.Library.Size() != lib {
		t.Errorf("library %d, want %d (nothing was drawn)", me.Library.Size(), lib)
	}
	if top := effectstest.LibraryTopNames(me, 1); len(top) == 1 && (top[0] == "Rest A" || top[0] == "Rest B" || top[0] == "Rest C") {
		t.Errorf("a revealed card is still on top: %s", top[0])
	}
}
