package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b254_cards_test.go — Aether Revolt, batch 254 (S82 older sets,
// #5536). Every test plays the real catalog entry through the engine.

const (
	oldB254RestorationOracle  = "36f71d49-0f64-4b5d-b1ee-bb0af5e9db9a"
	oldB254ResourcefulOracle  = "952ef791-5c89-4965-9f34-0228acb18bc4"
	oldB254SecretSalvageOracl = "f42fa525-dd17-448a-b0c3-29b90db2e30e"
	oldB254TezzeretMasterOrcl = "3fba19a9-44c6-4771-9587-9d2c4009e996"
	oldB254TezzeretSchemerOrc = "49d64ecb-3a4f-4013-9a34-8b1539a91e96"
	oldB254AjaniOracle        = "542423cb-fbfb-4a2f-9f6d-87b273a95e96"
	oldB254ScuttlerOracle     = "01a6f8e6-439b-4de4-9f0a-21a035fd849e"
	oldB254WheelsmithOracle   = "d3caee32-25b7-4ef7-9f9a-e31d08720086"
	oldB254GarrisonOracle     = "14ef5aa4-5f61-4fbd-9a8d-35920c40ec25"
	oldB254DragsterOracle     = "6aa96a6e-9a88-4662-aa26-b4c16c0ef650"
	oldB254HopeOracle         = "1a20e805-c80b-4710-917b-706d29ec395e"
	oldB254DockhandOracle     = "7cd95b92-8828-4b84-be58-4ce3e948dc8c"
)

func TestOldB254EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB254RestorationOracle, oldB254ResourcefulOracle, oldB254SecretSalvageOracl, oldB254TezzeretMasterOrcl,
		oldB254TezzeretSchemerOrc, oldB254AjaniOracle, oldB254ScuttlerOracle, oldB254WheelsmithOracle,
		oldB254GarrisonOracle, oldB254DragsterOracle, oldB254HopeOracle, oldB254DockhandOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func oldB254Count(g *game.Game, owner uuid.UUID, name string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == owner && c.Name == name {
			n++
		}
	}
	return n
}

func TestOldB254ResourcefulReturnDrawsOnlyWithAnArtifact(t *testing.T) {
	for _, artifact := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, _ := effectstest.SeatsForCombat(g)
		dead := effectstest.GyCard(me, "Dead Bear", "Creature — Bear", "{1}", "")
		if artifact {
			effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
		}
		library := me.Library.Size()
		effectstest.CastCatalogSpell(t, g, "Resourceful Return", "Sorcery", oldB254ResourcefulOracle, effectstest.CardTargets(dead))
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(dead) {
			t.Errorf("artifact=%v: the creature card is not in hand", artifact)
		}
		drew := library - me.Library.Size()
		if want := map[bool]int{false: 0, true: 1}[artifact]; drew != want {
			t.Errorf("artifact=%v: drew %d, want %d", artifact, drew, want)
		}
	}
}

func TestOldB254SalvageScuttlerReturnsAnArtifactOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	scuttler := effectstest.B12Push(g, me.ID, "Salvage Scuttler", "Creature — Crab", oldB254ScuttlerOracle, 4, 4)
	rock := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(scuttler, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("Salvage Scuttler queued no return pick")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{rock}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if !me.Hand.Contains(rock) {
		t.Error("the artifact was not returned to hand")
	}
}

func TestOldB254RenegadeWheelsmithStopsABlockerWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	wheel := effectstest.B12Push(g, me.ID, "Renegade Wheelsmith", "Creature — Dwarf Pilot", oldB254WheelsmithOracle, 3, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wheel, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityUntilChoice(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerPickTarget(t, g, blocker)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if got := effectstest.BlockRefusal(t, g, blocker, wheel); got == "" {
		t.Error("the targeted creature could block")
	}
}

func TestOldB254RestorationSpecialistReturnsAnArtifactAndAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	spec := effectstest.B12Push(g, me.ID, "Restoration Specialist", "Creature — Dwarf Artificer", oldB254RestorationOracle, 2, 1)
	art := effectstest.GyCard(me, "Old Rock", "Artifact", "{1}", "")
	ench := effectstest.GyCard(me, "Old Aura", "Enchantment", "{1}", "")
	other := effectstest.GyCard(me, "Old Bear", "Creature — Bear", "{1}", "")
	effectstest.FillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, spec, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(art, ench)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(art) || !me.Hand.Contains(ench) || me.Hand.Contains(other) {
		t.Error("the artifact and the enchantment were not both returned alone")
	}
	if _, ok := effectstest.BattlefieldCard(g, spec); ok {
		t.Error("the Specialist was not sacrificed")
	}
}

