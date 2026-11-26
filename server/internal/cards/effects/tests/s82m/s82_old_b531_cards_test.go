package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b531_cards_test.go — Shadowmoor, batch 531 (S82, #5591): activated
// abilities. Each test plays the real catalog entry through the engine.

const (
	oldB531PuresightOracle   = "012c665c-24b1-49d6-b575-2b4ee1c3a833"
	oldB531KnacksawOracle    = "06af8e20-6f60-400d-9120-c99311232dde"
	oldB531WayfarerOracle    = "06b8acf2-3f2f-4e8f-9584-8d2ffa18a496"
	oldB531KindlerOracle     = "16a96727-c53a-4563-8151-f8b7abb7e534"
	oldB531ScrapbasketOracle = "779269f9-78c1-4e65-bafe-b6dea40a2b24"
	oldB531RaidOracle        = "f46fe37f-c552-44e6-9657-ff63dbbd0ccf"
	oldB531WavebreakersOrcl  = "f6c325e8-b9bd-44ab-9370-196a40c71e1f"
	oldB531SilkbindOracle    = "fbf4ce62-26f1-4886-952a-efdc1f4946b6"
)

func oldB531Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	return g, me, opp
}

func oldB531Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
}

func oldB531Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	return effectstest.B16Tapped(t, g, id)
}

func TestOldB531EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB531PuresightOracle: "Puresight Merrow", oldB531KnacksawOracle: "Knacksaw Clique",
		oldB531WayfarerOracle: "Pale Wayfarer", oldB531KindlerOracle: "Sootstoke Kindler",
		oldB531ScrapbasketOracle: "Scrapbasket", oldB531RaidOracle: "Impromptu Raid",
		oldB531WavebreakersOrcl: "Merrow Wavebreakers", oldB531SilkbindOracle: "Silkbind Faerie",
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

// A {Q} ability needs a tapped source (CR 107.6): refused untapped, paid by
// untapping when tapped.
func TestOldB531SilkbindFaerieTapsATargetAndUntapsItself(t *testing.T) {
	g, me, opp := oldB531Table(t)
	faerie := effectstest.B12Push(g, me.ID, "Silkbind Faerie", "Creature — Faerie Rogue", oldB531SilkbindOracle, 1, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B10AddMana(me, "W", "C")
	params := game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}
	if err := g.ActivateCatalogAbility(me.ID, faerie, 0, params); err == nil {
		t.Fatal("{Q} was paid by an untapped creature")
	}
	oldB531Tap(g, faerie)
	if err := g.ActivateCatalogAbility(me.ID, faerie, 0, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB531Tapped(t, g, bear) {
		t.Error("the target creature was not tapped")
	}
	if oldB531Tapped(t, g, faerie) {
		t.Error("the Faerie is still tapped after paying {Q}")
	}
}

func TestOldB531KnacksawCliqueExilesTheTopCardAndLetsYouPlayIt(t *testing.T) {
	g, me, opp := oldB531Table(t)
	clique := effectstest.B12Push(g, me.ID, "Knacksaw Clique", "Creature — Faerie Rogue", oldB531KnacksawOracle, 1, 4)
	top := effectstest.StackLibrary(opp, "Their Top Card", "Creature — Bear", "{1}{G}")
	oldB531Tap(g, clique)
	effectstest.B10AddMana(me, "U", "C")
	if err := g.ActivateCatalogAbility(me.ID, clique, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("the opponent's top card was not exiled")
	}
	perm := effectstest.ExiledPermission(g, top)
	if !effectstest.PermissionLive(g, perm, me.ID) {
		t.Error("you have no live permission to play the exiled card")
	}
	if effectstest.PermissionLive(g, perm, opp.ID) {
		t.Error("the opponent was given the permission")
	}
	if oldB531Tapped(t, g, clique) {
		t.Error("the Clique is still tapped after paying {Q}")
	}
}

func TestOldB531ScrapbasketBecomesAllColors(t *testing.T) {
	g, me, _ := oldB531Table(t)
	basket := effectstest.B12Push(g, me.ID, "Scrapbasket", "Artifact Creature — Scarecrow", oldB531ScrapbasketOracle, 3, 2)
	if cs := effectstest.EffectiveColors(t, g, basket); len(cs) != 0 {
		t.Fatalf("colors before: %v", cs)
	}
	effectstest.B10AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, basket, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	cs := slices.Clone(effectstest.EffectiveColors(t, g, basket))
	slices.Sort(cs)
	if !slices.Equal(cs, []string{"B", "G", "R", "U", "W"}) {
		t.Errorf("colors after: %v, want all five", cs)
	}
	effectstest.AdvanceToEndStep(t, g)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("advance: %v", err)
	}
	if cs := effectstest.EffectiveColors(t, g, basket); len(cs) != 0 {
		t.Errorf("colors after cleanup: %v, want none", cs)
	}
}

