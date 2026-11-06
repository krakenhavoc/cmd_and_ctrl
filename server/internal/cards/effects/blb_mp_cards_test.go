package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blb_mp_cards_test.go — Bloomburrow cards M–P (#4467). Each test plays
// the real catalog entry and asserts what its text says.

const (
	blbMPMettleOracle      = "c657626f-7444-4c26-842c-dda9b9656fa3"
	blbMPMeekOracle        = "19b5c94b-b4fa-41a9-b649-51c4e6f03602"
	blbMPDrillOracle       = "3b10ab39-3050-475c-af8b-f293794d498c"
	blbMPSpringOracle      = "3428287b-cfd5-45bc-b5f9-1e3f5a425a68"
	blbMPClericOracle      = "d8067f42-68c0-42da-a1d6-859953939555"
	blbMPHungerOracle      = "ed7f64d1-69ba-40f5-9863-3edba10e23c7"
	blbMPOverprotectOracle = "c8c3a293-ae6d-4c52-be2d-2d4b6a8edb66"
	blbMPPearlOracle       = "91ddd2bd-2dec-4670-8edb-e89c009b2187"
	blbMPProvisionerOracle = "0534e20d-8352-4d9f-a617-378070602673"
	blbMPShoveOracle       = "2094fb68-3859-401f-97f2-c8bae64a4822"
	blbMPPolliwallopOracle = "efd2a4ca-1bd0-4762-a997-f099326b1746"
	blbMPProphetOracle     = "6eabebdf-c1a3-4789-a67f-859f0957c5df"
	blbMPWhorlOracle       = "3df4eb80-e8dd-4a9e-bd8a-b4b23fb9de6c"
)

func TestBlbMPEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		blbMPMettleOracle:      "Mabel's Mettle",
		blbMPMeekOracle:        "Might of the Meek",
		blbMPDrillOracle:       "Mind Drill Assailant",
		blbMPSpringOracle:      "Mind Spring",
		blbMPClericOracle:      "Moonrise Cleric",
		blbMPHungerOracle:      "Nocturnal Hunger",
		blbMPOverprotectOracle: "Overprotect",
		blbMPPearlOracle:       "Pearl of Wisdom",
		blbMPProvisionerOracle: "Pileated Provisioner",
		blbMPShoveOracle:       "Playful Shove",
		blbMPPolliwallopOracle: "Polliwallop",
		blbMPProphetOracle:     "Pond Prophet",
		blbMPWhorlOracle:       "Psychic Whorl",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// blbMPFillGraveyard puts n plain cards into p's graveyard, then nudges
// the battlefield so the layer cache is refreshed (a direct graveyard
// push is no game action).
func blbMPFillGraveyard(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n; i++ {
		pushGraveyardCardTyped(p, "Old Card", "Sorcery")
	}
	blbB1Push(g, p.ID, "Nudge", "Creature — Bear", "", 1, 1)
}

// --- Mabel's Mettle ----------------------------------------------------------

func TestBlbMPMabelsMettlePumpsBothTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	a := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	b := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Mabel's Mettle", "Instant", blbMPMettleOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: b, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, a); p != 4 || tough != 4 {
		t.Errorf("the first target is %d/%d, want 4/4", p, tough)
	}
	if p, tough := ptOf(t, g, b); p != 3 || tough != 3 {
		t.Errorf("the second target is %d/%d, want 3/3", p, tough)
	}
}

func TestBlbMPMabelsMettleSecondTargetIsOptionalAndOther(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	a := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	if err := castCatalogSpellErr(t, g, "Mabel's Mettle", "Instant", blbMPMettleOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: a, Slot: 1},
	}); err == nil {
		t.Fatal("the same creature was accepted twice; the second is \"other\"")
	}
	castCatalogSpell(t, g, "Mabel's Mettle", "Instant", blbMPMettleOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0},
	})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, a); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
}

// --- Might of the Meek -------------------------------------------------------

func TestBlbMPMightOfTheMeekWithoutAMouseGivesTrampleAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedLibrary(me, "Draw Me")
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Might of the Meek", "Instant", blbMPMeekOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("the target should have trample")
	}
	if p, _ := ptOf(t, g, bear); p != 2 {
		t.Errorf("no Mouse: the bear's power is %d, want 2", p)
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d (the spell left, a card was drawn)", got, before+1)
	}
}

func TestBlbMPMightOfTheMeekWithAMouseAlsoGivesPower(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedLibrary(me, "Draw Me")
	blbB1Push(g, me.ID, "Mouse", "Creature — Mouse Soldier", "", 1, 1)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Might of the Meek", "Instant", blbMPMeekOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("with a Mouse the bear is %d/%d, want 3/2", p, tough)
	}
}

// --- Mind Drill Assailant ----------------------------------------------------

