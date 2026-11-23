package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func rfaCards(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func rfaColoredWalker(g *game.Game, owner uuid.UUID, name string, loyalty int, colors ...string) uuid.UUID {
	id := uuid.New()
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
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
	_, ok := effectstest.BattlefieldCard(g, id)
	return ok
}

func TestRFAAllSeventeenAreRegistered(t *testing.T) {
	for _, id := range []string{
		effectstest.RfaAjaniResolute, effectstest.RfaAjaniUnrelenting, effectstest.RfaAwakenInferno, effectstest.RfaBreakUnderPress, effectstest.RfaChandra,
		effectstest.RfaCompelBrutality, effectstest.RfaCraftworkCrusher, effectstest.RfaDedicatedCommons, effectstest.RfaEdgar, effectstest.RfaEntrustTheSpark,
		effectstest.RfaEssenceBurn, effectstest.RfaExtendedAbsence, effectstest.RfaFaceYourself, effectstest.RfaFateholdAnnex, effectstest.RfaFlourishingGrapple,
		effectstest.RfaFormidableCommons, effectstest.RfaFulminousForte,
	} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness == effects.CompletenessUnreviewed {
			t.Errorf("%s ships without a completeness declaration", spec.Name)
		}
	}
	if spec, _ := effects.Lookup(effectstest.RfaChandra); spec.Completeness != effects.CompletenessFull {
		t.Error("Chandra has all four abilities since her -X landed (#1944), so she is Full")
	}
}

// --- the three lands ---------------------------------------------------

func TestRFACommonsEnterTappedWithoutAPlaneswalker(t *testing.T) {
	for name, oracle := range map[string]string{
		"Dedicated Commons":  effectstest.RfaDedicatedCommons,
		"Fatehold Annex":     effectstest.RfaFateholdAnnex,
		"Formidable Commons": effectstest.RfaFormidableCommons,
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.PlayLandFromHand(t, g, name, oracle)
		effectstest.Top100AssertEnteredTapped(t, g, id, name)
	}
}

func TestRFACommonsEnterUntappedWithAPlaneswalker(t *testing.T) {
	for name, oracle := range map[string]string{
		"Dedicated Commons":  effectstest.RfaDedicatedCommons,
		"Fatehold Annex":     effectstest.RfaFateholdAnnex,
		"Formidable Commons": effectstest.RfaFormidableCommons,
	} {
		g := effectstest.NewCatalogGame(t)
		me := effectstest.Top100ActiveSeat(g)
		effectstest.PushWalkerForTest(g, me.ID, "Walker", "", 3)
		id := effectstest.PlayLandFromHand(t, g, name, oracle)
		effectstest.Top100AssertEnteredUntapped(t, g, id, name)
	}
}

// Only YOUR planeswalker counts.
func TestRFACommonsIgnoreAnOpponentsPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	for _, s := range g.Seats {
		if s.ID != me.ID {
			effectstest.PushWalkerForTest(g, s.ID, "Their Walker", "", 3)
		}
	}
	id := effectstest.PlayLandFromHand(t, g, "Dedicated Commons", effectstest.RfaDedicatedCommons)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Dedicated Commons")
}

func TestRFACommonsMakeTheirTwoColours(t *testing.T) {
	for oracle, want := range map[string]string{
		effectstest.RfaDedicatedCommons:  "{R|W}",
		effectstest.RfaFateholdAnnex:     "{W|U}",
		effectstest.RfaFormidableCommons: "{B|G}",
	} {
		spec, _ := effects.Lookup(oracle)
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != want {
			t.Errorf("%s: mana abilities %+v, want one producing %s", spec.Name, spec.ManaAbilities, want)
		}
	}
}

// --- instants and sorceries --------------------------------------------

func TestRFABreakUnderPressureTakesTheGreatestAndGainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B10Creature(g, opp.ID, "Small", "Creature — Bear", "{1}{G}", 2, 2)
	big := effectstest.B10Creature(g, opp.ID, "Big", "Creature — Giant", "{4}{G}", 5, 5)
	mine := effectstest.B10Creature(g, me.ID, "Mine", "Creature — Giant", "{7}", 7, 7)
	life := me.Life

	effectstest.CastCatalogSpell(t, g, "Break Under Pressure", "Instant", effectstest.RfaBreakUnderPress,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SacrificeChoiceFor(g, opp.ID)
	if c == nil || len(c.SacrificeOptions) != 1 || c.SacrificeOptions[0] != big {
		t.Fatalf("the opponent is offered only the greatest, got %+v", c)
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, big)
	if rfaOnBattlefield(g, big) || !rfaOnBattlefield(g, small) || !rfaOnBattlefield(g, mine) {
		t.Error("only the opponent's greatest goes")
	}
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	if effectstest.SacrificeChoiceFor(g, g.Seats[2].ID) != nil {
		t.Error("only the targeted opponent is asked")
	}
}

