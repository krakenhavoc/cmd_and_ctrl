package testsg

import (
	"reflect"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gap_b174_cards_test.go — Jumpstart 2022 cards, S82 GAP batch 174
// (#5948). Every test plays the real catalog entry through the engine.

const (
	gapPhalanxTacticsOracle     = "9962a9cc-6e9f-461e-a9d3-8c28ec34caf1"
	gapRighteousnessOracle      = "3f6b2f76-0364-415e-a6a2-e9e5bf31b745"
	gapShimmerdriftValeOracle   = "2ac34f3e-822d-4fde-99ca-a31c4d9503fd"
	gapWarpedLandscapeOracle    = "543e6bb3-a867-43bf-a737-2f5d6d8dc631"
	gapBlightedFenOracle        = "b8f3da11-7c8f-4846-98a6-204bfd8d572b"
	gapPiranhaMarshOracle       = "6bc8fd7e-7616-484a-ac23-04d37c93733b"
	gapSandstoneBridgeOracle    = "08911e8e-cd67-4960-a927-958c33632469"
	gapArlinnVotPOracle         = "a15b83df-91b7-49f0-9bec-5776a3dac8bd"
	gapChandraFlamesFuryOracle  = "408966d2-08c0-4531-a3b4-e666840ee95c"
	gapSarkhanDragonspeakerOrcl = "80c26f58-04b0-4db0-b84e-9b136609abe0"
	gapJaceArcaneStrategistOrcl = "4126bab7-6a2a-43b2-ba65-77299f9bb8cf"
)

func TestGapB174CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapPhalanxTacticsOracle: "Phalanx Tactics", gapRighteousnessOracle: "Righteousness",
		gapShimmerdriftValeOracle: "Shimmerdrift Vale", gapWarpedLandscapeOracle: "Warped Landscape",
		gapBlightedFenOracle: "Blighted Fen", gapPiranhaMarshOracle: "Piranha Marsh",
		gapSandstoneBridgeOracle: "Sandstone Bridge", gapArlinnVotPOracle: "Arlinn, Voice of the Pack",
		gapChandraFlamesFuryOracle: "Chandra, Flame's Fury", gapSarkhanDragonspeakerOrcl: "Sarkhan, the Dragonspeaker",
		gapJaceArcaneStrategistOrcl: "Jace, Arcane Strategist",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registration %+v", name, spec)
		}
	}
}

func TestGapPhalanxTacticsPumpsTheTargetMoreThanTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	target := effectstest.B12Creature(g, me.ID, "Lead Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Phalanx Tactics", "Instant", gapPhalanxTacticsOracle, effectstest.CardTarget(theirs)); err == nil {
		t.Fatal("Phalanx Tactics targeted a creature I don't control")
	}
	effectstest.CastCatalogSpell(t, g, "Phalanx Tactics", "Instant", gapPhalanxTacticsOracle, effectstest.CardTarget(target))
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID][2]int{target: {4, 3}, other: {3, 3}, theirs: {2, 2}} {
		if p, tou := effectstest.PtOf(t, g, id); p != want[0] || tou != want[1] {
			t.Errorf("%s = %d/%d, want %d/%d", effectstest.LayeredCard(t, g, id).Name, p, tou, want[0], want[1])
		}
	}
}

func TestGapRighteousnessPumpsOnlyABlockingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Wall", 0, 4)
	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("block: %v", err)
	}
	for _, bad := range []uuid.UUID{idle, attacker} {
		spell := effectstest.PushCatalogHandCard(me, "Righteousness", "Instant", gapRighteousnessOracle)
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bad)}); err == nil {
			t.Fatal("Righteousness took a creature that isn't blocking")
		}
	}
	spell := effectstest.PushCatalogHandCard(me, "Righteousness", "Instant", gapRighteousnessOracle)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(blocker)}); err != nil {
		t.Fatalf("cast at the blocker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PtOf(t, g, blocker); p != 7 || tou != 11 {
		t.Errorf("blocker = %d/%d, want 7/11", p, tou)
	}
}

func TestGapShimmerdriftValeEntersTappedAndTapsForTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vale := effectstest.B13PlayAs(t, g, 0, "Shimmerdrift Vale", "Snow Land", gapShimmerdriftValeOracle)
	if !effectstest.B13Tapped(t, g, vale) {
		t.Error("Shimmerdrift Vale entered untapped")
	}
	if effectstest.PendingOfKind(g, game.PendingChoiceColor) == nil {
		t.Fatal("no colour prompt as the land entered")
	}
	effectstest.AnswerColor(t, g, me.ID, "G")
	effectstest.Untap(g, vale)
	if err := g.ActivateManaAbility(me.ID, vale, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"G"}) {
		t.Errorf("pool = %v, want [G]", got)
	}
}

