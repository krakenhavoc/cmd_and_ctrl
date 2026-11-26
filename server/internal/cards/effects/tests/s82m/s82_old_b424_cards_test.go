package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b424_cards_test.go — Return to Ravnica, overload, abilities, counters
// and bounce (S82 OLD batch 424, #5566). Each test plays the real catalog entry.

const (
	oldB424SkinOracle        = "02d998a6-0afd-4be1-a209-05618fcc43ea"
	oldB424DynachargeOracle  = "61869f52-51a8-4126-af06-5f5c3d80c28c"
	oldB424TeleportalOracle  = "d5b4876f-d28d-40dc-be7f-2c4f4c0bd6b3"
	oldB424TrickOracle       = "daf22b20-3d87-429c-99fe-2f3d4c6fee24"
	oldB424SentryOracle      = "29e8d973-0233-446e-b587-1c7508352642"
	oldB424TrollOracle       = "8661d773-3b92-4fcf-8ffc-aced1ff7be8a"
	oldB424DoorkeeperOracle  = "8a960cc2-3802-42bc-ba28-941bfe9743be"
	oldB424HeraldOracle      = "b47fd32d-1a56-414f-9d91-ddd46d4b5ac7"
	oldB424GavelOracle       = "5ad33099-0392-4f68-a4c5-7fa3ef2d473e"
	oldB424VoidwielderOracle = "6686b897-64a9-4482-9563-57c88bc83d71"
	oldB424RescueOracle      = "a5386844-ea24-48ee-9337-ae10bcfddfb9"
	oldB424RocOracle         = "b7564801-e35c-4d5c-afa9-99b33b35309c"
)

func TestOldB424EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB424SkinOracle:        "Mizzium Skin",
		oldB424DynachargeOracle:  "Dynacharge",
		oldB424TeleportalOracle:  "Teleportal",
		oldB424TrickOracle:       "Chemister's Trick",
		oldB424SentryOracle:      "Selesnya Sentry",
		oldB424TrollOracle:       "Trestle Troll",
		oldB424DoorkeeperOracle:  "Doorkeeper",
		oldB424HeraldOracle:      "Centaur's Herald",
		oldB424GavelOracle:       "Fall of the Gavel",
		oldB424VoidwielderOracle: "Voidwielder",
		oldB424RescueOracle:      "Dramatic Rescue",
		oldB424RocOracle:         "Skymark Roc",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB424Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestOldB424MizziumSkinGivesOneCreatureToughnessAndHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB424Seats(g)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Mizzium Skin", "Instant", oldB424SkinOracle, effectstest.B16TargetCard(a))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, a); p != 2 || tough != 3 {
		t.Errorf("target is %d/%d, want 2/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, a, "hexproof") {
		t.Error("target has no hexproof")
	}
	if p, tough := effectstest.PtOf(t, g, b); p != 2 || tough != 2 {
		t.Errorf("the other creature is %d/%d, want 2/2", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, b, "hexproof") {
		t.Error("the other creature has hexproof")
	}
}

func TestOldB424MizziumSkinOverloadedAffectsEachCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB424Seats(g)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastWithAltCost(t, g, "Mizzium Skin", "Instant", oldB424SkinOracle, "overload")
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := effectstest.PtOf(t, g, id); p != 2 || tough != 3 {
			t.Errorf("my creature is %d/%d, want 2/3", p, tough)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "hexproof") {
			t.Error("my creature has no hexproof")
		}
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their creature is %d/%d, want 2/2", p, tough)
	}
}

func TestOldB424MizziumSkinRefusesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := oldB424Seats(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Mizzium Skin", "Instant", oldB424SkinOracle, effectstest.B16TargetCard(theirs)); err == nil {
		t.Error("an opponent's creature was accepted")
	}
}

func TestOldB424DynachargePumpsOneOrEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB424Seats(g)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Dynacharge", "Instant", oldB424DynachargeOracle, effectstest.B16TargetCard(a))
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, a); p != 4 {
		t.Errorf("target power %d, want 4", p)
	}
	if p, _ := effectstest.PtOf(t, g, b); p != 2 {
		t.Errorf("the other creature's power %d, want 2", p)
	}
	effectstest.CastWithAltCost(t, g, "Dynacharge", "Instant", oldB424DynachargeOracle, "overload")
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, a); p != 6 {
		t.Errorf("overloaded: first creature power %d, want 6", p)
	}
	if p, _ := effectstest.PtOf(t, g, b); p != 4 {
		t.Errorf("overloaded: second creature power %d, want 4", p)
	}
}

func TestOldB424TeleportalPumpsAndMakesUnblockable(t *testing.T) {
	for _, overload := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, _ := oldB424Seats(g)
		a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
		b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
		if overload {
			effectstest.CastWithAltCost(t, g, "Teleportal", "Sorcery", oldB424TeleportalOracle, "overload")
		} else {
			effectstest.CastCatalogSpell(t, g, "Teleportal", "Sorcery", oldB424TeleportalOracle, effectstest.B16TargetCard(a))
		}
		effectstest.PassPriorityAroundTable(t, g)
		if p, _ := effectstest.PtOf(t, g, a); p != 3 {
			t.Errorf("overload=%v: target power %d, want 3", overload, p)
		}
		if got := effectstest.RestrictionsOf(t, g, a); got&game.CantBeBlocked == 0 {
			t.Errorf("overload=%v: the target is blockable", overload)
		}
		bp, _ := effectstest.PtOf(t, g, b)
		unblockable := effectstest.RestrictionsOf(t, g, b)&game.CantBeBlocked != 0
		if overload != (bp == 3) || overload != unblockable {
			t.Errorf("overload=%v: the other creature has power %d, unblockable %v", overload, bp, unblockable)
		}
	}
}

func TestOldB424ChemistersTrickShrinksAndForcesAnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB424Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 3, 3)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Chemister's Trick", "Instant", oldB424TrickOracle, effectstest.B16TargetCard(a))
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, a); p != 1 {
		t.Errorf("target power %d, want 1", p)
	}
	if p, _ := effectstest.PtOf(t, g, b); p != 3 {
		t.Errorf("the other creature's power %d, want 3", p)
	}
	if !effectstest.HasRequirementRecord(g) {
		t.Error("no attacks-if-able record")
	}
	if p, _ := effectstest.PtOf(t, g, mine); p != 3 {
		t.Errorf("my creature's power %d, want 3", p)
	}
}

func TestOldB424ChemistersTrickOverloadedAffectsEachOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB424Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 3, 3)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 3, 3)
	effectstest.CastWithAltCost(t, g, "Chemister's Trick", "Instant", oldB424TrickOracle, "overload")
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, _ := effectstest.PtOf(t, g, id); p != 1 {
			t.Errorf("their creature's power %d, want 1", p)
		}
	}
	if p, _ := effectstest.PtOf(t, g, mine); p != 3 {
		t.Errorf("my creature's power %d, want 3", p)
	}
}

func TestOldB424ChemistersTrickRefusesYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB424Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	if err := effectstest.CastCatalogSpellErr(t, g, "Chemister's Trick", "Instant", oldB424TrickOracle, effectstest.B16TargetCard(mine)); err == nil {
		t.Error("my own creature was accepted")
	}
}

func TestOldB424RegeneratingCreaturesShieldThemselves(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		mana                   []string
		p, tough               int
	}{
		{"Selesnya Sentry", oldB424SentryOracle, "Creature — Elephant Soldier", []string{"C", "C", "C", "C", "C", "G"}, 3, 2},
		{"Trestle Troll", oldB424TrollOracle, "Creature — Troll", []string{"C", "B", "G"}, 1, 4},
	} {
		g := effectstest.NewCatalogGame(t)
		me, _ := oldB424Seats(g)
		effectstest.AdvanceToMain(t, g)
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
		effectstest.RavB2MustActivate(t, g, me.ID, id, 0, game.ActivateAbilityParams{}, tc.mana...)
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(id, game.DestroyOptions{}); err != nil {
				t.Fatalf("%s: destroy: %v", tc.name, err)
			}
		})
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed through its regeneration shield", tc.name)
		}
	}
}

func TestOldB424TrestleTrollHasDefenderAndReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB424Seats(g)
	id := effectstest.B12Push(g, me.ID, "Trestle Troll", "Creature — Troll", oldB424TrollOracle, 1, 4)
	for _, kw := range []string{"defender", "reach"} {
		if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

func TestOldB424DoorkeeperMillsByCreaturesWithDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB424Seats(g)
	effectstest.AdvanceToMain(t, g)
	door := effectstest.B12Push(g, me.ID, "Doorkeeper", "Creature — Homunculus", oldB424DoorkeeperOracle, 0, 4)
	effectstest.ReqCreature(g, me.ID, "Wall", "", 0, 4, "defender")
	effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2)
	effectstest.ReqCreature(g, opp.ID, "Their Wall", "", 0, 4, "defender")
	for i := 0; i < 6; i++ {
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	lib, gy := opp.Library.Size(), opp.Graveyard.Size()
	effectstest.RavB2MustActivate(t, g, me.ID, door, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}, "C", "C", "U")
	// Doorkeeper and the Wall have defender; the Bear and their Wall do not count.
	if got := lib - opp.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
	if got := opp.Graveyard.Size() - gy; got != 2 {
		t.Errorf("%d cards reached the graveyard, want 2", got)
	}
}

func TestOldB424CentaursHeraldSacrificesForACentaur(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB424Seats(g)
	effectstest.AdvanceToMain(t, g)
	herald := effectstest.B12Push(g, me.ID, "Centaur's Herald", "Creature — Elf Scout", oldB424HeraldOracle, 0, 1)
	effectstest.RavB2MustActivate(t, g, me.ID, herald, 0, game.ActivateAbilityParams{}, "C", "C", "G")
	if g.Battlefield.Contains(herald) {
		t.Error("the Herald was not sacrificed")
	}
	var centaur game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Centaur" && c.Controller == me.ID {
			centaur = c
		}
	}
	if !centaur.IsToken() || centaur.Power != 3 || centaur.Toughness != 3 || !centaur.HasColor("G") {
		t.Errorf("token = %+v, want a 3/3 green Centaur token", centaur)
	}
}

func TestOldB424FallOfTheGavelCountersAndGainsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	// The opponent's spell on the stack, then my Fall of the Gavel in response.
	bolt := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Test Sorcery", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	life := me.Life
	gavel := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: gavel, Name: "Fall of the Gavel", TypeLine: "Instant", OracleID: oldB424GavelOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, gavel, game.CastSpellParams{Strict: true, Targets: effectstest.B16TargetCard(bolt)}); err != nil {
		t.Fatalf("cast Fall of the Gavel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d, want %d", me.Life, life+5)
	}
	if !me.Graveyard.Contains(bolt) {
		t.Error("the targeted spell was not countered")
	}
	_ = opp
}

func TestOldB424VoidwielderMayBounceACreature(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
		effectstest.CastHoldCreature(t, g, "Voidwielder", "Creature — Human Wizard", oldB424VoidwielderOracle, 1, 4)
		effectstest.B04WaitForPick(t, g, me.ID)
		effectstest.AnswerPickTarget(t, g, theirs)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(theirs) == yes {
			t.Errorf("answered %v but their creature on the battlefield is %v", yes, g.Battlefield.Contains(theirs))
		}
		if yes && !opp.Hand.Contains(theirs) {
			t.Error("the bounced creature is not in its owner's hand")
		}
	}
}

func TestOldB424DramaticRescueBouncesAndGainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB424Seats(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Dramatic Rescue", "Instant", oldB424RescueOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the creature did not return to its owner's hand")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestOldB424SkymarkRocBouncesASmallCreatureDefendingPlayerControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB424Seats(g)
	roc := effectstest.B12Push(g, me.ID, "Skymark Roc", "Creature — Bird", oldB424RocOracle, 3, 3)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(roc, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, big) || effectstest.HasID(pick.PickTargetCards, roc) {
		t.Error("a creature with toughness over 2, or the Roc, was offered")
	}
	effectstest.AnswerPickTarget(t, g, small)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || !opp.Hand.Contains(small) {
		t.Error("the small creature was not returned to hand")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the big creature left")
	}
}