func TestRFABreakUnderPressureCountsPlaneswalkersAndStillGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B10Creature(g, opp.ID, "Small", "Creature — Bear", "{1}{G}", 2, 2)
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test", ManaCost: "{3}{U}{U}",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterLoyalty: 4},
	})
	effectstest.CastCatalogSpell(t, g, "Break Under Pressure", "Instant", effectstest.RfaBreakUnderPress,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SacrificeChoiceFor(g, opp.ID)
	if c == nil || len(c.SacrificeOptions) != 1 || c.SacrificeOptions[0] != walker {
		t.Fatalf("the planeswalker has the greatest mana value, got %+v", c)
	}

	// An opponent with nothing still lets you gain the life.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	life := me2.Life
	effectstest.CastCatalogSpell(t, g2, "Break Under Pressure", "Instant", effectstest.RfaBreakUnderPress,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp2.ID}})
	effectstest.PassPriorityAroundTable(t, g2)
	if me2.Life != life+2 {
		t.Errorf("no creature to take, still +2 life: %d → %d", life, me2.Life)
	}
	_ = me
}

func TestRFAEssenceBurnExilesABlackOrGreenCreatureOrPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	green := effectstest.RfaColoredCreature(g, opp.ID, "Green Bear", 2, 2, "G")
	effectstest.CastCatalogSpell(t, g, "Essence Burn", "Instant", effectstest.RfaEssenceBurn, rfaCards(green))
	effectstest.PassPriorityAroundTable(t, g)
	if rfaOnBattlefield(g, green) || !effectstest.InExile(g, green) {
		t.Error("5 damage kills the Bear and it is exiled instead of dying")
	}

	black := rfaColoredWalker(g, opp.ID, "Black Walker", 4, "B")
	effectstest.CastCatalogSpell(t, g, "Essence Burn", "Instant", effectstest.RfaEssenceBurn, rfaCards(black))
	effectstest.PassPriorityAroundTable(t, g)
	if rfaOnBattlefield(g, black) || !effectstest.InExile(g, black) {
		t.Error("a black planeswalker is a legal target and is exiled when 5 loyalty comes off")
	}
}

func TestRFAEssenceBurnRefusesOtherColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	red := effectstest.RfaColoredCreature(g, opp.ID, "Red Bear", 2, 2, "R")
	colorless := effectstest.RfaColoredCreature(g, opp.ID, "Golem", 3, 3)
	for _, id := range []uuid.UUID{red, colorless} {
		if err := effectstest.CastCatalogSpellErr(t, g, "Essence Burn", "Instant", effectstest.RfaEssenceBurn, rfaCards(id)); err != game.ErrIllegalTarget {
			t.Errorf("err %v, want ErrIllegalTarget", err)
		}
	}
}

func TestRFAExtendedAbsenceExilesAndDrainsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.RfaColoredCreature(g, opp.ID, "Bear", 2, 2, "G")
	life, oppLife := me.Life, map[uuid.UUID]int{}
	for _, s := range g.Seats {
		oppLife[s.ID] = s.Life
	}
	effectstest.CastCatalogSpell(t, g, "Extended Absence", "Instant", effectstest.RfaExtendedAbsence, rfaCards(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, bear) {
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	walker := rfaColoredWalker(g, opp.ID, "Walker", 6, "U")
	effectstest.CastCatalogSpell(t, g, "Extended Absence", "Instant", effectstest.RfaExtendedAbsence, rfaCards(walker))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, walker) {
		t.Error("a planeswalker is a legal target")
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Extended Absence", "Instant", effectstest.RfaExtendedAbsence, rfaCards(effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact"))); err != game.ErrIllegalTarget {
		t.Errorf("an artifact is no target: %v", err)
	}
}