func TestBlbMPMindDrillAssailantThreshold(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	drill := blbB1Push(g, me.ID, "Mind Drill Assailant", "Creature — Rat Warlock", blbMPDrillOracle, 2, 5)
	blbMPFillGraveyard(g, me, 6)
	if p, _ := ptOf(t, g, drill); p != 2 {
		t.Fatalf("six cards in the graveyard: power %d, want 2", p)
	}
	blbMPFillGraveyard(g, me, 1)
	if p, tough := ptOf(t, g, drill); p != 5 || tough != 5 {
		t.Errorf("seven cards in the graveyard: %d/%d, want 5/5", p, tough)
	}
}

func TestBlbMPMindDrillAssailantSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedLibrary(me, "Top")
	drill := blbB1Push(g, me.ID, "Mind Drill Assailant", "Creature — Rat Warlock", blbMPDrillOracle, 2, 5)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "U")
	b16Activate(t, g, me.ID, drill, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the ability queued no surveil")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("surveil looked at %d cards, want 1", len(c.ScryCards))
	}
}

// --- Mind Spring -------------------------------------------------------------

func TestBlbMPMindSpringDrawsX(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B", "C", "D")
	before := me.Hand.Size()
	castXSpell(t, g, "Mind Spring", "Sorcery", blbMPSpringOracle, "{X}{U}{U}", 3, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+3 {
		t.Errorf("hand %d, want %d (X = 3)", got, before+3)
	}
}

// --- Moonrise Cleric ---------------------------------------------------------

func TestBlbMPMoonriseClericGainsLifeWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	cleric := blbB1Push(g, me.ID, "Moonrise Cleric", "Creature — Bat Cleric", blbMPClericOracle, 2, 3)
	if !hasKeywordOnBattlefield(t, g, cleric, "flying") {
		t.Error("the Cleric should have flying")
	}
	life := lifeOf(g, me.ID)
	declareAttack(t, g, opp.ID, cleric)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != life+1 {
		t.Errorf("I am at %d, want %d", got, life+1)
	}
}

// --- Nocturnal Hunger --------------------------------------------------------

func TestBlbMPNocturnalHungerPromisedDestroysWithoutLifeLoss(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	victim := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	life := lifeOf(g, me.ID)
	mustCastWithGift(t, g, "Nocturnal Hunger", "Instant", blbMPHungerOracle, cardTarget(victim), opp.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature should be destroyed")
	}
	if n, _ := countTokensNamed(g, opp.ID, "Food"); n != 1 {
		t.Errorf("the opponent has %d Food tokens, want 1", n)
	}
	if got := lifeOf(g, me.ID); got != life {
		t.Errorf("the gift was promised: I am at %d, want %d", got, life)
	}
}

func TestBlbMPNocturnalHungerUnpromisedCostsTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	victim := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	life := lifeOf(g, me.ID)
	mustCastWithGift(t, g, "Nocturnal Hunger", "Instant", blbMPHungerOracle, cardTarget(victim), uuid.Nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature should be destroyed")
	}
	if n, _ := countTokensNamed(g, opp.ID, "Food"); n != 0 {
		t.Errorf("%d Food tokens without the gift, want 0", n)
	}
	if got := lifeOf(g, me.ID); got != life-2 {
		t.Errorf("no gift: I am at %d, want %d", got, life-2)
	}
}

// --- Overprotect -------------------------------------------------------------

func TestBlbMPOverprotectPumpsAndProtectsYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mine := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	if err := castCatalogSpellErr(t, g, "Overprotect", "Instant", blbMPOverprotectOracle, cardTarget(theirs)); err == nil {
		t.Fatal("an opponent's creature was accepted; the target is a creature you control")
	}
	castCatalogSpell(t, g, "Overprotect", "Instant", blbMPOverprotectOracle, cardTarget(mine))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, mine); p != 5 || tough != 5 {
		t.Errorf("the bear is %d/%d, want 5/5", p, tough)
	}
	for _, kw := range []string{"trample", "hexproof", "indestructible"} {
		if !hasKeywordOnBattlefield(t, g, mine, kw) {
			t.Errorf("the bear should have %s", kw)
		}
	}
}

// --- Pearl of Wisdom ---------------------------------------------------------

func TestBlbMPPearlOfWisdomCostsLessWithAnOtter(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	if got := selfPricedMV(t, g, me, blbMPPearlOracle, "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("without an Otter it costs %d, want 3", got)
	}
	blbB1Push(g, me.ID, "Otter", "Creature — Otter Wizard", "", 1, 1)
	if got := selfPricedMV(t, g, me, blbMPPearlOracle, "Sorcery", "{2}{U}"); got != 2 {
		t.Errorf("with an Otter it costs %d, want 2", got)
	}
}

func TestBlbMPPearlOfWisdomDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B")
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Pearl of Wisdom", "Sorcery", blbMPPearlOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand %d, want %d", got, before+2)
	}
}

// --- Pileated Provisioner ----------------------------------------------------

