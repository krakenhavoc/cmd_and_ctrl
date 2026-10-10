package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reality Fracture slice fra-planeswalker-a (tracker #2795).

const (
	rfaAjaniResolute      = "0d60c095-4c6c-4618-8ef3-508f3d454efa"
	rfaAjaniUnrelenting   = "71620e92-45a4-49cb-94d5-348013ecf886"
	rfaAwakenInferno      = "e729fce5-a1cf-4e95-9fd5-fccf4f704d31"
	rfaBreakUnderPress    = "8dd23c50-cc84-4064-89cd-8d52fbfbd4ab"
	rfaChandra            = "c3dfa1e2-6785-49a0-a194-fb842a8eb63c"
	rfaCompelBrutality    = "3a9059ac-c595-4e71-b704-b04d2d47db70"
	rfaCraftworkCrusher   = "d25284b9-50e4-4dcb-9aaf-6f6478c4a90a"
	rfaDedicatedCommons   = "d4e0749a-e179-4c19-8536-bf8f1a198549"
	rfaEdgar              = "cab2bc93-f38b-4301-ac42-05765615586b"
	rfaEntrustTheSpark    = "b796ecc8-a2e8-4ab9-8736-6825b650434d"
	rfaEssenceBurn        = "f1216f36-00fd-4d03-85c1-f4910a02d4d0"
	rfaExtendedAbsence    = "973fb1e3-83d5-4763-bea8-799b426a8848"
	rfaFaceYourself       = "3c0fe71c-bee1-41d3-aeb5-1723028eedde"
	rfaFateholdAnnex      = "d35876a3-e891-43d3-a50c-69e8469497e0"
	rfaFlourishingGrapple = "6b4f1569-2025-46a8-89ba-79859208b04e"
	rfaFormidableCommons  = "3e51c060-1e86-4a0e-8fee-8b7273161468"
	rfaFulminousForte     = "2a29d028-d1a4-413f-9c01-d6842294420a"
)

func rfaCards(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func rfaColoredCreature(g *game.Game, owner uuid.UUID, name string, p, t int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", Colors: colors,
		Power: p, Toughness: t, Owner: owner, Controller: owner,
	})
}

func rfaColoredWalker(g *game.Game, owner uuid.UUID, name string, loyalty int, colors ...string) uuid.UUID {
	id := uuid.New()
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: id, Name: name, TypeLine: "Legendary Planeswalker — Test", Colors: colors,
		Owner: owner, Controller: owner, Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
	return id
}

func rfaTopOfLibrary(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

func rfaOnBattlefield(g *game.Game, id uuid.UUID) bool {
	_, ok := battlefieldCard(g, id)
	return ok
}

func TestRFAAllSeventeenAreRegistered(t *testing.T) {
	for _, id := range []string{
		rfaAjaniResolute, rfaAjaniUnrelenting, rfaAwakenInferno, rfaBreakUnderPress, rfaChandra,
		rfaCompelBrutality, rfaCraftworkCrusher, rfaDedicatedCommons, rfaEdgar, rfaEntrustTheSpark,
		rfaEssenceBurn, rfaExtendedAbsence, rfaFaceYourself, rfaFateholdAnnex, rfaFlourishingGrapple,
		rfaFormidableCommons, rfaFulminousForte,
	} {
		spec, ok := Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness == CompletenessUnreviewed {
			t.Errorf("%s ships without a completeness declaration", spec.Name)
		}
	}
	if spec, _ := Lookup(rfaChandra); spec.Completeness != CompletenessCaveats {
		t.Error("Chandra omits her -X, so she declares caveats")
	}
}

// --- the three lands ---------------------------------------------------

func TestRFACommonsEnterTappedWithoutAPlaneswalker(t *testing.T) {
	for name, oracle := range map[string]string{
		"Dedicated Commons":  rfaDedicatedCommons,
		"Fatehold Annex":     rfaFateholdAnnex,
		"Formidable Commons": rfaFormidableCommons,
	} {
		g := newCatalogGame(t)
		id := playLandFromHand(t, g, name, oracle)
		top100AssertEnteredTapped(t, g, id, name)
	}
}

func TestRFACommonsEnterUntappedWithAPlaneswalker(t *testing.T) {
	for name, oracle := range map[string]string{
		"Dedicated Commons":  rfaDedicatedCommons,
		"Fatehold Annex":     rfaFateholdAnnex,
		"Formidable Commons": rfaFormidableCommons,
	} {
		g := newCatalogGame(t)
		me := top100ActiveSeat(g)
		pushWalkerForTest(g, me.ID, "Walker", "", 3)
		id := playLandFromHand(t, g, name, oracle)
		top100AssertEnteredUntapped(t, g, id, name)
	}
}

// Only YOUR planeswalker counts.
func TestRFACommonsIgnoreAnOpponentsPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	for _, s := range g.Seats {
		if s.ID != me.ID {
			pushWalkerForTest(g, s.ID, "Their Walker", "", 3)
		}
	}
	id := playLandFromHand(t, g, "Dedicated Commons", rfaDedicatedCommons)
	top100AssertEnteredTapped(t, g, id, "Dedicated Commons")
}

