package testsru

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b22_cards_test.go — Ravnica: Clue Edition, S82 BNS batch 22
// (#5929): Rakdos, the Showstopper; Senator Peacock; Trostani Discordant;
// Clan Defiance; Corporeal Projection; Covetous Elegy; Direct Current;
// Firespout; Infest; Nissa's Judgment; Overcome; Incubation // Incongruity.

const (
	bnsRakdosOracle       = "03a4b997-4738-41a9-933f-e1f765e3a75a"
	bnsPeacockOracle      = "10d8f29a-d483-4f06-bc13-d41d62847ffe"
	bnsTrostaniOracle     = "9a2fdc0e-e0de-47fc-95e4-04e725ddf060"
	bnsClanDefianceOracle = "abcf5e77-88fc-4b76-a1e8-907a5b27d612"
	bnsCorporealOracle    = "09038989-3c1f-4426-a150-e0a9374f34c3"
	bnsElegyOracle        = "dd69ef80-d30c-4b15-aec8-4154dc20cb25"
	bnsDirectCurrent      = "54e34440-c890-4a07-83c6-63a2f7a65639"
	bnsFirespoutOracle    = "eb2524ae-899a-4f91-9feb-0716cf689db3"
	bnsInfestOracle       = "d6850616-7db5-4141-9ab0-ae8d1f08114f"
	bnsNissasJudgment     = "14247820-5ecb-40bd-8dee-0625cc745ec2"
	bnsOvercomeOracle     = "15353e27-96ca-48c7-83d7-efd57c094643"
	bnsIncubationOracle   = "363edb16-009f-4cc2-99e3-2b15db84e7af"
)

func TestBnsB22CardsAreRegisteredFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsRakdosOracle:            "Rakdos, the Showstopper",
		bnsPeacockOracle:           "Senator Peacock",
		bnsTrostaniOracle:          "Trostani Discordant",
		bnsClanDefianceOracle:      "Clan Defiance",
		bnsCorporealOracle:         "Corporeal Projection",
		bnsElegyOracle:             "Covetous Elegy",
		bnsDirectCurrent:           "Direct Current",
		bnsFirespoutOracle:         "Firespout",
		bnsInfestOracle:            "Infest",
		bnsNissasJudgment:          "Nissa's Judgment",
		bnsOvercomeOracle:          "Overcome",
		bnsIncubationOracle:        "Incubation",
		bnsIncubationOracle + "#1": "Incongruity",
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

// --- Infest / Overcome ---------------------------------------------------------

func TestBnsB22InfestShrinksEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	dmuB09Cast(t, g, "Infest", "Sorcery", bnsInfestOracle, false, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("my creature is %d/%d, want 2/2", p, tough)
	}
	if effectstest.FindBattlefieldCardByID(g, theirs) != nil {
		t.Error("their 2/2 survived the -2/-2")
	}
	later := effectstest.B12Creature(g, me.ID, "Later", "Creature — Bear", 3, 3)
	if p, tough := effectstest.PtOf(t, g, later); p != 3 || tough != 3 {
		t.Errorf("a creature that arrived afterwards is %d/%d, want 3/3", p, tough)
	}
}

func TestBnsB22OvercomePumpsOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	dmuB09Cast(t, g, "Overcome", "Sorcery", bnsOvercomeOracle, false, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("my creature is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, mine, "trample") {
		t.Error("my creature did not gain trample")
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their creature is %d/%d, want 2/2", p, tough)
	}
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "trample") {
		t.Error("their creature gained trample")
	}
}

// --- Direct Current -------------------------------------------------------------