func TestOldB531PuresightMerrowExilesTheTopCardOnlyIfYouSayYes(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g, me, _ := oldB531Table(t)
		merrow := effectstest.B12Push(g, me.ID, "Puresight Merrow", "Creature — Merfolk Wizard", oldB531PuresightOracle, 2, 2)
		top := effectstest.StackLibrary(me, "Top", "Sorcery", "{1}")
		oldB531Tap(g, merrow)
		effectstest.B10AddMana(me, "W")
		effectstest.B16Activate(t, g, me.ID, merrow, 0, game.ActivateAbilityParams{Strict: true})
		effectstest.PassPriorityAroundTable(t, g)
		if yes {
			effectstest.AnswerChooseCards(t, g, me.ID, top)
		} else {
			effectstest.AnswerChooseCards(t, g, me.ID)
		}
		if got := g.Exile.Contains(top); got != yes {
			t.Errorf("yes=%v: card exiled = %v", yes, got)
		}
		if !yes && !me.Library.Contains(top) {
			t.Error("declining took the card off the library")
		}
	}
}

func TestOldB531MerrowWavebreakersGainsFlyingAndUntaps(t *testing.T) {
	g, me, _ := oldB531Table(t)
	merrow := effectstest.B12Push(g, me.ID, "Merrow Wavebreakers", "Creature — Merfolk Soldier", oldB531WavebreakersOrcl, 3, 3)
	oldB531Tap(g, merrow)
	effectstest.B10AddMana(me, "U", "C")
	if effectstest.HasEffectiveKeyword(t, g, merrow, "flying") {
		t.Fatal("flying before the ability")
	}
	effectstest.B16Activate(t, g, me.ID, merrow, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, merrow, "flying") {
		t.Error("no flying after the ability")
	}
	if oldB531Tapped(t, g, merrow) {
		t.Error("still tapped after paying {Q}")
	}
}

func TestOldB531PaleWayfarerAsksTheTargetsControllerForTheColor(t *testing.T) {
	g, me, opp := oldB531Table(t)
	wayfarer := effectstest.B12Push(g, me.ID, "Pale Wayfarer", "Creature — Spirit Giant", oldB531WayfarerOracle, 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	oldB531Tap(g, wayfarer)
	effectstest.B10AddMana(me, "W", "W", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, wayfarer, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.PendingOfKind(g, game.PendingChoiceColor); c == nil || c.Chooser != opp.ID {
		t.Fatalf("the color prompt = %+v, want one for the target's controller", c)
	}
	effectstest.AnswerColor(t, g, opp.ID, "R")
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, theirs), game.ProtectionFromColor("R")) {
		t.Error("the creature did not gain protection from red")
	}
}

func TestOldB531SootstokeKindlerGivesHasteToBlackOrRedCreaturesOnly(t *testing.T) {
	g, me, _ := oldB531Table(t)
	kindler := effectstest.B12Push(g, me.ID, "Sootstoke Kindler", "Creature — Elemental Shaman", oldB531KindlerOracle, 1, 1)
	black := effectstest.B16Creature(g, me.ID, "Black Bear", "Creature — Bear", 2, 2, "B")
	green := effectstest.B16Creature(g, me.ID, "Green Bear", "Creature — Bear", 2, 2, "G")
	if err := g.ActivateCatalogAbility(me.ID, kindler, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(green)}); err == nil {
		t.Fatal("a green creature was accepted")
	}
	effectstest.B16Activate(t, g, me.ID, kindler, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(black)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, black, "haste") {
		t.Error("the black creature has no haste")
	}
	if effectstest.HasEffectiveKeyword(t, g, green, "haste") {
		t.Error("the green creature has haste")
	}
}

func TestOldB531ImpromptuRaidPutsACreatureInWithHasteAndSacrificesItAtTheEnd(t *testing.T) {
	g, me, _ := oldB531Table(t)
	raid := effectstest.B12Push(g, me.ID, "Impromptu Raid", "Enchantment", oldB531RaidOracle, 0, 0)
	_ = raid
	me.Library.Cards = nil
	top := effectstest.PushLibraryTyped(me, "Raider", "Creature — Ogre")
	effectstest.B10AddMana(me, "R", "R", "R")
	effectstest.B16Activate(t, g, me.ID, raid, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(top) {
		t.Fatal("the revealed creature did not enter the battlefield")
	}
	if !effectstest.HasEffectiveKeyword(t, g, top, "haste") {
		t.Error("the creature has no haste")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(top) {
		t.Error("the creature was not sacrificed at the beginning of the end step")
	}
	if !effectstest.GraveyardHas(me, top) {
		t.Error("the sacrificed creature is not in the graveyard")
	}
}

func TestOldB531ImpromptuRaidPutsANoncreatureCardIntoTheGraveyard(t *testing.T) {
	g, me, _ := oldB531Table(t)
	raid := effectstest.B12Push(g, me.ID, "Impromptu Raid", "Enchantment", oldB531RaidOracle, 0, 0)
	me.Library.Cards = nil
	top := effectstest.PushLibraryTyped(me, "Sorcery Card", "Sorcery")
	effectstest.B10AddMana(me, "R", "R", "R")
	effectstest.B16Activate(t, g, me.ID, raid, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.GraveyardHas(me, top) || g.Battlefield.Contains(top) {
		t.Error("a noncreature card was not put into the graveyard")
	}
}