func TestRFAFulminousForteSweepsOpponentsCreaturesAndPlaneswalkers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.RfaColoredCreature(g, me.ID, "Mine", 3, 3, "R")
	myWalker := rfaColoredWalker(g, me.ID, "My Walker", 4, "R")
	theirs := effectstest.RfaColoredCreature(g, opp.ID, "Theirs", 3, 3, "G")
	theirWalker := rfaColoredWalker(g, opp.ID, "Their Walker", 4, "G")
	effectstest.CastModal(t, g, "Fulminous Forte", "Instant", effectstest.RfaFulminousForte, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageMarkedOn(g, theirs) != 1 {
		t.Errorf("an opponent's creature takes 1: %d", effectstest.DamageMarkedOn(g, theirs))
	}
	if effectstest.LoyaltyOf(g, theirWalker) != 3 {
		t.Errorf("an opponent's planeswalker takes 1: loyalty %d", effectstest.LoyaltyOf(g, theirWalker))
	}
	if effectstest.DamageMarkedOn(g, mine) != 0 || effectstest.LoyaltyOf(g, myWalker) != 4 {
		t.Error("your own permanents are untouched")
	}
}

func TestRFAFulminousForteDealsFiveToATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	walker := rfaColoredWalker(g, opp.ID, "Their Walker", 6, "G")
	effectstest.CastModal(t, g, "Fulminous Forte", "Instant", effectstest.RfaFulminousForte, []int{1}, rfaCards(walker))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoyaltyOf(g, walker) != 1 {
		t.Errorf("5 damage to a 6-loyalty walker: %d", effectstest.LoyaltyOf(g, walker))
	}
}

func TestRFACompelBrutalityCreatureBite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.RfaColoredCreature(g, me.ID, "Beast", 4, 4, "G")
	theirs := effectstest.RfaColoredCreature(g, opp.ID, "Bear", 2, 2, "G")
	effectstest.CastModal(t, g, "Compel Brutality", "Instant", effectstest.RfaCompelBrutality, []int{0},
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if rfaOnBattlefield(g, theirs) {
		t.Error("4 damage kills the Bear")
	}
	if effectstest.DamageMarkedOn(g, mine) != 0 {
		t.Error("the bite is one-sided")
	}
}

func TestRFACompelBrutalityLoyaltyBite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := rfaColoredWalker(g, me.ID, "My Walker", 5, "G")
	theirs := effectstest.RfaColoredCreature(g, opp.ID, "Giant", 5, 5, "G")
	effectstest.CastModal(t, g, "Compel Brutality", "Instant", effectstest.RfaCompelBrutality, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: walker}, {Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if rfaOnBattlefield(g, theirs) {
		t.Error("5 damage (its loyalty) kills the 5/5")
	}
	if effectstest.LoyaltyOf(g, walker) != 5 {
		t.Errorf("the walker is unharmed: loyalty %d", effectstest.LoyaltyOf(g, walker))
	}
}

func TestRFACompelBrutalityRefusesYourOwnVictim(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.RfaColoredCreature(g, me.ID, "A", 2, 2, "G")
	b := effectstest.RfaColoredCreature(g, me.ID, "B", 2, 2, "G")
	err := func() error {
		active := g.Seats[g.Turn.ActiveSeat]
		id := uuid.New()
		active.Hand.PushTop(game.Card{InstanceID: id, Name: "Compel Brutality", TypeLine: "Instant",
			OracleID: effectstest.RfaCompelBrutality, Owner: active.ID, Controller: active.ID})
		effectstest.AdvanceToMain(t, g)
		return g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{0},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}})
	}()
	if err != game.ErrIllegalTarget {
		t.Errorf("the victim must be an opponent's: %v", err)
	}
}

func TestRFAFlourishingGrappleStripsAbilitiesAndBites(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.RfaColoredCreature(g, me.ID, "Beast", 3, 3, "G")
	victim := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird", Colors: []string{"W"},
		Power: 2, Toughness: 6, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	if !effectstest.HasEffectiveKeyword(t, g, victim, "flying") {
		t.Fatal("fixture: the Flier flies")
	}
	effectstest.CastCatalogSpell(t, g, "Flourishing Grapple", "Instant", effectstest.RfaFlourishingGrapple,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}, {Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, victim, "flying") {
		t.Error("the Flier loses all abilities until end of turn")
	}
	if effectstest.DamageMarkedOn(g, victim) != 3 {
		t.Errorf("the Beast deals 3: %d", effectstest.DamageMarkedOn(g, victim))
	}
	if effectstest.DamageMarkedOn(g, mine) != 0 {
		t.Error("one-sided")
	}
}