func TestBnsB22DirectCurrentDealsTwoAndJumpStarts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dmuB09Cast(t, g, "Direct Current", "Sorcery", bnsDirectCurrent, false, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != 38 {
		t.Fatalf("life %d, want 38", got)
	}
	id := effectstest.SeedGraveyardCard(t, g, "Direct Current", "Sorcery", bnsDirectCurrent)
	fodder := effectstest.HandCard(me, "Fodder", "Sorcery")
	effectstest.S39Fund(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: "jump-start", Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err == nil {
		t.Fatal("jump-start was castable without discarding a card")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: "jump-start", AltCostIDs: []uuid.UUID{fodder}, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("jump-start cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != 36 {
		t.Errorf("life %d after jump-start, want 36", got)
	}
	if !g.Exile.Contains(id) {
		t.Error("the jump-started card is not exiled")
	}
}

func TestBnsB22DirectCurrentHitsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 5, 5)
	dmuB09Cast(t, g, "Direct Current", "Sorcery", bnsDirectCurrent, false, effectstest.CardTarget(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, theirs); got != 2 {
		t.Errorf("damage %d, want 2", got)
	}
}

// --- Firespout ------------------------------------------------------------------

func bnsB22FlyingCreature(g *game.Game, owner uuid.UUID, name string, p, tough int) uuid.UUID {
	return effectstest.PushTribalCreature(g, owner, name, "Creature — Bird", p, tough, "flying")
}

func bnsB22Firespout(t *testing.T, colors ...string) (g *game.Game, ground, air uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ground = effectstest.B12Creature(g, opp.ID, "Walker", "Creature — Bear", 4, 5)
	air = bnsB22FlyingCreature(g, opp.ID, "Flapper", 4, 5)
	effectstest.B06AddMana(me, colors...)
	if err := dmuCastPaid(g, me, "Firespout", "Sorcery", "{2}{R/G}", bnsFirespoutOracle, false, nil); err != nil {
		t.Fatalf("cast Firespout with %v: %v", colors, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return g, ground, air
}

func TestBnsB22FirespoutRedHitsOnlyCreaturesWithoutFlying(t *testing.T) {
	g, ground, air := bnsB22Firespout(t, "R", "R", "R")
	if got := effectstest.DamageOn(t, g, ground); got != 3 {
		t.Errorf("ground damage %d, want 3", got)
	}
	if got := effectstest.DamageOn(t, g, air); got != 0 {
		t.Errorf("flyer damage %d, want 0", got)
	}
}

func TestBnsB22FirespoutGreenHitsOnlyCreaturesWithFlying(t *testing.T) {
	g, ground, air := bnsB22Firespout(t, "G", "G", "G")
	if got := effectstest.DamageOn(t, g, ground); got != 0 {
		t.Errorf("ground damage %d, want 0", got)
	}
	if got := effectstest.DamageOn(t, g, air); got != 3 {
		t.Errorf("flyer damage %d, want 3", got)
	}
}

func TestBnsB22FirespoutBothColorsHitsEverything(t *testing.T) {
	g, ground, air := bnsB22Firespout(t, "R", "G", "G")
	if got := effectstest.DamageOn(t, g, ground); got != 3 {
		t.Errorf("ground damage %d, want 3", got)
	}
	if got := effectstest.DamageOn(t, g, air); got != 3 {
		t.Errorf("flyer damage %d, want 3", got)
	}
}

func TestBnsB22FirespoutPaidWithColorlessDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := dmuCastPaid(g, me, "Firespout", "Sorcery", "{2}{R/G}", bnsFirespoutOracle, false, nil); err == nil {
		t.Error("a hybrid {R/G} was payable with colorless mana")
	}
}

// --- Clan Defiance --------------------------------------------------------------

func TestBnsB22ClanDefianceDealsXToEachChosenTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ground := effectstest.B12Creature(g, opp.ID, "Walker", "Creature — Bear", 4, 5)
	air := bnsB22FlyingCreature(g, opp.ID, "Flapper", 4, 5)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardForTest(me, "Clan Defiance", "Sorcery", bnsClanDefianceOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, Modes: []int{0, 1, 2}, Targets: []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, air, 0, 0), effectstest.ModeRef(game.TargetCard, ground, 1, 0), effectstest.ModeRef(game.TargetPlayer, opp.ID, 2, 0),
	}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, air); got != 2 {
		t.Errorf("flyer damage %d, want 2", got)
	}
	if got := effectstest.DamageOn(t, g, ground); got != 2 {
		t.Errorf("ground damage %d, want 2", got)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != 38 {
		t.Errorf("life %d, want 38", got)
	}
}

func TestBnsB22ClanDefianceRefusesAFlyerForTheWithoutFlyingMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	air := bnsB22FlyingCreature(g, opp.ID, "Flapper", 4, 5)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardForTest(me, "Clan Defiance", "Sorcery", bnsClanDefianceOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 1, Modes: []int{1}, Targets: []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, air, 0, 0),
	}}); err == nil {
		t.Error("a creature with flying was accepted for the without-flying mode")
	}
}