func TestOldB254MobileGarrisonUntapsAnotherOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	garrison := effectstest.B12Push(g, me.ID, "Mobile Garrison", "Artifact Creature — Vehicle", oldB254GarrisonOracle, 3, 4)
	bear := effectstest.B12Creature(g, me.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(garrison, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityUntilChoice(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerPickTarget(t, g, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertTapped(t, g, false, bear)
}

func oldB254DragsterToEndOfCombat(t *testing.T, g *game.Game, me, opp *game.Player, dragster uuid.UUID, attack bool) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if attack {
		if err := g.DeclareAttacker(dragster, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.LockInAttacks(t, g)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB254DaredevilDragsterGetsAVelocityCounterAfterAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	dragster := effectstest.B12Push(g, me.ID, "Daredevil Dragster", "Artifact Creature — Vehicle", oldB254DragsterOracle, 4, 4)
	library := me.Library.Size()
	oldB254DragsterToEndOfCombat(t, g, me, opp, dragster, true)
	if got := effectstest.AangCounterCount(g, dragster, "velocity"); got != 1 {
		t.Fatalf("velocity counters = %d, want 1", got)
	}
	if me.Library.Size() != library {
		t.Error("drew with one counter")
	}
}

func TestOldB254DaredevilDragsterIsQuietWhenItDidNotAttackOrBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	dragster := effectstest.B12Push(g, me.ID, "Daredevil Dragster", "Artifact Creature — Vehicle", oldB254DragsterOracle, 4, 4)
	oldB254DragsterToEndOfCombat(t, g, me, opp, dragster, false)
	if got := effectstest.AangCounterCount(g, dragster, "velocity"); got != 0 {
		t.Errorf("velocity counters = %d, want 0", got)
	}
}

func TestOldB254DaredevilDragsterSacrificesItselfAndDrawsTwoOnTheSecondCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	dragster := effectstest.B12Push(g, me.ID, "Daredevil Dragster", "Artifact Creature — Vehicle", oldB254DragsterOracle, 4, 4)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(dragster, "velocity", 1) })
	library := me.Library.Size()
	oldB254DragsterToEndOfCombat(t, g, me, opp, dragster, true)
	if _, ok := effectstest.BattlefieldCard(g, dragster); ok {
		t.Error("the Vehicle was not sacrificed")
	}
	if got := library - me.Library.Size(); got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
}

func TestOldB254MerchantsDockhandTakesOneOfTheTopX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	dock := effectstest.B12Push(g, me.ID, "Merchant's Dockhand", "Artifact Creature — Construct", oldB254DockhandOracle, 1, 2)
	rockA := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock A", TypeLine: "Artifact"})
	rockB := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock B", TypeLine: "Artifact"})
	top := effectstest.PushLibraryTyped(me, "Third", "Sorcery")
	second := effectstest.PushLibraryTyped(me, "Second", "Sorcery")
	first := effectstest.PushLibraryTyped(me, "First", "Sorcery")
	_ = top
	effectstest.FillPoolColored(me, "U", 4)
	if err := g.ActivateCatalogAbility(me.ID, dock, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, TapIDs: []uuid.UUID{rockA, rockB}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, second)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(second) || me.Hand.Contains(first) {
		t.Error("the chosen card of the top two did not reach the hand alone")
	}
}

