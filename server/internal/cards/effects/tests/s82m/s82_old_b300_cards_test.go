package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 300 (Oath of the Gatewatch: support, enters and dies triggers,
// removal, equipment, pump and other creatures, #5545). Every test plays the
// real catalog entry through the engine.

const (
	oldB300Unity     = "9fbcae6e-5523-47a3-b0ff-ab7a8bb72822"
	oldB300Press     = "a729a684-9171-4ca9-83ad-1d199afad7bb"
	oldB300Auxiliary = "b74082da-435c-4e45-9424-1188ad7ef5b3"
	oldB300Embodim   = "0e8adf34-0834-4027-9cda-18df69239e2b"
	oldB300Mimic     = "cddb3ec4-01a2-4f4d-a690-3486c73c3df8"
	oldB300Searing   = "2bad07d2-6284-4f87-9e2e-2e9ec5d2dd57"
	oldB300Zone      = "7c1890cf-36db-4720-a472-9158bdc0d8aa"
	oldB300Claws     = "4cacd7de-83ac-4767-b6f2-f5a89f04a0b9"
	oldB300Master    = "83db5c28-e172-4a4d-9324-ed5ade5255bc"
	oldB300Trainer   = "d193f3b7-d205-4c49-ae52-6a095282e0b3"
	oldB300Walker    = "fb818376-6e87-4aa3-a050-4b5f82942593"
	oldB300Deceiver  = "18c58328-518f-4228-bb1b-2433f3fb21f2"
)

func TestOldB300EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB300Unity: "Unity of Purpose", oldB300Press: "Press into Service",
		oldB300Auxiliary: "Joraga Auxiliary", oldB300Embodim: "Embodiment of Fury",
		oldB300Mimic: "Eldrazi Mimic", oldB300Searing: "Searing Light",
		oldB300Zone: "Isolation Zone", oldB300Claws: "Captain's Claws",
		oldB300Master: "Stoneforge Masterwork", oldB300Trainer: "Weapons Trainer",
		oldB300Walker: "Walker of the Wastes", oldB300Deceiver: "Deceiver of Form",
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

func TestOldB300UnityOfPurposeSupportsThenUntapsEveryCounteredCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	fresh := effectstest.B12Creature(g, me.ID, "Fresh", "Creature — Bear", 2, 2)
	old := effectstest.B12Creature(g, me.ID, "Old", "Creature — Bear", 2, 2)
	bare := effectstest.B12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.B28AddCounter(t, g, old, 1)
	effectstest.TapForTest(t, g, fresh, old, bare, theirs)
	effectstest.CastCatalogSpell(t, g, "Unity of Purpose", "Instant", oldB300Unity, effectstest.CardRefs(fresh))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, fresh) != 1 {
		t.Error("the supported creature got no counter")
	}
	if effectstest.TappedForTest(t, g, fresh) || effectstest.TappedForTest(t, g, old) {
		t.Error("a creature with a +1/+1 counter stayed tapped")
	}
	if !effectstest.TappedForTest(t, g, bare) {
		t.Error("a creature with no counter untapped")
	}
	if !effectstest.TappedForTest(t, g, theirs) {
		t.Error("an opponent's creature untapped")
	}
}

func TestOldB300PressIntoServiceSupportsAndSteals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, foe.ID, "Theirs", "Creature — Bear", 3, 3)
	effectstest.TapForTest(t, g, theirs)
	effectstest.CastCatalogSpell(t, g, "Press into Service", "Sorcery", oldB300Press, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: theirs, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, mine) != 1 {
		t.Error("the supported creature got no counter")
	}
	c := effectstest.MustBattlefieldCard(t, g, theirs)
	if c.Controller != me.ID {
		t.Error("control did not change")
	}
	if effectstest.TappedForTest(t, g, theirs) {
		t.Error("the stolen creature is still tapped")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, theirs, "haste") {
		t.Error("the stolen creature has no haste")
	}
}

func TestOldB300JoragaAuxiliarySupportsOthersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	aux := effectstest.B12Push(g, me.ID, "Joraga Auxiliary", "Creature — Elf Soldier Ally", oldB300Auxiliary, 2, 3)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, g.Seats[1].ID, "B", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "CCCCGW")
	if err := g.ActivateCatalogAbility(me.ID, aux, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(aux)}); err == nil {
		t.Fatal("the Auxiliary was accepted as its own support target")
	}
	effectstest.ApaActivate(t, g, me, aux, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(a, b)})
	if effectstest.PlusOnes(g, a) != 1 || effectstest.PlusOnes(g, b) != 1 {
		t.Error("each target should get a +1/+1 counter")
	}
	if effectstest.PlusOnes(g, aux) != 0 {
		t.Error("the Auxiliary got a counter")
	}
}