func TestBnsB22ClanDefianceOneModeOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardForTest(me, "Clan Defiance", "Sorcery", bnsClanDefianceOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3, Modes: []int{2}, Targets: []game.TargetRef{
		effectstest.ModeRef(game.TargetPlayer, opp.ID, 0, 0),
	}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != 37 {
		t.Errorf("life %d, want 37", got)
	}
}

// --- Nissa's Judgment -----------------------------------------------------------

func TestBnsB22NissasJudgmentSupportsThenEveryCounteredCreatureBites(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 3, 3)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 6, 6)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 1, 9)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardForTest(me, "Nissa's Judgment", "Sorcery", bnsNissasJudgment)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: b, Slot: 0}, {Kind: game.TargetCard, ID: victim, Slot: 1},
	}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCountersOf(g, a, game.CounterPlusOne); got != 1 {
		t.Errorf("A has %d counters, want 1", got)
	}
	// A is 3/3 and B is 4/4 after their counters; the uncountered 6/6 deals nothing.
	if got := effectstest.DamageOn(t, g, victim); got != 7 {
		t.Errorf("victim took %d, want 7 (3 + 4)", got)
	}
	_ = plain
}

func TestBnsB22NissasJudgmentWithNoVictimStillSupports(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardForTest(me, "Nissa's Judgment", "Sorcery", bnsNissasJudgment)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a, Slot: 0}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCountersOf(g, a, game.CounterPlusOne); got != 1 {
		t.Errorf("A has %d counters, want 1", got)
	}
}

func TestBnsB22NissasJudgmentRefusesAnOwnCreatureAsTheVictim(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardForTest(me, "Nissa's Judgment", "Sorcery", bnsNissasJudgment)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine, Slot: 1}}}); err == nil {
		t.Error("a creature you control was accepted as the victim")
	}
}

// --- Corporeal Projection -------------------------------------------------------

func TestBnsB22CorporealProjectionGrantsMyriadToTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushTribalCreature(g, me.ID, "A", "Creature — Bear", 2, 2)
	other := effectstest.PushTribalCreature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Corporeal Projection", "Sorcery", bnsCorporealOracle, effectstest.CardTarget(a))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, a, other)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if n := len(effectstest.S46Tokens(g, me.ID, "A")); n != 2 {
		t.Errorf("%d copies of the target, want 2 (one per other opponent)", n)
	}
	if n := len(effectstest.S46Tokens(g, me.ID, "Other")); n != 0 {
		t.Errorf("the other creature got myriad: %d copies", n)
	}
}

func TestBnsB22CorporealProjectionOverloadGrantsEveryCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushTribalCreature(g, me.ID, "A", "Creature — Bear", 2, 2)
	theirs := effectstest.PushTribalCreature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardForTest(me, "Corporeal Projection", "Sorcery", bnsCorporealOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "overload"}); err != nil {
		t.Fatalf("overload cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if n := len(effectstest.S46Tokens(g, me.ID, "A")); n != 2 {
		t.Errorf("%d copies of my creature, want 2", n)
	}
	_ = theirs
}

// --- Covetous Elegy -------------------------------------------------------------

func TestBnsB22CovetousElegyKeepsTwoEachAndMakesTappedTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := []uuid.UUID{
		effectstest.B12Creature(g, me.ID, "M1", "Creature — Bear", 2, 2), effectstest.B12Creature(g, me.ID, "M2", "Creature — Bear", 2, 2),
		effectstest.B12Creature(g, me.ID, "M3", "Creature — Bear", 2, 2),
	}
	theirs := []uuid.UUID{
		effectstest.B12Creature(g, opp.ID, "T1", "Creature — Bear", 2, 2), effectstest.B12Creature(g, opp.ID, "T2", "Creature — Bear", 2, 2),
		effectstest.B12Creature(g, opp.ID, "T3", "Creature — Bear", 2, 2), effectstest.B12Creature(g, opp.ID, "T4", "Creature — Bear", 2, 2),
	}
	dmuB09Cast(t, g, "Covetous Elegy", "Sorcery", bnsElegyOracle, false, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, mine[0], mine[1])
	effectstest.AnswerChooseCards(t, g, opp.ID, theirs[2], theirs[3])
	if !g.Battlefield.Contains(mine[0]) || !g.Battlefield.Contains(mine[1]) || g.Battlefield.Contains(mine[2]) {
		t.Error("my creatures: the two kept should stay and the third be sacrificed")
	}
	if g.Battlefield.Contains(theirs[0]) || g.Battlefield.Contains(theirs[1]) || !g.Battlefield.Contains(theirs[2]) || !g.Battlefield.Contains(theirs[3]) {
		t.Error("their creatures: the two kept should stay and the others be sacrificed")
	}
	// Two survivors for the opponent that chose; the other two opponents have none.
	treasures := effectstest.S46Tokens(g, me.ID, "Treasure")
	if len(treasures) != 2 {
		t.Fatalf("%d Treasures, want 2 (one per surviving opposing creature)", len(treasures))
	}
	for _, c := range treasures {
		if !c.Tapped {
			t.Error("a Treasure entered untapped")
		}
	}
}