func TestRFAFlourishingGrappleOnlyRedOrWhiteOpposition(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.RfaColoredCreature(g, me.ID, "Beast", 3, 3, "G")
	green := effectstest.RfaColoredCreature(g, opp.ID, "Green", 2, 2, "G")
	ownRed := effectstest.RfaColoredCreature(g, me.ID, "Own Red", 2, 2, "R")
	for _, v := range []uuid.UUID{green, ownRed} {
		err := effectstest.CastCatalogSpellErr(t, g, "Flourishing Grapple", "Instant", effectstest.RfaFlourishingGrapple,
			[]game.TargetRef{{Kind: game.TargetCard, ID: v}, {Kind: game.TargetCard, ID: mine}})
		if err != game.ErrIllegalTarget {
			t.Errorf("err %v, want ErrIllegalTarget", err)
		}
	}
}

func TestRFAAwakenTheInfernoBurnsAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.RfaColoredCreature(g, me.ID, "Mine", 2, 2, "R")
	theirs := effectstest.RfaColoredCreature(g, opp.ID, "Theirs", 6, 6, "G")
	effectstest.CastCatalogSpell(t, g, "Awaken the Inferno", "Sorcery", effectstest.RfaAwakenInferno,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}, {Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)
	if rfaOnBattlefield(g, theirs) {
		t.Error("6 damage kills the 6/6")
	}
	if effectstest.CounterOn(g, mine, game.CounterPlusOne) != 1 {
		t.Error("a +1/+1 counter on the creature you control")
	}
}

func TestRFAAwakenTheInfernoWorksWithoutTheCounterTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	walker := rfaColoredWalker(g, opp.ID, "Walker", 7, "U")
	effectstest.CastCatalogSpell(t, g, "Awaken the Inferno", "Sorcery", effectstest.RfaAwakenInferno, rfaCards(walker))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoyaltyOf(g, walker) != 1 {
		t.Errorf("6 damage to a 7-loyalty walker: %d", effectstest.LoyaltyOf(g, walker))
	}
	spec, _ := effects.Lookup(effectstest.RfaAwakenInferno)
	if len(spec.Activated) != 1 {
		t.Errorf("basic landcycling is declared: %d abilities", len(spec.Activated))
	}
}

func TestRFAEntrustTheSparkSacrificesAWalkerForAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	old := effectstest.PushWalkerForTest(g, me.ID, "Old Walker", "", 3)
	needle := effectstest.PushLibraryCardForTest(me, game.Card{
		Name: "Library Walker", TypeLine: "Legendary Planeswalker — Test", StartingLoyalty: 4,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Entrust the Spark", "Sorcery", effectstest.RfaEntrustTheSpark, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerSacrifice(t, g, me.ID, old)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if rfaOnBattlefield(g, old) {
		t.Error("the old walker was sacrificed")
	}
	if !rfaOnBattlefield(g, needle) {
		t.Fatalf("the library walker should be on the battlefield: pending %+v", g.PendingChoices)
	}
}

func TestRFAEntrustTheSparkDeclinedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	old := effectstest.PushWalkerForTest(g, me.ID, "Old Walker", "", 3)
	needle := effectstest.PushLibraryCardForTest(me, game.Card{
		Name: "Library Walker", TypeLine: "Legendary Planeswalker — Test", StartingLoyalty: 4,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Entrust the Spark", "Sorcery", effectstest.RfaEntrustTheSpark, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !rfaOnBattlefield(g, old) || rfaOnBattlefield(g, needle) {
		t.Error("declining sacrifices nothing and fetches nothing")
	}
}

func TestRFAEntrustTheSparkAsksNothingWithoutAWalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	needle := effectstest.PushLibraryCardForTest(me, game.Card{
		Name: "Library Walker", TypeLine: "Legendary Planeswalker — Test", StartingLoyalty: 4,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Entrust the Spark", "Sorcery", effectstest.RfaEntrustTheSpark, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("no planeswalker to sacrifice, so no question")
	}
	if rfaOnBattlefield(g, needle) {
		t.Error("nothing was fetched")
	}
}