func TestGapWarpedLandscapeFetchesABasicTappedForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Warped Landscape", "Land", gapWarpedLandscapeOracle, false)
	forest := effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	nonbasic := effectstest.StapleLibraryCard(me, "Taiga", "Land — Mountain Forest")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without paying {2}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, still := effectstest.AangCardOnBF(g, land); still {
		t.Error("the land should be sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	got, ok := effectstest.AangCardOnBF(g, forest)
	if !ok || !got.Tapped {
		t.Errorf("the Forest should be on the battlefield tapped: %+v %v", got, ok)
	}
	if _, ok := effectstest.AangCardOnBF(g, nonbasic); ok {
		t.Error("a nonbasic land was fetched")
	}
}

func TestGapBlightedFenMakesAnOpponentSacrificeACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	fen := effectstest.PushCatalogPermanent(g, me.ID, "Blighted Fen", "Land", gapBlightedFenOracle, false)
	only := effectstest.B12Creature(g, opp.ID, "Their Only Creature", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Creature", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	if err := g.ActivateCatalogAbility(me.ID, fen, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("activated without paying {4}{B}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, fen, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("targeted myself: it says target opponent")
	}
	if err := g.ActivateCatalogAbility(me.ID, fen, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, opp.ID) == nil {
		t.Fatal("the opponent was not asked to sacrifice a creature")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, only)
	if g.Battlefield.Contains(only) {
		t.Error("the opponent's only creature was not sacrificed")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own creature was sacrificed")
	}
}

func TestGapPiranhaMarshEntersTappedAndDrainsOneLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	before := opp.Life
	marsh := effectstest.B13PlayAs(t, g, g.Turn.ActiveSeat, "Piranha Marsh", "Land", gapPiranhaMarshOracle)
	if !effectstest.B13Tapped(t, g, marsh) {
		t.Error("Piranha Marsh entered untapped")
	}
	pick := effectstest.PassUntilPickTarget(t, g, me.ID)
	if pick == nil {
		t.Fatal("no target was asked for")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("life = %d, want %d", opp.Life, before-1)
	}
}

func TestGapSandstoneBridgeEntersTappedAndPumpsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	bridge := effectstest.B13PlayAs(t, g, g.Turn.ActiveSeat, "Sandstone Bridge", "Land", gapSandstoneBridgeOracle)
	if !effectstest.B13Tapped(t, g, bridge) {
		t.Error("Sandstone Bridge entered untapped")
	}
	pick := effectstest.PassUntilPickTarget(t, g, me.ID)
	if pick == nil {
		t.Fatal("no target was asked for")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PtOf(t, g, bear); p != 3 || tou != 3 {
		t.Errorf("Bear = %d/%d, want 3/3", p, tou)
	}
	if !effectstest.P1g2HasKeyword(g, bear, "vigilance") {
		t.Error("the Bear did not gain vigilance")
	}
}

func TestGapArlinnVoiceOfThePackBuffsWolvesAndWerewolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	arlinn := effectstest.PushWalkerForTest(g, me.ID, "Arlinn, Voice of the Pack", gapArlinnVotPOracle, 7)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	if err := g.ActivateCatalogAbility(me.ID, arlinn, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, arlinn); got != 5 {
		t.Errorf("loyalty = %d, want 5", got)
	}
	var wolf uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Wolf" {
			wolf = c.InstanceID
		}
	}
	if wolf == uuid.Nil {
		t.Fatal("no Wolf token")
	}
	if p, tou := effectstest.PtOf(t, g, wolf); p != 3 || tou != 3 {
		t.Errorf("the Wolf token = %d/%d, want 3/3 (it entered with the counter)", p, tou)
	}
	// A Werewolf creature card entering gets the counter; a Bear does not.
	were := effectstest.B13PlayAs(t, g, g.Turn.ActiveSeat, "Test Werewolf", "Creature — Werewolf", "")
	effectstest.PassPriorityAroundTable(t, g)
	bear := effectstest.B13PlayAs(t, g, g.Turn.ActiveSeat, "Test Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, were, game.CounterPlusOne); n != 1 {
		t.Errorf("Werewolf counters = %d, want 1", n)
	}
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("Bear counters = %d, want 0", n)
	}
}

func TestGapChandraFlamesFuryPlusOneDealsTwoToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	chandra := effectstest.PushWalkerForTest(g, me.ID, "Chandra, Flame's Fury", gapChandraFlamesFuryOracle, 4)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, chandra, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("life = %d, want %d", opp.Life, before-2)
	}
	if got := effectstest.LoyaltyOf(g, chandra); got != 5 {
		t.Errorf("loyalty = %d, want 5", got)
	}
}

