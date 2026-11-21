package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eoe_mp_cards_test.go — Edge of Eternities cards whose names start with M
// to P (#4461). Each test plays the real catalog entry and asserts what its
// text says, and what it refuses.

const (
	eoeMPAssemblerOracle    = "2c2ef1ed-2f93-48c5-bf86-e46318599aeb"
	eoeMPNavigatorOracle    = "fa72c607-7eb4-41b7-9af1-f3d6c69abf09"
	eoeMPMechanozoaOracle   = "1aeb43f7-4d8a-4648-9c13-fd101081b448"
	eoeMPMoxiteOracle       = "86e0e718-82b5-4c99-9837-1b4d8729a27e"
	eoeMPEulogistOracle     = "046b60da-0a14-40b0-a36c-5328f6b8972b"
	eoeMPGearOracle         = "1d6742a3-592b-4e67-8cc6-56e2962ad7c7"
	eoeMPTeamLeaderOracle   = "c71d3775-0c71-4e37-979d-35e319f357f0"
	eoeMPModulationOracle   = "448d9391-9d15-4aa1-87b3-a1d755666c03"
	eoeMPMmmenonOracle      = "6c05b882-51d9-43d8-93b8-c3317590f76a"
	eoeMPModifierOracle     = "d0e58b8e-2c41-4602-a905-c2a6086ecf70"
	eoeMPSentryOracle       = "0b4cb111-ecf7-4ff7-a9e3-c944054f3370"
	eoeMPNebulaDragonOracle = "724b4a59-9c69-4f2b-a436-72183c62f1ff"
	eoeMPHellkiteOracle     = "9635d79e-62ec-4d80-ae37-2793489cc13b"
	eoeMPKillShipOracle     = "ad556aee-3dbf-4c8b-9f3a-31947e26c6f5"
)

func TestEoeMPEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eoeMPAssemblerOracle:    "Mechan Assembler",
		eoeMPNavigatorOracle:    "Mechan Navigator",
		eoeMPMechanozoaOracle:   "Mechanozoa",
		eoeMPMoxiteOracle:       "Melded Moxite",
		eoeMPEulogistOracle:     "Meltstrider Eulogist",
		eoeMPGearOracle:         "Meltstrider's Gear",
		eoeMPTeamLeaderOracle:   "Memorial Team Leader",
		eoeMPModulationOracle:   "Mental Modulation",
		eoeMPMmmenonOracle:      "Mm'menon, Uthros Exile",
		eoeMPModifierOracle:     "Molecular Modifier",
		eoeMPSentryOracle:       "Monoist Sentry",
		eoeMPNebulaDragonOracle: "Nebula Dragon",
		eoeMPHellkiteOracle:     "Nova Hellkite",
		eoeMPKillShipOracle:     "Pinnacle Kill-Ship",
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
	for oracle, cost := range map[string]string{
		eoeMPMechanozoaOracle: "{2}{U}",
		eoeMPTeamLeaderOracle: "{1}{R}",
		eoeMPHellkiteOracle:   "{2}{R}",
	} {
		if ac := game.AlternativeCostByKey(oracle, "warp"); ac == nil || ac.ManaCost != cost {
			t.Errorf("%s: warp %+v, want %s", oracle, ac, cost)
		}
	}
}

// --- helpers ---------------------------------------------------------------

// eoeMPArtifactEnters has a noncreature artifact token enter under owner's
// control and settles what it triggers, up to any target prompt.
func eoeMPArtifactEnters(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	effectstest.AvB1OpponentCreatureEnters(t, g, owner, "Artifact — Clue")
}

// eoeMPDiscardOne answers player's open discard prompt with the first card
// it offers.
func eoeMPDiscardOne(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	c := effectstest.DiscardChoiceFor(g, player)
	if c == nil || len(c.ChooseCards) == 0 {
		t.Fatalf("no discard prompt with a card to discard: %+v", c)
	}
	if err := g.ResolveChooseCards(c.ID, player, c.ChooseCards[:1]); err != nil {
		t.Fatalf("discard: %v", err)
	}
}

// --- Mechan Assembler --------------------------------------------------------