func TestRFAFaceYourselfCopiesWithHasteAndSacrificesWithoutAWalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.RfaColoredCreature(g, opp.ID, "Their Bear", 2, 2, "G")
	effectstest.CastCatalogSpell(t, g, "Face Yourself", "Sorcery", effectstest.RfaFaceYourself,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	var copyID uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Bear" && c.Controller == me.ID {
			copyID = c.InstanceID
		}
	}
	if copyID == uuid.Nil {
		t.Fatal("a copy of their Bear under your control")
	}
	if !effectstest.HasEffectiveKeyword(t, g, copyID, "haste") {
		t.Error("the copy has haste")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if rfaOnBattlefield(g, copyID) {
		t.Error("with no planeswalker the copy is sacrificed at the end step")
	}
	if !rfaOnBattlefield(g, bear) {
		t.Error("the original is untouched")
	}
}

func TestRFAFaceYourselfCopiesSurviveWithAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.RfaColoredCreature(g, opp.ID, "Their Bear", 2, 2, "G")
	effectstest.RfaColoredCreature(g, opp.ID, "Their Wolf", 3, 3, "G")
	effectstest.PushWalkerForTest(g, me.ID, "My Walker", "", 4)
	effectstest.CastCatalogSpell(t, g, "Face Yourself", "Sorcery", effectstest.RfaFaceYourself,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	mine := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsCreature() {
			mine++
		}
	}
	if mine != 2 {
		t.Fatalf("one copy per creature: %d", mine)
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ajani := effectstest.PushCatalogPermanent(g, me.ID, "Ajani Resolute", "Legendary Planeswalker — Ajani", effectstest.RfaAjaniResolute, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ajani, game.CounterLoyalty, 4) })
	effectstest.AdvanceToMain(t, g)

	// 0: gain 1 life, and the gain ticks him up.
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("0 gains 1 life: %d → %d", life, me.Life)
	}
	if effectstest.LoyaltyOf(g, ajani) != 5 {
		t.Errorf("a loyalty counter per life gain: %d, want 5", effectstest.LoyaltyOf(g, ajani))
	}

}

func TestRFAAjaniResolutesMinusFourMakesAGrowingPridemate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ajani := effectstest.PushCatalogPermanent(g, me.ID, "Ajani Resolute", "Legendary Planeswalker — Ajani", effectstest.RfaAjaniResolute, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ajani, game.CounterLoyalty, 5) })
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, ajani, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ajani's Pridemate" && c.Controller == me.ID {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("−4 makes Ajani's Pridemate")
	}
	if effectstest.EffectivePower(t, g, token) != 2 {
		t.Errorf("a 2/2: %d", effectstest.EffectivePower(t, g, token))
	}
	// A real life gain (his own 0, next turn) triggers both his counter and the token.
	effectstest.B39NextTurnOf(t, g, 0)
	effectstest.B16Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TriggerOrderPrompt(g) != nil {
		effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if effectstest.LoyaltyOf(g, ajani) != 2 {
		t.Errorf("his own counter too: %d", effectstest.LoyaltyOf(g, ajani))
	}
	if effectstest.CounterOn(g, token, game.CounterPlusOne) != 1 {
		t.Errorf("one +1/+1 counter per gain: %d", effectstest.CounterOn(g, token, game.CounterPlusOne))
	}
}

func TestRFAAjaniResolutesEmblemIsAnAnthem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ajani := effectstest.PushCatalogPermanent(g, me.ID, "Ajani Resolute", "Legendary Planeswalker — Ajani", effectstest.RfaAjaniResolute, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ajani, game.CounterLoyalty, 10) })
	bear := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.B39Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	before := effectstest.EffectivePower(t, g, bear)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, ajani, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectivePower(t, g, bear) != before+2 {
		t.Errorf("+2/+2 for your creatures: %d → %d", before, effectstest.EffectivePower(t, g, bear))
	}
	if effectstest.EffectivePower(t, g, theirs) != 2 {
		t.Error("not for theirs")
	}
	if me.Emblems.Size() != 1 {
		t.Errorf("one emblem: %d", me.Emblems.Size())
	}
}