func TestRFACommonsMakeTheirTwoColours(t *testing.T) {
	for oracle, want := range map[string]string{
		rfaDedicatedCommons:  "{R|W}",
		rfaFateholdAnnex:     "{W|U}",
		rfaFormidableCommons: "{B|G}",
	} {
		spec, _ := Lookup(oracle)
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != want {
			t.Errorf("%s: mana abilities %+v, want one producing %s", spec.Name, spec.ManaAbilities, want)
		}
	}
}

// --- instants and sorceries --------------------------------------------

func TestRFABreakUnderPressureTakesTheGreatestAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := b10Creature(g, opp.ID, "Small", "Creature — Bear", "{1}{G}", 2, 2)
	big := b10Creature(g, opp.ID, "Big", "Creature — Giant", "{4}{G}", 5, 5)
	mine := b10Creature(g, me.ID, "Mine", "Creature — Giant", "{7}", 7, 7)
	life := me.Life

	castCatalogSpell(t, g, "Break Under Pressure", "Instant", rfaBreakUnderPress,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	c := sacrificeChoiceFor(g, opp.ID)
	if c == nil || len(c.SacrificeOptions) != 1 || c.SacrificeOptions[0] != big {
		t.Fatalf("the opponent is offered only the greatest, got %+v", c)
	}
	answerSacrifice(t, g, opp.ID, big)
	if rfaOnBattlefield(g, big) || !rfaOnBattlefield(g, small) || !rfaOnBattlefield(g, mine) {
		t.Error("only the opponent's greatest goes")
	}
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	if sacrificeChoiceFor(g, g.Seats[2].ID) != nil {
		t.Error("only the targeted opponent is asked")
	}
}

func TestRFABreakUnderPressureCountsPlaneswalkersAndStillGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b10Creature(g, opp.ID, "Small", "Creature — Bear", "{1}{G}", 2, 2)
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test", ManaCost: "{3}{U}{U}",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterLoyalty: 4},
	})
	castCatalogSpell(t, g, "Break Under Pressure", "Instant", rfaBreakUnderPress,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	c := sacrificeChoiceFor(g, opp.ID)
	if c == nil || len(c.SacrificeOptions) != 1 || c.SacrificeOptions[0] != walker {
		t.Fatalf("the planeswalker has the greatest mana value, got %+v", c)
	}

	// An opponent with nothing still lets you gain the life.
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	life := me2.Life
	castCatalogSpell(t, g2, "Break Under Pressure", "Instant", rfaBreakUnderPress,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp2.ID}})
	passPriorityAroundTable(t, g2)
	if me2.Life != life+2 {
		t.Errorf("no creature to take, still +2 life: %d → %d", life, me2.Life)
	}
	_ = me
}