func TestEoeMPMechanAssemblerMakesOneRobotEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Mechan Assembler", "Artifact Creature — Robot Artificer", eoeMPAssemblerOracle, 4, 4)

	eoeMPArtifactEnters(t, g, opp.ID)
	if got := effectstest.EoeCDCount(g, me.ID, "Robot"); got != 0 {
		t.Fatalf("an opponent's artifact made %d Robots, want 0", got)
	}
	eoeMPArtifactEnters(t, g, me.ID)
	if got := effectstest.EoeCDCount(g, me.ID, "Robot"); got != 1 {
		t.Fatalf("your artifact made %d Robots, want 1", got)
	}
	robot := effectstest.AvB1Named(g, me.ID, "Robot")
	if c := effectstest.LayeredCard(t, g, robot); !c.IsArtifact() || !c.IsCreature() || c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("the token is a 2/2 artifact creature, got %d/%d", c.CurrentPower(), c.CurrentToughness())
	}
	eoeMPArtifactEnters(t, g, me.ID)
	if got := effectstest.EoeCDCount(g, me.ID, "Robot"); got != 1 {
		t.Errorf("a second artifact this turn made Robots too: %d, want 1", got)
	}
}

// --- Mechan Navigator --------------------------------------------------------

func TestEoeMPMechanNavigatorLootsWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nav := effectstest.B12Push(g, me.ID, "Mechan Navigator", "Artifact Creature — Robot Pilot", eoeMPNavigatorOracle, 2, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	hand := me.Hand.Size()
	if err := g.DeclareAttacker(nav, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d, want %d: the draw comes first", got, hand+1)
	}
	eoeMPDiscardOne(t, g, me.ID)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d after the discard", got, hand)
	}
}

// --- Mechanozoa --------------------------------------------------------------

func TestEoeMPMechanozoaTapsAndStunsAnOpponentsArtifactThenWarpsOut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	zoa := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Mechanozoa", TypeLine: "Artifact Creature — Robot Jellyfish", ManaCost: "{4}{U}{U}",
		OracleID: eoeMPMechanozoaOracle, Power: 5, Toughness: 5}, game.CastSpellParams{AlternativeCost: "warp"})
	effectstest.PassPriorityAroundTable(t, g)
	offered := eoeABOfferedCards(t, g)
	if !offered[rock] || offered[mine] {
		t.Fatalf("an opponent's artifact is offered and your creature is not: %v", offered)
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.AvB1Card(t, g, rock); !c.Tapped || c.Counters[game.CounterStun] != 1 {
		t.Errorf("the artifact is tapped with one stun counter: tapped %v, stun %d", c.Tapped, c.Counters[game.CounterStun])
	}
	effectstest.AdvanceThroughEndStep(t, g)
	if !g.Exile.Contains(zoa) {
		t.Error("the warped Mechanozoa is exiled at the end step")
	}
}

// --- Melded Moxite -----------------------------------------------------------

func TestEoeMPMeldedMoxiteRummagesThenSacrificesForATappedRobot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	moxite := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Melded Moxite", TypeLine: "Artifact", ManaCost: "{1}{R}",
		OracleID: eoeMPMoxiteOracle}, game.CastSpellParams{})
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil || c.ChooseMin != 0 || c.ChooseMax != 1 {
		t.Fatalf("an optional discard of one card: %+v", c)
	}
	eoeMPDiscardOne(t, g, me.ID)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d, want %d: one discarded, two drawn", got, hand+1)
	}

	effectstest.FillPoolColored(me, "C", 3)
	if err := effectstest.EoeCDActivate(g, me.ID, moxite, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(moxite) {
		t.Error("the Moxite was sacrificed")
	}
	robot := effectstest.AvB1Named(g, me.ID, "Robot")
	if robot == uuid.Nil {
		t.Fatal("no Robot token")
	}
	if c := effectstest.LayeredCard(t, g, robot); !c.Tapped || !c.IsArtifact() || c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("a tapped 2/2 Robot artifact: tapped %v, %d/%d", c.Tapped, c.CurrentPower(), c.CurrentToughness())
	}
}

func TestEoeMPMeldedMoxiteDrawsNothingWithoutADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AvB1Cast(t, g, me, game.Card{Name: "Melded Moxite", TypeLine: "Artifact", ManaCost: "{1}{R}",
		OracleID: eoeMPMoxiteOracle}, game.CastSpellParams{})
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d: no discard, no draw", got, hand)
	}
}

// --- Meltstrider Eulogist ----------------------------------------------------