func TestBnsB22CovetousElegyAsksNobodyWithTwoOrFewerCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "M1", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "T1", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "T2", "Creature — Bear", 2, 2)
	dmuB09Cast(t, g, "Covetous Elegy", "Sorcery", bnsElegyOracle, false, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, opp.ID) != nil || effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a player with two or fewer creatures was asked to choose")
	}
	if got := len(effectstest.S46Tokens(g, me.ID, "Treasure")); got != 2 {
		t.Errorf("%d Treasures, want 2", got)
	}
}

// --- Rakdos, the Showstopper ----------------------------------------------------

func TestBnsB22RakdosSparesDemonsDevilsAndImpsAndFlipsForTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	demon := effectstest.B12Creature(g, opp.ID, "A Demon", "Creature — Demon", 1, 1)
	devil := effectstest.B12Creature(g, opp.ID, "A Devil", "Creature — Devil", 1, 1)
	imp := effectstest.B12Creature(g, me.ID, "An Imp", "Creature — Imp", 1, 1)
	const n = 40
	var bears []uuid.UUID
	for i := 0; i < n; i++ {
		bears = append(bears, effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2))
	}
	effectstest.CastCatalogSpell(t, g, "Rakdos, the Showstopper", "Legendary Creature — Demon", bnsRakdosOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{demon, devil, imp} {
		if !g.Battlefield.Contains(id) {
			t.Error("a Demon, Devil or Imp was destroyed")
		}
	}
	survivors := 0
	for _, id := range bears {
		if g.Battlefield.Contains(id) {
			survivors++
		}
	}
	if survivors == 0 || survivors == n {
		t.Errorf("%d of %d bears survived; a coin per creature should split them", survivors, n)
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Rakdos, the Showstopper"); got != 1 {
		t.Errorf("Rakdos on the battlefield: %d, want 1 (it is a Demon)", got)
	}
}

// --- Trostani Discordant --------------------------------------------------------