func TestRFAEssenceBurnExilesABlackOrGreenCreatureOrPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	green := rfaColoredCreature(g, opp.ID, "Green Bear", 2, 2, "G")
	castCatalogSpell(t, g, "Essence Burn", "Instant", rfaEssenceBurn, rfaCards(green))
	passPriorityAroundTable(t, g)
	if rfaOnBattlefield(g, green) || !inExile(g, green) {
		t.Error("5 damage kills the Bear and it is exiled instead of dying")
	}

	black := rfaColoredWalker(g, opp.ID, "Black Walker", 4, "B")
	castCatalogSpell(t, g, "Essence Burn", "Instant", rfaEssenceBurn, rfaCards(black))
	passPriorityAroundTable(t, g)
	if rfaOnBattlefield(g, black) || !inExile(g, black) {
		t.Error("a black planeswalker is a legal target and is exiled when 5 loyalty comes off")
	}
}

func TestRFAEssenceBurnRefusesOtherColours(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	red := rfaColoredCreature(g, opp.ID, "Red Bear", 2, 2, "R")
	colorless := rfaColoredCreature(g, opp.ID, "Golem", 3, 3)
	for _, id := range []uuid.UUID{red, colorless} {
		if err := castCatalogSpellErr(t, g, "Essence Burn", "Instant", rfaEssenceBurn, rfaCards(id)); err != game.ErrIllegalTarget {
			t.Errorf("err %v, want ErrIllegalTarget", err)
		}
	}
}

func TestRFAExtendedAbsenceExilesAndDrainsOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := rfaColoredCreature(g, opp.ID, "Bear", 2, 2, "G")
	life, oppLife := me.Life, map[uuid.UUID]int{}
	for _, s := range g.Seats {
		oppLife[s.ID] = s.Life
	}
	castCatalogSpell(t, g, "Extended Absence", "Instant", rfaExtendedAbsence, rfaCards(bear))
	passPriorityAroundTable(t, g)
	if !inExile(g, bear) {
		t.Error("the target is exiled")
	}
	if me.Life != life+1 {
		t.Errorf("you gain 1: %d → %d", life, me.Life)
	}
	for _, s := range g.Seats {
		if s.ID != me.ID && s.Life != oppLife[s.ID]-1 {
			t.Errorf("each opponent loses 1: %d → %d", oppLife[s.ID], s.Life)
		}
	}
}

func TestRFAExtendedAbsenceExilesAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	walker := rfaColoredWalker(g, opp.ID, "Walker", 6, "U")
	castCatalogSpell(t, g, "Extended Absence", "Instant", rfaExtendedAbsence, rfaCards(walker))
	passPriorityAroundTable(t, g)
	if !inExile(g, walker) {
		t.Error("a planeswalker is a legal target")
	}
	if err := castCatalogSpellErr(t, g, "Extended Absence", "Instant", rfaExtendedAbsence, rfaCards(seedPermanentFor(g, opp.ID, "Rock", "Artifact"))); err != game.ErrIllegalTarget {
		t.Errorf("an artifact is no target: %v", err)
	}
}

func TestRFAFulminousForteSweepsOpponentsCreaturesAndPlaneswalkers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := rfaColoredCreature(g, me.ID, "Mine", 3, 3, "R")
	myWalker := rfaColoredWalker(g, me.ID, "My Walker", 4, "R")
	theirs := rfaColoredCreature(g, opp.ID, "Theirs", 3, 3, "G")
	theirWalker := rfaColoredWalker(g, opp.ID, "Their Walker", 4, "G")
	castModal(t, g, "Fulminous Forte", "Instant", rfaFulminousForte, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if damageMarkedOn(g, theirs) != 1 {
		t.Errorf("an opponent's creature takes 1: %d", damageMarkedOn(g, theirs))
	}
	if loyaltyOf(g, theirWalker) != 3 {
		t.Errorf("an opponent's planeswalker takes 1: loyalty %d", loyaltyOf(g, theirWalker))
	}
	if damageMarkedOn(g, mine) != 0 || loyaltyOf(g, myWalker) != 4 {
		t.Error("your own permanents are untouched")
	}
}