func TestEoeMPMeltstriderEulogistDrawsForYourCreaturesWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	eulogist := effectstest.B12Push(g, me.ID, "Meltstrider Eulogist", "Creature — Insect Soldier", eoeMPEulogistOracle, 3, 3)
	grown := effectstest.B12Creature(g, me.ID, "Grown Bear", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	shrunk := effectstest.B12Creature(g, me.ID, "Shrunk Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for id, kind := range map[uuid.UUID]string{grown: game.CounterPlusOne, theirs: game.CounterPlusOne, shrunk: game.CounterMinusOne, eulogist: game.CounterPlusOne} {
		if err := g.AddCounter(id, kind, 1); err != nil {
			t.Fatalf("AddCounter: %v", err)
		}
	}
	for _, step := range []struct {
		id   uuid.UUID
		draw int
		what string
	}{
		{plain, 0, "your creature with no counter"},
		{shrunk, 0, "your creature with only a -1/-1 counter"},
		{theirs, 0, "an opponent's creature with a +1/+1 counter"},
		{grown, 1, "your creature with a +1/+1 counter"},
		{eulogist, 1, "the Eulogist itself with a +1/+1 counter"},
	} {
		hand := me.Hand.Size()
		effectstest.B15Destroy(t, g, step.id)
		if got := me.Hand.Size() - hand; got != step.draw {
			t.Errorf("%s died: drew %d, want %d", step.what, got, step.draw)
		}
	}
}

// --- Meltstrider's Gear ------------------------------------------------------

func TestEoeMPMeltstridersGearAttachesOnEntryAndEquipsForFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	gear := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Meltstrider's Gear", TypeLine: "Artifact — Equipment", ManaCost: "{G}",
		OracleID: eoeMPGearOracle}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EoeABPickIfAsked(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AvB1Card(t, g, gear).IsAttachedTo(bear) {
		t.Fatal("the Gear is attached to your creature as it enters")
	}
	if p, tough := effectstest.EoeCDPT(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("equipped creature is %d/%d, want 4/3", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "reach") {
		t.Error("equipped creature has reach")
	}

	other := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	effectstest.FillPoolColored(me, "C", 4)
	if err := effectstest.EoeCDActivate(g, me.ID, gear, 0, effectstest.EoeCDCard(other)); err == nil {
		t.Fatal("equip was paid with four mana")
	}
	effectstest.FillPoolColored(me, "C", 1)
	if err := effectstest.EoeCDActivate(g, me.ID, gear, 0, effectstest.EoeCDCard(other)); err != nil {
		t.Fatalf("equip {5}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AvB1Card(t, g, gear).IsAttachedTo(other) {
		t.Error("equip moves the Gear")
	}
	if p, tough := effectstest.EoeCDPT(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the old host is back to %d/%d, want 2/2", p, tough)
	}
}

// --- Memorial Team Leader ----------------------------------------------------

func TestEoeMPMemorialTeamLeaderPumpsTheOthersOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leader := effectstest.B12Push(g, me.ID, "Memorial Team Leader", "Creature — Kavu Soldier", eoeMPTeamLeaderOracle, 4, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p, tough := effectstest.EoeCDPT(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("during your turn your other creature is %d/%d, want 3/2", p, tough)
	}
	if p, _ := effectstest.EoeCDPT(t, g, leader); p != 4 {
		t.Errorf("the Leader itself has power %d, want 4", p)
	}
	if p, _ := effectstest.EoeCDPT(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
	effectstest.EndTurn(t, g)
	if g.Turn.ActiveSeat == 0 {
		t.Fatal("still your turn")
	}
	if p, _ := effectstest.EoeCDPT(t, g, bear); p != 2 {
		t.Errorf("during an opponent's turn your creature has power %d, want 2", p)
	}
}

// --- Mental Modulation -------------------------------------------------------

func TestEoeMPMentalModulationTapsDrawsAndCostsLessOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	card := game.Card{Name: "Mental Modulation", TypeLine: "Instant", ManaCost: "{1}{U}", OracleID: eoeMPModulationOracle}
	if err := effectstest.SocB19CastWithPool(g, me, card, "U", 1, game.CastSpellParams{Targets: effectstest.CardRefs(ogre)}); err != nil {
		t.Fatalf("cast for {U} during your turn: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AvB1Card(t, g, ogre).Tapped {
		t.Error("the target is tapped")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}

	effectstest.EndTurn(t, g)
	if g.Turn.ActiveSeat == 0 {
		t.Fatal("still your turn")
	}
	card.InstanceID = uuid.New()
	if err := effectstest.SocB19CastWithPool(g, me, card, "U", 1, game.CastSpellParams{Targets: effectstest.CardRefs(ogre)}); err == nil {
		t.Fatal("cast for {U} during an opponent's turn")
	}
	effectstest.FillPoolColored(me, "C", 1)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(ogre)}); err != nil {
		t.Fatalf("cast for {1}{U} during an opponent's turn: %v", err)
	}
}

// --- Mm'menon, Uthros Exile --------------------------------------------------

func TestEoeMPMmmenonPutsACounterOnTargetCreatureForEachArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mm := effectstest.B12Push(g, me.ID, "Mm'menon, Uthros Exile", "Legendary Creature — Jellyfish Advisor", eoeMPMmmenonOracle, 1, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, mm, "flying") {
		t.Error("printed flying is missing")
	}
	eoeMPArtifactEnters(t, g, opp.ID)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("an opponent's artifact triggered it")
	}
	eoeMPArtifactEnters(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("any creature can be the target: %d counters, want 1", got)
	}
	eoeMPArtifactEnters(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, mm)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, mm, game.CounterPlusOne); got != 1 {
		t.Errorf("a second artifact triggers it again: %d counters, want 1", got)
	}
}