func TestRFAAjaniUnrelentingMakesACadetPerLoyaltyActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ajani := effectstest.PushCatalogPermanent(g, me.ID, "Ajani Unrelenting", "Legendary Planeswalker — Ajani", effectstest.RfaAjaniUnrelenting, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ajani, game.CounterLoyalty, 5) })
	other := effectstest.PushCatalogPermanent(g, me.ID, "Ajani, the Greathearted", "Legendary Planeswalker — Ajani", effectstest.B39AjaniOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(other, game.CounterLoyalty, 5) })
	effectstest.AdvanceToMain(t, g)

	effectstest.B16Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Cadet"); n != 1 {
		t.Fatalf("a Cadet for his own activation: %d", n)
	}
	// Another planeswalker's loyalty ability triggers him too.
	effectstest.B16Activate(t, g, me.ID, other, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Cadet"); n != 2 {
		t.Errorf("and for another walker's: %d", n)
	}
}

func TestRFAAjaniUnrelentingPlusOnePumpsCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ajani := effectstest.PushCatalogPermanent(g, me.ID, "Ajani Unrelenting", "Legendary Planeswalker — Ajani", effectstest.RfaAjaniUnrelenting, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ajani, game.CounterLoyalty, 5) })
	bear := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.B39Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	before := effectstest.EffectivePower(t, g, bear)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectivePower(t, g, bear) != before+1 || !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Errorf("+1/+0 and haste: power %d → %d", before, effectstest.EffectivePower(t, g, bear))
	}
	if effectstest.EffectivePower(t, g, theirs) != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature gets nothing")
	}
}

func TestRFAAjaniUnrelentingMinusTwoDiscardsThenDrawsPerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ajani := effectstest.PushCatalogPermanent(g, me.ID, "Ajani Unrelenting", "Legendary Planeswalker — Ajani", effectstest.RfaAjaniUnrelenting, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ajani, game.CounterLoyalty, 5) })
	effectstest.SeedBear(g, me.ID)
	effectstest.SeedBear(g, me.ID)
	for i := 0; i < 8; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	effectstest.AdvanceToMain(t, g)
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Hand A", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Hand B", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	graveBefore := len(me.Graveyard.Cards)

	effectstest.B16Activate(t, g, me.ID, ajani, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Graveyard.Cards) < graveBefore+2 {
		t.Errorf("the hand was discarded: graveyard %d → %d", graveBefore, len(me.Graveyard.Cards))
	}
	// Two Bears plus the Cadet his own activation made.
	if me.Hand.Size() != 3 {
		t.Errorf("drew a card per creature (2 Bears + the Cadet): hand %d, want 3", me.Hand.Size())
	}
}

func TestRFAAjaniUnrelentingMinusThreeSparesYourTokensOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ajani := effectstest.PushCatalogPermanent(g, me.ID, "Ajani Unrelenting", "Legendary Planeswalker — Ajani", effectstest.RfaAjaniUnrelenting, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ajani, game.CounterLoyalty, 5) })
	myTok := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Token", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	myCard := effectstest.B39Creature(g, me.ID, "My Card", "Creature — Bear", 2, 2)
	theirTok := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Token", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	theirCard := effectstest.B39Creature(g, opp.ID, "Their Card", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, ajani, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chandra := effectstest.PushCatalogPermanent(g, me.ID, "Chandra, Chill of Compliance", "Legendary Planeswalker — Chandra", effectstest.RfaChandra, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(chandra, game.CounterLoyalty, 3) })
	spell := rfaTopOfLibrary(me, "Opt", "Instant")
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("a surveil prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(spell) {
		t.Error("the binned instant goes to hand")
	}
	if effectstest.LoyaltyOf(g, chandra) != 4 {
		t.Errorf("+1 loyalty: %d", effectstest.LoyaltyOf(g, chandra))
	}
}

func TestRFAChandraFirstPlusOneLeavesACreatureInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chandra := effectstest.PushCatalogPermanent(g, me.ID, "Chandra, Chill of Compliance", "Legendary Planeswalker — Chandra", effectstest.RfaChandra, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(chandra, game.CounterLoyalty, 3) })
	bear := rfaTopOfLibrary(me, "Bear", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("a surveil prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(bear) || !me.Graveyard.Contains(bear) {
		t.Error("a creature card stays in the graveyard")
	}
}