func TestRFAFulminousForteDealsFiveToATarget(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	walker := rfaColoredWalker(g, opp.ID, "Their Walker", 6, "G")
	castModal(t, g, "Fulminous Forte", "Instant", rfaFulminousForte, []int{1}, rfaCards(walker))
	passPriorityAroundTable(t, g)
	if loyaltyOf(g, walker) != 1 {
		t.Errorf("5 damage to a 6-loyalty walker: %d", loyaltyOf(g, walker))
	}
}

func TestRFACompelBrutalityCreatureBite(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := rfaColoredCreature(g, me.ID, "Beast", 4, 4, "G")
	theirs := rfaColoredCreature(g, opp.ID, "Bear", 2, 2, "G")
	castModal(t, g, "Compel Brutality", "Instant", rfaCompelBrutality, []int{0},
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if rfaOnBattlefield(g, theirs) {
		t.Error("4 damage kills the Bear")
	}
	if damageMarkedOn(g, mine) != 0 {
		t.Error("the bite is one-sided")
	}
}

func TestRFACompelBrutalityLoyaltyBite(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := rfaColoredWalker(g, me.ID, "My Walker", 5, "G")
	theirs := rfaColoredCreature(g, opp.ID, "Giant", 5, 5, "G")
	castModal(t, g, "Compel Brutality", "Instant", rfaCompelBrutality, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: walker}, {Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if rfaOnBattlefield(g, theirs) {
		t.Error("5 damage (its loyalty) kills the 5/5")
	}
	if loyaltyOf(g, walker) != 5 {
		t.Errorf("the walker is unharmed: loyalty %d", loyaltyOf(g, walker))
	}
}

func TestRFACompelBrutalityRefusesYourOwnVictim(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := rfaColoredCreature(g, me.ID, "A", 2, 2, "G")
	b := rfaColoredCreature(g, me.ID, "B", 2, 2, "G")
	err := func() error {
		active := g.Seats[g.Turn.ActiveSeat]
		id := uuid.New()
		active.Hand.PushTop(game.Card{InstanceID: id, Name: "Compel Brutality", TypeLine: "Instant",
			OracleID: rfaCompelBrutality, Owner: active.ID, Controller: active.ID})
		advanceToMain(t, g)
		return g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{0},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}})
	}()
	if err != game.ErrIllegalTarget {
		t.Errorf("the victim must be an opponent's: %v", err)
	}
}

func TestRFAFlourishingGrappleStripsAbilitiesAndBites(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := rfaColoredCreature(g, me.ID, "Beast", 3, 3, "G")
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird", Colors: []string{"W"},
		Power: 2, Toughness: 6, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	if !hasEffectiveKeyword(t, g, victim, "flying") {
		t.Fatal("fixture: the Flier flies")
	}
	castCatalogSpell(t, g, "Flourishing Grapple", "Instant", rfaFlourishingGrapple,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}, {Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, victim, "flying") {
		t.Error("the Flier loses all abilities until end of turn")
	}
	if damageMarkedOn(g, victim) != 3 {
		t.Errorf("the Beast deals 3: %d", damageMarkedOn(g, victim))
	}
	if damageMarkedOn(g, mine) != 0 {
		t.Error("one-sided")
	}
}

func TestRFAFlourishingGrappleOnlyRedOrWhiteOpposition(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := rfaColoredCreature(g, me.ID, "Beast", 3, 3, "G")
	green := rfaColoredCreature(g, opp.ID, "Green", 2, 2, "G")
	ownRed := rfaColoredCreature(g, me.ID, "Own Red", 2, 2, "R")
	for _, v := range []uuid.UUID{green, ownRed} {
		err := castCatalogSpellErr(t, g, "Flourishing Grapple", "Instant", rfaFlourishingGrapple,
			[]game.TargetRef{{Kind: game.TargetCard, ID: v}, {Kind: game.TargetCard, ID: mine}})
		if err != game.ErrIllegalTarget {
			t.Errorf("err %v, want ErrIllegalTarget", err)
		}
	}
}

func TestRFAAwakenTheInfernoBurnsAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := rfaColoredCreature(g, me.ID, "Mine", 2, 2, "R")
	theirs := rfaColoredCreature(g, opp.ID, "Theirs", 6, 6, "G")
	castCatalogSpell(t, g, "Awaken the Inferno", "Sorcery", rfaAwakenInferno,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}, {Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)
	if rfaOnBattlefield(g, theirs) {
		t.Error("6 damage kills the 6/6")
	}
	if counterOn(g, mine, game.CounterPlusOne) != 1 {
		t.Error("a +1/+1 counter on the creature you control")
	}
}

func TestRFAAwakenTheInfernoWorksWithoutTheCounterTarget(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	walker := rfaColoredWalker(g, opp.ID, "Walker", 7, "U")
	castCatalogSpell(t, g, "Awaken the Inferno", "Sorcery", rfaAwakenInferno, rfaCards(walker))
	passPriorityAroundTable(t, g)
	if loyaltyOf(g, walker) != 1 {
		t.Errorf("6 damage to a 7-loyalty walker: %d", loyaltyOf(g, walker))
	}
	spec, _ := Lookup(rfaAwakenInferno)
	if len(spec.Activated) != 1 {
		t.Errorf("basic landcycling is declared: %d abilities", len(spec.Activated))
	}
}

func TestRFAEntrustTheSparkSacrificesAWalkerForAnother(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	old := pushWalkerForTest(g, me.ID, "Old Walker", "", 3)
	needle := pushLibraryCardForTest(me, game.Card{
		Name: "Library Walker", TypeLine: "Legendary Planeswalker — Test", StartingLoyalty: 4,
		Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Entrust the Spark", "Sorcery", rfaEntrustTheSpark, nil)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	answerSacrifice(t, g, me.ID, old)
	passPriorityAroundTable(t, g)
	if rfaOnBattlefield(g, old) {
		t.Error("the old walker was sacrificed")
	}
	if !rfaOnBattlefield(g, needle) {
		t.Fatalf("the library walker should be on the battlefield: pending %+v", g.PendingChoices)
	}
}

func TestRFAEntrustTheSparkDeclinedDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	old := pushWalkerForTest(g, me.ID, "Old Walker", "", 3)
	needle := pushLibraryCardForTest(me, game.Card{
		Name: "Library Walker", TypeLine: "Legendary Planeswalker — Test", StartingLoyalty: 4,
		Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Entrust the Spark", "Sorcery", rfaEntrustTheSpark, nil)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !rfaOnBattlefield(g, old) || rfaOnBattlefield(g, needle) {
		t.Error("declining sacrifices nothing and fetches nothing")
	}
}

func TestRFAEntrustTheSparkAsksNothingWithoutAWalker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	needle := pushLibraryCardForTest(me, game.Card{
		Name: "Library Walker", TypeLine: "Legendary Planeswalker — Test", StartingLoyalty: 4,
		Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Entrust the Spark", "Sorcery", rfaEntrustTheSpark, nil)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("no planeswalker to sacrifice, so no question")
	}
	if rfaOnBattlefield(g, needle) {
		t.Error("nothing was fetched")
	}
}

func TestRFAFaceYourselfCopiesWithHasteAndSacrificesWithoutAWalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := rfaColoredCreature(g, opp.ID, "Their Bear", 2, 2, "G")
	castCatalogSpell(t, g, "Face Yourself", "Sorcery", rfaFaceYourself,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	var copyID uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Bear" && c.Controller == me.ID {
			copyID = c.InstanceID
		}
	}
	if copyID == uuid.Nil {
		t.Fatal("a copy of their Bear under your control")
	}
	if !hasEffectiveKeyword(t, g, copyID, "haste") {
		t.Error("the copy has haste")
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if rfaOnBattlefield(g, copyID) {
		t.Error("with no planeswalker the copy is sacrificed at the end step")
	}
	if !rfaOnBattlefield(g, bear) {
		t.Error("the original is untouched")
	}
}

func TestRFAFaceYourselfCopiesSurviveWithAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rfaColoredCreature(g, opp.ID, "Their Bear", 2, 2, "G")
	rfaColoredCreature(g, opp.ID, "Their Wolf", 3, 3, "G")
	pushWalkerForTest(g, me.ID, "My Walker", "", 4)
	castCatalogSpell(t, g, "Face Yourself", "Sorcery", rfaFaceYourself,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	mine := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsCreature() {
			mine++
		}
	}
	if mine != 2 {
		t.Fatalf("one copy per creature: %d", mine)
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	mine = 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsCreature() {
			mine++
		}
	}
	if mine != 2 {
		t.Errorf("the copies survive while you control a planeswalker: %d", mine)
	}
}