// --- Molecular Modifier ------------------------------------------------------

func TestEoeMPMolecularModifierGivesFirstStrikeAtCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Molecular Modifier", "Creature — Kavu Artificer", eoeMPModifierOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, effectstest.EoeCDCard(theirs)); err == nil {
		t.Fatal("an opponent's creature was accepted")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if pw, tough := effectstest.EoeCDPT(t, g, bear); pw != 3 || tough != 2 {
		t.Errorf("the target is %d/%d, want 3/2", pw, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the target gains first strike")
	}
	effectstest.EndTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("first strike outlived the turn")
	}
}

// --- Monoist Sentry ----------------------------------------------------------

func TestEoeMPMonoistSentryCantAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentry := effectstest.B12Push(g, me.ID, "Monoist Sentry", "Artifact Creature — Robot", eoeMPSentryOracle, 4, 1)
	if !effectstest.HasKeywordOnBattlefield(t, g, sentry, "defender") {
		t.Error("printed defender is missing")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sentry, opp.ID); err == nil {
		t.Error("a creature with defender attacked")
	}
}

// --- Nebula Dragon -----------------------------------------------------------

func TestEoeMPNebulaDragonDealsThreeOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	dragon := effectstest.AvB1ResolveCreature(t, g, me, "Nebula Dragon", "Creature — Dragon", eoeMPNebulaDragonOracle)
	if !effectstest.HasKeywordOnBattlefield(t, g, dragon, "flying") {
		t.Error("printed flying is missing")
	}
	effectstest.PickCard(t, g, me.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(ogre) || !opp.Graveyard.Contains(ogre) {
		t.Error("3 damage kills a 3/3")
	}
}

// --- Nova Hellkite -----------------------------------------------------------

func TestEoeMPNovaHellkitePingsAnOpponentsCreatureAndWarpsOut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	elf := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	kite := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Nova Hellkite", TypeLine: "Creature — Dragon", ManaCost: "{3}{R}{R}",
		OracleID: eoeMPHellkiteOracle, Power: 4, Toughness: 5}, game.CastSpellParams{AlternativeCost: "warp"})
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, kite, kw) {
			t.Errorf("printed %s is missing", kw)
		}
	}
	offered := eoeABOfferedCards(t, g)
	if !offered[elf] || offered[mine] || offered[kite] {
		t.Fatalf("only a creature an opponent controls is offered: %v", offered)
	}
	effectstest.PickCard(t, g, me.ID, elf)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(elf) {
		t.Error("1 damage kills a 1/1")
	}
	effectstest.AdvanceThroughEndStep(t, g)
	if !g.Exile.Contains(kite) {
		t.Error("the warped Hellkite is exiled at the end step")
	}
}

// --- Pinnacle Kill-Ship ------------------------------------------------------

func TestEoeMPPinnacleKillShipDealsTenAndFliesAtSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	titan := effectstest.B12Creature(g, opp.ID, "Their Titan", "Creature — Giant", 9, 9)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ship := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Pinnacle Kill-Ship", TypeLine: "Artifact — Spacecraft", ManaCost: "{7}",
		OracleID: eoeMPKillShipOracle, Power: 7, Toughness: 7}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, titan)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(titan) {
		t.Error("10 damage kills a 9/9")
	}
	if effectstest.LayeredCard(t, g, ship).IsCreature() {
		t.Error("below 7 charge counters it is not a creature")
	}
	if err := g.AddCounter(ship, game.CounterCharge, 7); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	at7 := effectstest.LayeredCard(t, g, ship)
	if !at7.IsCreature() || !game.HasKeyword(&at7, "flying") || at7.CurrentPower() != 7 || at7.CurrentToughness() != 7 {
		t.Errorf("at 7+ it is a 7/7 flying creature, got %d/%d", at7.CurrentPower(), at7.CurrentToughness())
	}
}

func TestEoeMPPinnacleKillShipMayEnterWithNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AvB1Cast(t, g, me, game.Card{Name: "Pinnacle Kill-Ship", TypeLine: "Artifact — Spacecraft", ManaCost: "{7}",
		OracleID: eoeMPKillShipOracle, Power: 7, Toughness: 7}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("choosing no target: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(bear) {
		t.Error("with no target chosen nothing is dealt damage")
	}
}