func TestOldB254SecretSalvageExilesAndFetchesEveryCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	gone := effectstest.GyCard(me, "Grizzly Test", "Creature — Bear", "{1}{G}", "")
	copyA := effectstest.PushLibraryTyped(me, "Grizzly Test", "Creature — Bear")
	copyB := effectstest.PushLibraryTyped(me, "Grizzly Test", "Creature — Bear")
	effectstest.PushLibraryTyped(me, "Other Card", "Sorcery")
	effectstest.CastCatalogSpell(t, g, "Secret Salvage", "Sorcery", oldB254SecretSalvageOracl, effectstest.CardTargets(gone))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, copyA, copyB)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(gone) {
		t.Error("the targeted card is still in the graveyard")
	}
	if !me.Hand.Contains(copyA) || !me.Hand.Contains(copyB) {
		t.Error("both library copies were not put into the hand")
	}
}

func TestOldB254HopeOfGhirapurBansNoncreatureSpellsForThePlayerItDamaged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	hope := effectstest.B12Push(g, me.ID, "Hope of Ghirapur", "Legendary Artifact Creature — Thopter", oldB254HopeOracle, 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(hope, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	third := g.Seats[2]
	if err := g.ActivateCatalogAbility(me.ID, hope, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(third.ID)}}); err == nil {
		t.Error("a player Hope did not damage was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, hope, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	opt := effectstest.HandCardOf(g, opp, "Opt", "Instant", "{U}", "")
	effectstest.FillPoolColored(opp, "U", 1)
	if err := g.CastSpell(opp.ID, opt, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the damaged player cast a noncreature spell")
	}
}

// oldB254Walker seats a planeswalker on the active player's board in their
// main phase.
func oldB254Walker(t *testing.T, name, oracle string, loyalty int) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	pw := effectstest.PushWalkerForTest(g, me.ID, name, oracle, loyalty)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	return g, me, opp, pw
}