func TestBlbMPPileatedProvisionerCountersACreatureWithoutFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	bird := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID,
	})
	prov := castHoldCreature(t, g, "Pileated Provisioner", "Creature — Bird Scout", blbMPProvisionerOracle, 3, 4)
	ch := pickTargetPrompt(t, g)
	for _, bad := range []uuid.UUID{bird, prov} {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatal("a creature with flying was accepted; the target is a creature without flying")
		}
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 1 {
		t.Errorf("the bear has %d +1/+1 counters, want 1", got)
	}
}

// --- Playful Shove -----------------------------------------------------------

func TestBlbMPPlayfulShoveDealsOneAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	seedLibrary(me, "Draw Me")
	before, life := me.Hand.Size(), lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Playful Shove", "Sorcery", blbMPShoveOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-1 {
		t.Errorf("the opponent is at %d, want %d", got, life-1)
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}

// --- Polliwallop -------------------------------------------------------------

func TestBlbMPPolliwallopHasAffinityForFrogs(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	if got := selfPricedMV(t, g, me, blbMPPolliwallopOracle, "Instant", "{3}{G}"); got != 4 {
		t.Errorf("no Frogs: it costs %d, want 4", got)
	}
	blbB1Push(g, me.ID, "Frog", "Creature — Frog", "", 1, 1)
	blbB1Push(g, me.ID, "Frog", "Creature — Frog", "", 1, 1)
	if got := selfPricedMV(t, g, me, blbMPPolliwallopOracle, "Instant", "{3}{G}"); got != 2 {
		t.Errorf("two Frogs: it costs %d, want 2", got)
	}
	for i := 0; i < 3; i++ {
		blbB1Push(g, me.ID, "Frog", "Creature — Frog", "", 1, 1)
	}
	if got := selfPricedMV(t, g, me, blbMPPolliwallopOracle, "Instant", "{3}{G}"); got != 1 {
		t.Errorf("five Frogs: it costs %d, want 1 (only the generic part is reduced)", got)
	}
}

func TestBlbMPPolliwallopDealsTwiceThePower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mine := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 3, 3)
	other := blbB1Push(g, me.ID, "My Other", "Creature — Bear", "", 9, 9)
	theirs := blbB1Push(g, opp.ID, "Big", "Creature — Beast", "", 9, 9)
	if err := castCatalogSpellErr(t, g, "Polliwallop", "Instant", blbMPPolliwallopOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: other, Slot: 1},
	}); err == nil {
		t.Fatal("your own creature was accepted as the creature you don't control")
	}
	castCatalogSpell(t, g, "Polliwallop", "Instant", blbMPPolliwallopOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, theirs); got != 6 {
		t.Errorf("the creature has %d damage, want 6", got)
	}
	if got := damageOn(t, g, mine); got != 0 {
		t.Errorf("the source took %d damage; the damage is one-sided", got)
	}
}

// --- Pond Prophet ------------------------------------------------------------

func TestBlbMPPondProphetDrawsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedLibrary(me, "Draw Me")
	before := me.Hand.Size()
	prophet := castHoldCreature(t, g, "Pond Prophet", "Creature — Frog Advisor", blbMPProphetOracle, 1, 1)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(prophet) {
		t.Fatal("the Prophet is not on the battlefield")
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d (the Prophet left, a card was drawn)", got, before+1)
	}
}

// --- Psychic Whorl -----------------------------------------------------------

// blbMPCastWhorl casts Psychic Whorl at opp, who holds three cards, and
// answers the discard off the top of their hand.
func blbMPCastWhorl(t *testing.T, g *game.Game, opp *game.Player) {
	t.Helper()
	for _, n := range []string{"One", "Two", "Three"} {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	castCatalogSpell(t, g, "Psychic Whorl", "Sorcery", blbMPWhorlOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, opp.ID)
	g.SettleResolution()
	passPriorityAroundTable(t, g)
}

func TestBlbMPPsychicWhorlDiscardsTwoAndSurveilsWithARat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B")
	blbB1Push(g, me.ID, "Rat", "Creature — Rat Rogue", "", 1, 1)
	hand, yard := opp.Hand.Size(), opp.Graveyard.Size()
	blbMPCastWhorl(t, g, opp)
	if got := opp.Hand.Size(); got != hand+1 {
		t.Errorf("the opponent holds %d cards, want %d", got, hand+1)
	}
	if got := opp.Graveyard.Size(); got != yard+2 {
		t.Errorf("the opponent's graveyard has %d cards, want %d", got, yard+2)
	}
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("with a Rat the spell should surveil 2")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("surveil looked at %d cards, want 2", len(c.ScryCards))
	}
}

func TestBlbMPPsychicWhorlWithoutARatOnlyDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B")
	yard := opp.Graveyard.Size()
	blbMPCastWhorl(t, g, opp)
	if got := opp.Graveyard.Size(); got != yard+2 {
		t.Errorf("the opponent's graveyard has %d cards, want %d", got, yard+2)
	}
	if surveilChoiceFor(g, me.ID) != nil {
		t.Error("without a Rat there is no surveil")
	}
}