func TestOldB300EmbodimentOfFuryAnimatesALandOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	embodiment := effectstest.B12Push(g, me.ID, "Embodiment of Fury", "Creature — Elemental", oldB300Embodim, 4, 3)
	old := effectstest.PushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	land := effectstest.PlayLandFromHand(t, g, "Mountain", "")
	effectstest.AnswerPickTarget(t, g, old)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.MustBattlefieldCard(t, g, old)
	if !c.IsCreature() || !c.IsLand() || !c.HasSubtype("Elemental") {
		t.Fatalf("the Forest is %q, want a land creature — Elemental", c.TypeLine)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, old, "haste") || !effectstest.HasKeywordOnBattlefield(t, g, old, "trample") {
		t.Error("the animated land lacks haste or the Embodiment's trample")
	}
	if p := effectstest.RealPowerOf(t, g, old); p != 3 {
		t.Errorf("animated land has power %d, want 3", p)
	}
	if effectstest.HasKeywordOnBattlefield(t, g, embodiment, "trample") == false {
		t.Error("the Embodiment lost its own trample")
	}
	if c := effectstest.MustBattlefieldCard(t, g, land); c.IsCreature() {
		t.Error("the land that entered became a creature though it was not the target")
	}
	// It ends with the turn.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if c := effectstest.MustBattlefieldCard(t, g, old); c.IsCreature() {
		t.Error("the land is still a creature the next turn")
	}
}

func TestOldB300EldraziMimicCopiesAColorlessCreaturesStats(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	mimic := effectstest.B12Push(g, me.ID, "Eldrazi Mimic", "Creature — Eldrazi", oldB300Mimic, 2, 1)
	effectstest.CastHoldCreature(t, g, "Big Drone", "Creature — Eldrazi Drone", "", 5, 4)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.RealPowerOf(t, g, mimic); p != 5 {
		t.Errorf("the Mimic has power %d, want 5", p)
	}
	if c := effectstest.MustBattlefieldCard(t, g, mimic); c.CurrentToughness() != 4 {
		t.Errorf("the Mimic has toughness %d, want 4", c.CurrentToughness())
	}
}

func TestOldB300SearingLightDestroysAnAttackerWithPowerTwoOrLess(t *testing.T) {
	for _, tc := range []struct {
		power int
		ok    bool
	}{{2, true}, {3, false}} {
		g := effectstest.NewCatalogGame(t)
		foe := g.Seats[1]
		attacker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Attacker", TypeLine: "Creature — Bear",
			Power: tc.power, Toughness: 3, Owner: foe.ID, Controller: foe.ID, AttackingTarget: g.Seats[0].ID})
		err := effectstest.CastCatalogSpellErrStrict(t, g, "Searing Light", "Instant", oldB300Searing, effectstest.CardRefs(attacker))
		if (err == nil) != tc.ok {
			t.Errorf("power %d: cast error %v", tc.power, err)
			continue
		}
		if tc.ok {
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.OnBattlefieldByID(g, attacker) {
				t.Error("the attacker survived")
			}
		}
	}
	// A creature that is neither attacking nor blocking is not a target.
	g := effectstest.NewCatalogGame(t)
	idle := effectstest.B12Creature(g, g.Seats[1].ID, "Idle", "Creature — Bear", 1, 1)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Searing Light", "Instant", oldB300Searing, effectstest.CardRefs(idle)); err == nil {
		t.Error("an idle creature was accepted as a target")
	}
}

func TestOldB300IsolationZoneExilesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, foe.ID, "Victim", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	zone := effectstest.CastCatalogSpell(t, g, "Isolation Zone", "Enchantment", oldB300Zone, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
			t.Error("my own creature was accepted as a target")
		}
	}
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, victim) {
		t.Fatal("the creature was not exiled")
	}
	effectstest.B25Destroy(g, zone)
	g.RunStateChecksForTest()
	if n := effectstest.NamedOnBattlefield(g, "Victim", foe.ID); n != 1 {
		t.Errorf("%d of the exiled creature back under its owner's control after Isolation Zone left, want 1", n)
	}
}