func oldB254Activate(t *testing.T, g *game.Game, me *game.Player, pw uuid.UUID, idx int, targets ...game.TargetRef) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, pw, idx, game.ActivateAbilityParams{Targets: targets}); err != nil {
		t.Fatalf("ability %d: %v", idx, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB254AjaniPlusTwoPutsTwoCountersOnUpToOneCreature(t *testing.T) {
	g, me, _, ajani := oldB254Walker(t, "Ajani, Valiant Protector", oldB254AjaniOracle, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB254Activate(t, g, me, ajani, 0, effectstest.CardTarget(bear)...)
	if got := effectstest.PlusOnes(g, bear); got != 2 {
		t.Errorf("+2 put %d counters, want 2", got)
	}
	if got := effectstest.LoyaltyOf(g, ajani); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestOldB254AjaniPlusOneRevealsToACreatureAndBottomsTheRest(t *testing.T) {
	g, me, _, ajani := oldB254Walker(t, "Ajani, Valiant Protector", oldB254AjaniOracle, 4)
	deep := effectstest.PushLibraryTyped(me, "Deep Card", "Sorcery")
	creature := effectstest.PushLibraryTyped(me, "Library Bear", "Creature — Bear")
	land := effectstest.PushLibraryTyped(me, "Library Land", "Land")
	oldB254Activate(t, g, me, ajani, 1)
	if !me.Hand.Contains(creature) {
		t.Fatal("the revealed creature card is not in hand")
	}
	if len(me.Library.Cards) == 0 || me.Library.Cards[0].InstanceID != land || !me.Library.Contains(deep) {
		t.Error("the revealed non-creature card was not put on the bottom of the library")
	}
}

func TestOldB254AjaniMinusElevenPutsLifeTotalCountersAndTrample(t *testing.T) {
	g, me, _, ajani := oldB254Walker(t, "Ajani, Valiant Protector", oldB254AjaniOracle, 11)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	oldB254Activate(t, g, me, ajani, 2, effectstest.CardTarget(bear)...)
	if got := effectstest.PlusOnes(g, bear); got != life {
		t.Errorf("-11 put %d counters, want %d (the life total)", got, life)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("no trample")
	}
}

func TestOldB254TezzeretMasterPlusOneTakesTheFirstArtifact(t *testing.T) {
	g, me, _, tez := oldB254Walker(t, "Tezzeret, Master of Metal", oldB254TezzeretMasterOrcl, 5)
	deep := effectstest.PushLibraryTyped(me, "Deep Card", "Sorcery")
	art := effectstest.PushLibraryTyped(me, "Library Rock", "Artifact")
	skipped := effectstest.PushLibraryTyped(me, "Library Land", "Land")
	oldB254Activate(t, g, me, tez, 0)
	if !me.Hand.Contains(art) {
		t.Fatal("the artifact card is not in hand")
	}
	if len(me.Library.Cards) == 0 || me.Library.Cards[0].InstanceID != skipped || !me.Library.Contains(deep) {
		t.Error("the revealed non-artifact was not put on the bottom of the library")
	}
}

func TestOldB254TezzeretMasterMinusThreeDrainsByArtifacts(t *testing.T) {
	g, me, opp, tez := oldB254Walker(t, "Tezzeret, Master of Metal", oldB254TezzeretMasterOrcl, 5)
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock A", TypeLine: "Artifact"})
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock B", TypeLine: "Artifact"})
	life := opp.Life
	oldB254Activate(t, g, me, tez, 1, effectstest.PlayerTarget(opp.ID))
	if opp.Life != life-2 {
		t.Errorf("life %d -> %d, want 2 lost (two artifacts)", life, opp.Life)
	}
}

func TestOldB254TezzeretMasterMinusEightTakesArtifactsAndCreatures(t *testing.T) {
	g, me, opp, tez := oldB254Walker(t, "Tezzeret, Master of Metal", oldB254TezzeretMasterOrcl, 8)
	theirRock := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirLand := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"})
	oldB254Activate(t, g, me, tez, 2, effectstest.PlayerTarget(opp.ID))
	for _, id := range []uuid.UUID{theirRock, theirBear} {
		if c, ok := effectstest.BattlefieldCard(g, id); !ok || c.Controller != me.ID {
			t.Errorf("%s was not taken", id)
		}
	}
	if c, ok := effectstest.BattlefieldCard(g, theirLand); !ok || c.Controller != opp.ID {
		t.Error("their land was taken")
	}
}

func TestOldB254TezzeretSchemerPlusOneMakesAnEtheriumCell(t *testing.T) {
	g, me, _, tez := oldB254Walker(t, "Tezzeret the Schemer", oldB254TezzeretSchemerOrc, 5)
	oldB254Activate(t, g, me, tez, 0)
	if oldB254Count(g, me.ID, "Etherium Cell") != 1 {
		t.Fatal("no Etherium Cell token")
	}
}

func TestOldB254TezzeretSchemerMinusTwoIsPlusXMinusX(t *testing.T) {
	g, me, opp, tez := oldB254Walker(t, "Tezzeret the Schemer", oldB254TezzeretSchemerOrc, 5)
	big := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 5)
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock A", TypeLine: "Artifact"})
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock B", TypeLine: "Artifact"})
	oldB254Activate(t, g, me, tez, 1, effectstest.CardTarget(big)...)
	if p, tt := effectstest.PtOf(t, g, big); p != 5 || tt != 3 {
		t.Errorf("the Ogre is %d/%d, want 5/3 (two artifacts)", p, tt)
	}
}

func TestOldB254TezzeretSchemerEmblemAnimatesAnArtifactEachCombat(t *testing.T) {
	g, me, _, tez := oldB254Walker(t, "Tezzeret the Schemer", oldB254TezzeretSchemerOrc, 7)
	rock := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock A", TypeLine: "Artifact"})
	oldB254Activate(t, g, me, tez, 2)
	if got := len(effectstest.EmblemsOf(g, me.ID)); got != 1 {
		t.Fatalf("emblems = %d, want 1", got)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityUntilChoice(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerPickTarget(t, g, rock)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, rock); p != 5 || tt != 5 {
		t.Errorf("the artifact is %d/%d, want 5/5", p, tt)
	}
	if c := effectstest.LayeredCard(t, g, rock); !c.IsCreature() || !c.IsArtifact() {
		t.Error("the artifact is not an artifact creature")
	}
}