// --- planeswalkers ------------------------------------------------------

func TestRFAAjaniResolute(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ajani := pushCatalogPermanent(g, me.ID, "Ajani Resolute", "Legendary Planeswalker — Ajani", rfaAjaniResolute, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ajani, game.CounterLoyalty, 4) })
	advanceToMain(t, g)

	// 0: gain 1 life, and the gain ticks him up.
	life := me.Life
	b16Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("0 gains 1 life: %d → %d", life, me.Life)
	}
	if loyaltyOf(g, ajani) != 5 {
		t.Errorf("a loyalty counter per life gain: %d, want 5", loyaltyOf(g, ajani))
	}

}

func TestRFAAjaniResolutesMinusFourMakesAGrowingPridemate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ajani := pushCatalogPermanent(g, me.ID, "Ajani Resolute", "Legendary Planeswalker — Ajani", rfaAjaniResolute, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ajani, game.CounterLoyalty, 5) })
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, ajani, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ajani's Pridemate" && c.Controller == me.ID {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("−4 makes Ajani's Pridemate")
	}
	if effectivePower(t, g, token) != 2 {
		t.Errorf("a 2/2: %d", effectivePower(t, g, token))
	}
	// A real life gain (his own 0, next turn) triggers both his counter and the token.
	b39NextTurnOf(t, g, 0)
	b16Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if triggerOrderPrompt(g) != nil {
		answerTriggerOrderInOfferedOrder(t, g)
		passPriorityAroundTable(t, g)
	}
	if loyaltyOf(g, ajani) != 2 {
		t.Errorf("his own counter too: %d", loyaltyOf(g, ajani))
	}
	if counterOn(g, token, game.CounterPlusOne) != 1 {
		t.Errorf("one +1/+1 counter per gain: %d", counterOn(g, token, game.CounterPlusOne))
	}
}

func TestRFAAjaniResolutesEmblemIsAnAnthem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ajani := pushCatalogPermanent(g, me.ID, "Ajani Resolute", "Legendary Planeswalker — Ajani", rfaAjaniResolute, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ajani, game.CounterLoyalty, 10) })
	bear := seedBear(g, me.ID)
	theirs := b39Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	before := effectivePower(t, g, bear)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, ajani, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, bear) != before+2 {
		t.Errorf("+2/+2 for your creatures: %d → %d", before, effectivePower(t, g, bear))
	}
	if effectivePower(t, g, theirs) != 2 {
		t.Error("not for theirs")
	}
	if me.Emblems.Size() != 1 {
		t.Errorf("one emblem: %d", me.Emblems.Size())
	}
}

func TestRFAAjaniUnrelentingMakesACadetPerLoyaltyActivation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ajani := pushCatalogPermanent(g, me.ID, "Ajani Unrelenting", "Legendary Planeswalker — Ajani", rfaAjaniUnrelenting, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ajani, game.CounterLoyalty, 5) })
	other := pushCatalogPermanent(g, me.ID, "Ajani, the Greathearted", "Legendary Planeswalker — Ajani", b39AjaniOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(other, game.CounterLoyalty, 5) })
	advanceToMain(t, g)

	b16Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Cadet"); n != 1 {
		t.Fatalf("a Cadet for his own activation: %d", n)
	}
	// Another planeswalker's loyalty ability triggers him too.
	b16Activate(t, g, me.ID, other, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Cadet"); n != 2 {
		t.Errorf("and for another walker's: %d", n)
	}
}