func TestOldB300CaptainsClawsPumpsAndMakesAnAttackingKorAlly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, foe := g.Seats[0], g.Seats[1]
	claws := effectstest.B12Push(g, me.ID, "Captain's Claws", "Artifact — Equipment", oldB300Claws, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "C")
	effectstest.ApaActivate(t, g, me, claws, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if p := effectstest.RealPowerOf(t, g, bear); p != 3 {
		t.Errorf("equipped creature has power %d, want 3", p)
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(bear, foe.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	ally := effectstest.BattlefieldTokenNamed(t, g, "Kor Ally")
	if !ally.Tapped || ally.AttackingTarget != foe.ID {
		t.Errorf("the Kor Ally is tapped %v, attacking %v; want tapped and attacking the same player", ally.Tapped, ally.AttackingTarget)
	}
}

func TestOldB300StoneforgeMasterworkCountsCreaturesSharingATypeWithTheHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, foe := g.Seats[0], g.Seats[1]
	work := effectstest.B12Push(g, me.ID, "Stoneforge Masterwork", "Artifact — Equipment", oldB300Master, 0, 0)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Kor Soldier", 2, 2)
	effectstest.B12Creature(g, me.ID, "Kor", "Creature — Kor Wizard", 1, 1)
	effectstest.B12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, foe.ID, "Their Kor", "Creature — Kor", 1, 1)
	effectstest.FloatForTest(g, me, "CC")
	effectstest.ApaActivate(t, g, me, work, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(host)})
	// Two other creatures I control share a type (Kor, Soldier); the Bear and the opponent's Kor don't count.
	if p := effectstest.RealPowerOf(t, g, host); p != 4 {
		t.Errorf("host has power %d, want 4", p)
	}
}

func TestOldB300WeaponsTrainerPumpsOthersOnlyWhileYouControlAnEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, foe := g.Seats[0], g.Seats[1]
	trainer := effectstest.B12Push(g, me.ID, "Weapons Trainer", "Creature — Human Soldier Ally", oldB300Trainer, 3, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, foe.ID, "Theirs", "Creature — Bear", 2, 2)
	if p := effectstest.RealPowerOf(t, g, bear); p != 2 {
		t.Errorf("without an Equipment the Bear has power %d, want 2", p)
	}
	sword := effectstest.B12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	if p := effectstest.RealPowerOf(t, g, bear); p != 3 {
		t.Errorf("with an Equipment the Bear has power %d, want 3", p)
	}
	if p := effectstest.RealPowerOf(t, g, trainer); p != 3 {
		t.Errorf("the Trainer itself has power %d, want 3", p)
	}
	if p := effectstest.RealPowerOf(t, g, theirs); p != 2 {
		t.Errorf("the opponent's creature has power %d, want 2", p)
	}
	effectstest.B25Destroy(g, sword)
	g.RunStateChecksForTest()
	if p := effectstest.RealPowerOf(t, g, bear); p != 2 {
		t.Errorf("after the Equipment left the Bear has power %d, want 2", p)
	}
}

func TestOldB300WalkerOfTheWastesGrowsForEachWastes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, foe := g.Seats[0], g.Seats[1]
	walker := effectstest.B12Push(g, me.ID, "Walker of the Wastes", "Creature — Eldrazi", oldB300Walker, 4, 4)
	effectstest.PushPermanentForTest(g, me.ID, "Wastes", "", "Basic Land")
	effectstest.PushPermanentForTest(g, me.ID, "Wastes", "", "Basic Land")
	effectstest.PushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	effectstest.PushPermanentForTest(g, foe.ID, "Wastes", "", "Basic Land")
	effectstest.AssertKeywords(t, g, walker, "trample")
	if p := effectstest.RealPowerOf(t, g, walker); p != 6 {
		t.Errorf("Walker has power %d, want 6", p)
	}
}

func TestOldB300DeceiverOfFormCopiesTheRevealedCreature(t *testing.T) {
	for _, tc := range []struct{ copies, bottom bool }{{true, true}, {false, false}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		deceiver := effectstest.B12Push(g, me.ID, "Deceiver of Form", "Creature — Eldrazi", oldB300Deceiver, 8, 8)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		ids := effectstest.SeedSearchLibrary(me,
			game.Card{Name: "Top Dragon", TypeLine: "Creature — Dragon", Power: 5, Toughness: 5, ManaCost: "{4}{R}"},
			game.Card{Name: "Second", TypeLine: "Creature — Bear", Power: 1, Toughness: 1, ManaCost: "{G}"},
		)
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerMayChoice(t, g, me.ID, tc.copies)
		effectstest.AnswerMayChoice(t, g, me.ID, tc.bottom)
		effectstest.PassPriorityAroundTable(t, g)
		got := effectstest.MustBattlefieldCard(t, g, bear)
		if (got.Name == "Top Dragon") != tc.copies {
			t.Errorf("copies %v: the Bear is now %q", tc.copies, got.Name)
		}
		if d := effectstest.MustBattlefieldCard(t, g, deceiver); d.Name != "Deceiver of Form" {
			t.Errorf("the Deceiver became %q", d.Name)
		}
		// The library's last card is its top.
		if onBottom := me.Library.Cards[0].InstanceID == ids[0]; onBottom != tc.bottom {
			t.Errorf("bottom %v: the revealed card is on the bottom: %v", tc.bottom, onBottom)
		}
	}
}