func TestGapChandraFlamesFuryMinusTwoHitsTheCreatureAndItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	chandra := effectstest.PushWalkerForTest(g, me.ID, "Chandra, Flame's Fury", gapChandraFlamesFuryOracle, 4)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 4)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, chandra, 1, game.ActivateAbilityParams{Targets: effectstest.CardTarget(ogre)}); err != nil {
		t.Fatalf("-2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("a 3/4 survived 4 damage")
	}
	if opp.Life != before-2 {
		t.Errorf("the creature's controller's life = %d, want %d", opp.Life, before-2)
	}
}

func TestGapChandraFlamesFuryMinusEightBurnsAPlayerAndTheirCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := effectstest.BlbB1Seats(g)
	chandra := effectstest.PushWalkerForTest(g, me.ID, "Chandra, Flame's Fury", gapChandraFlamesFuryOracle, 8)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 9)
	bystander := effectstest.B12Creature(g, third.ID, "Bystander", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, chandra, 2, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("-8: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-10 {
		t.Errorf("life = %d, want %d", opp.Life, before-10)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("their 5/9 survived 10 damage")
	}
	if !g.Battlefield.Contains(bystander) || !g.Battlefield.Contains(mine) {
		t.Error("a creature the target player doesn't control died")
	}
}

func TestGapSarkhanTheDragonspeakerBecomesADragonAndFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	sarkhan := effectstest.PushWalkerForTest(g, me.ID, "Sarkhan, the Dragonspeaker", gapSarkhanDragonspeakerOrcl, 4)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	if err := g.ActivateCatalogAbility(me.ID, sarkhan, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PtOf(t, g, sarkhan); p != 4 || tou != 4 {
		t.Errorf("P/T = %d/%d, want 4/4", p, tou)
	}
	c := effectstest.LayeredCard(t, g, sarkhan)
	if !c.IsCreature() || c.IsPlaneswalker() {
		t.Errorf("types = %v, want a creature that is no longer a planeswalker", c.Effective().Types)
	}
	if !slices.Contains(c.Effective().Subtypes, "Dragon") || !slices.Equal(c.Effective().Colors, []string{"R"}) {
		t.Errorf("subtypes %v colours %v, want a red Dragon", c.Effective().Subtypes, c.Effective().Colors)
	}
	for _, kw := range []string{"flying", "indestructible", "haste"} {
		if !effectstest.P1g2HasKeyword(g, sarkhan, kw) {
			t.Errorf("no %s", kw)
		}
	}
	effectstest.EndTurn(t, g)
	if effectstest.LayeredCard(t, g, sarkhan).IsCreature() {
		t.Error("still a creature after the turn")
	}
	_ = opp
}

func TestGapSarkhanTheDragonspeakerMinusThreeAndTheEmblem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	sarkhan := effectstest.PushWalkerForTest(g, me.ID, "Sarkhan, the Dragonspeaker", gapSarkhanDragonspeakerOrcl, 9)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 4)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	if err := g.ActivateCatalogAbility(me.ID, sarkhan, 1, game.ActivateAbilityParams{Targets: effectstest.CardTarget(ogre)}); err != nil {
		t.Fatalf("-3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("a 3/4 survived 4 damage")
	}
	// Only one loyalty ability a turn: the emblem waits for the next turn.
	effectstest.EndTurn(t, g)
	effectstest.EndTurn(t, g)
	effectstest.EndTurn(t, g)
	effectstest.EndTurn(t, g)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	if err := g.ActivateCatalogAbility(me.ID, sarkhan, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-6: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.EmblemsOf(g, me.ID)); got != 1 {
		t.Fatalf("emblems = %d, want 1", got)
	}
	handBefore := len(me.Hand.Cards)
	if handBefore == 0 {
		t.Fatal("test needs a hand to discard")
	}
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != 0 {
		t.Errorf("hand = %d after my end step, want 0", got)
	}
}

func TestGapJaceArcaneStrategistDrawsAndCountersOnTheSecondDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	jace := effectstest.PushWalkerForTest(g, me.ID, "Jace, Arcane Strategist", gapJaceArcaneStrategistOrcl, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	// The draw step was the first draw of the turn; the +1 is the second.
	handBefore := len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, jace, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	pick := effectstest.PassUntilPickTarget(t, g, me.ID)
	if pick == nil {
		t.Fatal("the second-draw trigger asked for no target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != handBefore+1 {
		t.Errorf("hand = %d, want %d", got, handBefore+1)
	}
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("counters on the Bear = %d, want 1 from the second draw", n)
	}
}

func TestGapJaceArcaneStrategistMinusSevenMakesYourCreaturesUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	jace := effectstest.PushWalkerForTest(g, me.ID, "Jace, Arcane Strategist", gapJaceArcaneStrategistOrcl, 7)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	if err := g.ActivateCatalogAbility(me.ID, jace, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-7: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err == nil {
		t.Error("an unblockable creature was blocked")
	}
}