func TestBnsB22TrostaniAnthemsMakesLifelinkSoldiersAndReturnsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	stolen := effectstest.DcSeed(g, opp.ID, "Stolen Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { g.GainControlForEffect(uuid.Nil, stolen, me.ID, game.IndefiniteDuration(), "test theft") })
	effectstest.CastCatalogSpell(t, g, "Trostani Discordant", "Legendary Creature — Dryad", bnsTrostaniOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("my bear is %d/%d, want 3/3", p, tough)
	}
	soldiers := effectstest.S46Tokens(g, me.ID, "Soldier")
	if len(soldiers) != 2 {
		t.Fatalf("%d Soldier tokens, want 2", len(soldiers))
	}
	for _, s := range soldiers {
		if !effectstest.EffectiveAbilitiesContain(t, g, s.InstanceID, "lifelink") {
			t.Error("a Soldier lacks lifelink")
		}
		if p, tough := effectstest.PtOf(t, g, s.InstanceID); p != 2 || tough != 2 {
			t.Errorf("a Soldier is %d/%d, want 2/2 with the anthem", p, tough)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, stolen); got != opp.ID {
		t.Error("the stolen creature was not returned to its owner at the end step")
	}
	if got := effectstest.ControllerOf(t, g, bear); got != me.ID {
		t.Error("my own creature changed hands")
	}
}

// --- Senator Peacock ------------------------------------------------------------

func TestBnsB22PeacockMakesArtifactsCluesThatSacrificeToDrawAndEvadeBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Senator Peacock", "Legendary Creature — Human Advisor", bnsPeacockOracle, 3, 4)
	mine := effectstest.B12Permanent(g, me.ID, "Trinket", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Trinket", "Artifact")
	runner := effectstest.B12Creature(g, me.ID, "Runner", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, mine); c == nil || !slices.Contains(c.Effective().Subtypes, "Clue") {
		t.Fatal("my artifact is not a Clue")
	}
	if c := effectstest.FindBattlefieldCardByID(g, theirs); c == nil || slices.Contains(c.Effective().Subtypes, "Clue") {
		t.Fatal("an opposing artifact became a Clue")
	}
	hand := me.Hand.Size()
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, mine, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the granted Clue ability: %v", err)
	}
	effectstest.AnswerPickTarget(t, g, runner)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("the Clue was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.DeclareAttack(t, g, opp.ID, runner)
	if reason := effectstest.BlockRefusal(t, g, wall, runner); reason == "" {
		t.Error("the targeted creature could be blocked after a Clue was sacrificed")
	}
}

// --- Incubation // Incongruity --------------------------------------------------

func bnsB22Split(g *game.Game, me *game.Player) uuid.UUID {
	card := game.Card{
		InstanceID: uuid.New(), OracleID: bnsIncubationOracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Incubation", TypeLine: "Sorcery", ManaCost: "{G/U}", Colors: []string{"G", "U"},
				OracleText: "Look at the top five cards of your library. You may reveal a creature card from among them and put it into your hand. Put the rest on the bottom of your library in a random order."},
			{Name: "Incongruity", TypeLine: "Instant", ManaCost: "{1}{G}{U}", Colors: []string{"G", "U"},
				OracleText: "Exile target creature. That creature's controller creates a 3/3 green Frog Lizard creature token."},
		},
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	return card.InstanceID
}

func TestBnsB22IncubationTakesACreatureAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var top [5]uuid.UUID
	cards := []struct{ Name, TypeLine string }{
		{"Forest", "Basic Land — Forest"}, {"Bear", "Creature — Bear"}, {"Shock", "Instant"},
		{"Golem", "Artifact Creature — Golem"}, {"Rancor", "Enchantment — Aura"},
	}
	for i := len(cards) - 1; i >= 0; i-- {
		top[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: top[i], Name: cards[i].Name, TypeLine: cards[i].TypeLine, Owner: me.ID, Controller: me.ID})
	}
	effectstest.AdvanceToMain(t, g)
	id := bnsB22Split(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 0}); err != nil {
		t.Fatalf("cast Incubation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick opened")
	}
	offered := map[uuid.UUID]bool{}
	for _, c := range pick.ChooseCards {
		offered[c] = true
	}
	if len(offered) != 2 || !offered[top[1]] || !offered[top[3]] || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (min %d max %d), want the two creature cards, up to one", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[1]) {
		t.Fatal("the creature card is not in hand")
	}
	if effectstest.PutInLibraryChoiceFor(g, me.ID) != nil {
		t.Error("the rest go on the bottom in a RANDOM order: no order prompt is owed")
	}
	if !me.Library.Contains(top[0]) || !me.Library.Contains(top[2]) {
		t.Error("the rest left the library")
	}
}

func TestBnsB22IncongruityExilesAndGivesItsControllerAFrogLizard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 5, 5)
	effectstest.AdvanceToMain(t, g)
	id := bnsB22Split(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1, Targets: effectstest.CardTarget(theirs)}); err != nil {
		t.Fatalf("cast Incongruity: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !g.Exile.Contains(theirs) {
		t.Error("the creature was not exiled")
	}
	frogs := effectstest.S46Tokens(g, opp.ID, "Frog Lizard")
	if len(frogs) != 1 {
		t.Fatalf("%d Frog Lizards for its controller, want 1", len(frogs))
	}
	if p, tough := effectstest.PtOf(t, g, frogs[0].InstanceID); p != 3 || tough != 3 {
		t.Errorf("Frog Lizard is %d/%d, want 3/3", p, tough)
	}
	if len(effectstest.S46Tokens(g, me.ID, "Frog Lizard")) != 0 {
		t.Error("the caster got a Frog Lizard")
	}
}