func TestRFAAjaniUnrelentingPlusOnePumpsCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ajani := pushCatalogPermanent(g, me.ID, "Ajani Unrelenting", "Legendary Planeswalker — Ajani", rfaAjaniUnrelenting, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ajani, game.CounterLoyalty, 5) })
	bear := seedBear(g, me.ID)
	theirs := b39Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	before := effectivePower(t, g, bear)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, bear) != before+1 || !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Errorf("+1/+0 and haste: power %d → %d", before, effectivePower(t, g, bear))
	}
	if effectivePower(t, g, theirs) != 2 || hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature gets nothing")
	}
}

func TestRFAAjaniUnrelentingMinusTwoDiscardsThenDrawsPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ajani := pushCatalogPermanent(g, me.ID, "Ajani Unrelenting", "Legendary Planeswalker — Ajani", rfaAjaniUnrelenting, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ajani, game.CounterLoyalty, 5) })
	seedBear(g, me.ID)
	seedBear(g, me.ID)
	for i := 0; i < 8; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	advanceToMain(t, g)
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Hand A", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Hand B", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	graveBefore := len(me.Graveyard.Cards)

	b16Activate(t, g, me.ID, ajani, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if len(me.Graveyard.Cards) < graveBefore+2 {
		t.Errorf("the hand was discarded: graveyard %d → %d", graveBefore, len(me.Graveyard.Cards))
	}
	// Two Bears plus the Cadet his own activation made.
	if me.Hand.Size() != 3 {
		t.Errorf("drew a card per creature (2 Bears + the Cadet): hand %d, want 3", me.Hand.Size())
	}
}

func TestRFAAjaniUnrelentingMinusThreeSparesYourTokensOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ajani := pushCatalogPermanent(g, me.ID, "Ajani Unrelenting", "Legendary Planeswalker — Ajani", rfaAjaniUnrelenting, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ajani, game.CounterLoyalty, 5) })
	myTok := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Token", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	myCard := b39Creature(g, me.ID, "My Card", "Creature — Bear", 2, 2)
	theirTok := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Token", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	theirCard := b39Creature(g, opp.ID, "Their Card", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, ajani, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !rfaOnBattlefield(g, myTok) {
		t.Error("your tokens are spared")
	}
	for name, id := range map[string]uuid.UUID{"your nontoken": myCard, "their token": theirTok, "their nontoken": theirCard} {
		if rfaOnBattlefield(g, id) {
			t.Errorf("%s creature takes 4 and dies", name)
		}
	}
	if !rfaOnBattlefield(g, ajani) {
		t.Error("Ajani is a planeswalker and is not damaged")
	}
}

func TestRFAChandraFirstPlusOneReturnsANoncreatureNonland(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chandra := pushCatalogPermanent(g, me.ID, "Chandra, Chill of Compliance", "Legendary Planeswalker — Chandra", rfaChandra, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(chandra, game.CounterLoyalty, 3) })
	spell := rfaTopOfLibrary(me, "Opt", "Instant")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("a surveil prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(spell) {
		t.Error("the binned instant goes to hand")
	}
	if loyaltyOf(g, chandra) != 4 {
		t.Errorf("+1 loyalty: %d", loyaltyOf(g, chandra))
	}
}

func TestRFAChandraFirstPlusOneLeavesACreatureInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chandra := pushCatalogPermanent(g, me.ID, "Chandra, Chill of Compliance", "Legendary Planeswalker — Chandra", rfaChandra, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(chandra, game.CounterLoyalty, 3) })
	bear := rfaTopOfLibrary(me, "Bear", "Creature — Bear")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("a surveil prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(bear) || !me.Graveyard.Contains(bear) {
		t.Error("a creature card stays in the graveyard")
	}
}

func TestRFAChandraFirstPlusOneKeepingTheCardOnTopReturnsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chandra := pushCatalogPermanent(g, me.ID, "Chandra, Chill of Compliance", "Legendary Planeswalker — Chandra", rfaChandra, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(chandra, game.CounterLoyalty, 3) })
	spell := rfaTopOfLibrary(me, "Opt", "Instant")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("a surveil prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("kept on top: neither in hand nor in the graveyard")
	}
}