func TestRFAChandraFirstPlusOneKeepingTheCardOnTopReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chandra := effectstest.PushCatalogPermanent(g, me.ID, "Chandra, Chill of Compliance", "Legendary Planeswalker — Chandra", effectstest.RfaChandra, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(chandra, game.CounterLoyalty, 3) })
	spell := rfaTopOfLibrary(me, "Opt", "Instant")
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("a surveil prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("kept on top: neither in hand nor in the graveyard")
	}
}

func TestRFAChandraSecondPlusOneAddsRestrictedBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chandra := effectstest.PushCatalogPermanent(g, me.ID, "Chandra, Chill of Compliance", "Legendary Planeswalker — Chandra", effectstest.RfaChandra, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(chandra, game.CounterLoyalty, 3) })
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, chandra, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var toks []game.ManaToken
	g.ReadSnapshot(func() { toks = append(toks, me.ManaPool...) })
	if len(toks) != 1 || toks[0].Color != "U" {
		t.Fatalf("one blue mana: %+v", toks)
	}
	if len(toks[0].Restrictions) != 2 {
		t.Errorf("restricted to casting a noncreature spell: %v", toks[0].Restrictions)
	}
	if effectstest.LoyaltyOf(g, chandra) != 4 {
		t.Errorf("+1 loyalty: %d", effectstest.LoyaltyOf(g, chandra))
	}
}

func TestRFAChandraEmblemDrawsOnEverySpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chandra := effectstest.PushCatalogPermanent(g, me.ID, "Chandra, Chill of Compliance", "Legendary Planeswalker — Chandra", effectstest.RfaChandra, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(chandra, game.CounterLoyalty, 6) })
	for i := 0; i < 6; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, chandra, 3, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Emblems.Size() != 1 {
		t.Fatalf("one emblem: %d", me.Emblems.Size())
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Some Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("whenever you cast a spell, draw a card: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestRFAEdgarGainsLifeWhenAnotherCreatureOrPlaneswalkerOfYoursDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Edgar, Ancient Bloodlord", "Legendary Creature — Vampire Noble", effectstest.RfaEdgar, false)
	bear := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.B39Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	walker := effectstest.PushWalkerForTest(g, me.ID, "Walker", "", 3)

	life := me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("your creature dies: %d → %d", life, me.Life)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("an opponent's creature dying gains nothing: %d", me.Life)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(walker) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("your planeswalker dies: %d, want %d", me.Life, life+2)
	}
}

func TestRFAEdgarSacrificeGrowsAndGainsMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	edgar := effectstest.PushCatalogPermanent(g, me.ID, "Edgar, Ancient Bloodlord", "Legendary Creature — Vampire Noble", effectstest.RfaEdgar, false)
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool = append(me.ManaPool, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}) })
	effectstest.B16Activate(t, g, me.ID, edgar, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if rfaOnBattlefield(g, bear) {
		t.Error("the Bear was sacrificed")
	}
	if effectstest.CounterOn(g, edgar, game.CounterPlusOne) != 1 {
		t.Error("a +1/+1 counter on Edgar")
	}
	if !effectstest.HasEffectiveKeyword(t, g, edgar, "menace") {
		t.Error("menace until end of turn")
	}
}

func TestRFACraftworkCrusherChoosesTwoModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.RfaColoredCreature(g, opp.ID, "Their Bear", 2, 2, "G")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Craftwork Crusher", "Artifact Creature — Boar Construct", effectstest.RfaCraftworkCrusher, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("the enters trigger asks for two modes: %+v", g.PendingChoices)
	}
	if c.ModeMin != 2 || c.ModeMax != 2 {
		t.Errorf("choose exactly two: %d–%d", c.ModeMin, c.ModeMax)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("the damage bullet asks for a target: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if rfaOnBattlefield(g, theirs) {
		t.Error("4 damage kills the Bear")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("and draws a card: hand %d → %d", hand, me.Hand.Size())
	}
	if effectstest.B16CountNamed(g, "Cadet") != 0 {
		t.Error("the Cadet bullet was not chosen")
	}
}

func TestRFACraftworkCrusherCadetAndDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Craftwork Crusher", "Artifact Creature — Boar Construct", effectstest.RfaCraftworkCrusher, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("the enters trigger asks for two modes: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1, 2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16CountNamed(g, "Cadet") != 1 || me.Hand.Size() != hand+1 {
		t.Errorf("a Cadet and a card: cadets %d, hand %d → %d", effectstest.B16CountNamed(g, "Cadet"), hand, me.Hand.Size())
	}
}