func TestRFAChandraSecondPlusOneAddsRestrictedBlue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chandra := pushCatalogPermanent(g, me.ID, "Chandra, Chill of Compliance", "Legendary Planeswalker — Chandra", rfaChandra, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(chandra, game.CounterLoyalty, 3) })
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, chandra, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	var toks []game.ManaToken
	g.ReadSnapshot(func() { toks = append(toks, me.ManaPool...) })
	if len(toks) != 1 || toks[0].Color != "U" {
		t.Fatalf("one blue mana: %+v", toks)
	}
	if len(toks[0].Restrictions) != 2 {
		t.Errorf("restricted to casting a noncreature spell: %v", toks[0].Restrictions)
	}
	if loyaltyOf(g, chandra) != 4 {
		t.Errorf("+1 loyalty: %d", loyaltyOf(g, chandra))
	}
}

func TestRFAChandraEmblemDrawsOnEverySpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chandra := pushCatalogPermanent(g, me.ID, "Chandra, Chill of Compliance", "Legendary Planeswalker — Chandra", rfaChandra, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(chandra, game.CounterLoyalty, 6) })
	for i := 0; i < 6; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, chandra, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if me.Emblems.Size() != 1 {
		t.Fatalf("one emblem: %d", me.Emblems.Size())
	}
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Some Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("whenever you cast a spell, draw a card: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestRFAEdgarGainsLifeWhenAnotherCreatureOrPlaneswalkerOfYoursDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Edgar, Ancient Bloodlord", "Legendary Creature — Vampire Noble", rfaEdgar, false)
	bear := seedBear(g, me.ID)
	theirs := b39Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	walker := pushWalkerForTest(g, me.ID, "Walker", "", 3)

	life := me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("your creature dies: %d → %d", life, me.Life)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("an opponent's creature dying gains nothing: %d", me.Life)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(walker) })
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("your planeswalker dies: %d, want %d", me.Life, life+2)
	}
}

func TestRFAEdgarSacrificeGrowsAndGainsMenace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	edgar := pushCatalogPermanent(g, me.ID, "Edgar, Ancient Bloodlord", "Legendary Creature — Vampire Noble", rfaEdgar, false)
	bear := seedBear(g, me.ID)
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool = append(me.ManaPool, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}) })
	b16Activate(t, g, me.ID, edgar, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	passPriorityAroundTable(t, g)
	if rfaOnBattlefield(g, bear) {
		t.Error("the Bear was sacrificed")
	}
	if counterOn(g, edgar, game.CounterPlusOne) != 1 {
		t.Error("a +1/+1 counter on Edgar")
	}
	if !hasEffectiveKeyword(t, g, edgar, "menace") {
		t.Error("menace until end of turn")
	}
}

func TestRFACraftworkCrusherChoosesTwoModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := rfaColoredCreature(g, opp.ID, "Their Bear", 2, 2, "G")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Craftwork Crusher", "Artifact Creature — Boar Construct", rfaCraftworkCrusher, nil)
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("the enters trigger asks for two modes: %+v", g.PendingChoices)
	}
	if c.ModeMin != 2 || c.ModeMax != 2 {
		t.Errorf("choose exactly two: %d–%d", c.ModeMin, c.ModeMax)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("the damage bullet asks for a target: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if rfaOnBattlefield(g, theirs) {
		t.Error("4 damage kills the Bear")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("and draws a card: hand %d → %d", hand, me.Hand.Size())
	}
	if b16CountNamed(g, "Cadet") != 0 {
		t.Error("the Cadet bullet was not chosen")
	}
}

func TestRFACraftworkCrusherCadetAndDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Craftwork Crusher", "Artifact Creature — Boar Construct", rfaCraftworkCrusher, nil)
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("the enters trigger asks for two modes: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1, 2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b16CountNamed(g, "Cadet") != 1 || me.Hand.Size() != hand+1 {
		t.Errorf("a Cadet and a card: cadets %d, hand %d → %d", b16CountNamed(g, "Cadet"), hand, me.Hand.Size())
	}
}
