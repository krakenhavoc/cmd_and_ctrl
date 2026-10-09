package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_fra_creature_b_more_test.go — the second half of the
// fra-creature-b card tests (see reality_fracture_fra_creature_b_test.go).

func rfcbOnBattlefield(g *game.Game, id uuid.UUID) bool {
	_, ok := battlefieldCard(g, id)
	return ok
}

// --- Eardrum Rattler --------------------------------------------------

func TestEardrumRattlerMakesAnotherSmallCreatureUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rattler := rfcbPush(g, me.ID, "Eardrum Rattler", "Creature — Human Bard", rfcbEardrumRattler, 2, 2)
	small := pushRestrictionBear(g, me, "Small Beast")
	big := rfcbPush(g, me.ID, "Big Beast", "Creature — Beast", "", 3, 3)
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	for name, target := range map[string]uuid.UUID{"itself": rattler, "a power-3 creature": big} {
		err := g.ActivateCatalogAbility(me.ID, rattler, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		})
		if !errors.Is(err, game.ErrIllegalTarget) {
			t.Errorf("targeting %s: err = %v, want ErrIllegalTarget", name, err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, rattler, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, small, game.CantBeBlocked)
	assertRestrictions(t, g, big, 0)
}

// --- Divining Duelist -------------------------------------------------

func rfcbDuelistEnters(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	castAndResolveCreature(t, g, "Divining Duelist", "Creature — Merfolk Wizard", rfcbDiviningDuelist)
	mode := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt from the enters trigger: %+v", g.PendingChoices)
	}
	return mode
}

func TestDiviningDuelistTapsAndUntapsATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	theirs := seedCreature(g, "Theirs", opp.ID)

	mode := rfcbDuelistEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("choosing tap: %v", err)
	}
	if err := sgAnswerTargets(t, g, theirs); err != nil {
		t.Fatalf("target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tapped, _ := battlefieldCardTapped(g, theirs); !tapped {
		t.Error("the targeted creature should be tapped")
	}

	mode = rfcbDuelistEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("choosing untap: %v", err)
	}
	if err := sgAnswerTargets(t, g, theirs); err != nil {
		t.Fatalf("target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tapped, _ := battlefieldCardTapped(g, theirs); tapped {
		t.Error("the targeted creature should be untapped")
	}
}

func TestDiviningDuelistLootsWithNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	mode := rfcbDuelistEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{2}); err != nil {
		t.Fatalf("choosing the loot bullet needs no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); got == nil {
		t.Error("the loot should open a discard prompt after drawing")
	}
	if me.Hand.Size() <= before {
		t.Errorf("hand = %d, want it to have grown by the draw (was %d)", me.Hand.Size(), before)
	}
}

// --- Fateshaper Aspirant ----------------------------------------------

func rfcbAspirantEnters(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	castAndResolveCreature(t, g, "Fateshaper Aspirant", "Creature — Rhino Cleric", rfcbFateshaperAspirant)
	mode := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt from the enters trigger: %+v", g.PendingChoices)
	}
	return mode
}

func TestFateshaperAspirantReturnsALegendaryCardAndRefusesOthers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	legend := batch01GraveyardCard(me, "Legend", "Legendary Artifact")
	plain := batch01GraveyardCard(me, "Plain", "Creature — Bear")
	mode := rfcbAspirantEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("choosing the return bullet: %v", err)
	}
	if err := sgAnswerTargets(t, g, plain); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a nonlegendary card: err = %v, want ErrIllegalTarget", err)
	}
	if err := sgAnswerTargets(t, g, legend); err != nil {
		t.Fatalf("a legendary card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(legend) {
		t.Error("the legendary card should be in hand")
	}
	if !me.Graveyard.Contains(plain) {
		t.Error("the other card stays in the graveyard")
	}
}

func TestFateshaperAspirantCountersAndProtectsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	theirs := seedCreature(g, "Theirs", opp.ID)
	mode := rfcbAspirantEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("choosing the counter bullet: %v", err)
	}
	if err := sgAnswerTargets(t, g, theirs); err != nil {
		t.Fatalf("target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := rfcbCounters(t, g, theirs); got != 1 {
		t.Errorf("counters = %d, want 1", got)
	}
	abilities := effectiveAbilities(t, g, theirs)
	if !hasAbility(abilities, "vigilance") || !hasAbility(abilities, "indestructible") {
		t.Errorf("abilities = %v, want vigilance and indestructible", abilities)
	}
}

// --- Flickering Hound --------------------------------------------------

func TestFlickeringHoundFlickersAnotherCreatureWhenYouCastACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hound := rfcbPush(g, me.ID, "Flickering Hound", "Creature — Dog", rfcbFlickeringHound, 2, 2)
	other := rfcbPush(g, me.ID, "Flickered Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "New Bear", "Creature — Bear", "", nil)
	if err := sgAnswerTargets(t, g, hound); err == nil {
		t.Fatal("the Hound must not be able to target itself")
	}
	if err := sgAnswerTargets(t, g, other); err != nil {
		t.Fatalf("target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if rfcbOnBattlefield(g, other) {
		t.Error("the original object should be gone (a new object returned)")
	}
	if got := b16CountNamed(g, "Flickered Bear"); got != 1 {
		t.Errorf("Flickered Bear copies on the battlefield = %d, want 1", got)
	}
	if !rfcbOnBattlefield(g, hound) {
		t.Error("the Hound stays")
	}
}

func TestFlickeringHoundIgnoresNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rfcbPush(g, me.ID, "Flickering Hound", "Creature — Dog", rfcbFlickeringHound, 2, 2)
	other := rfcbPush(g, me.ID, "Flickered Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Cantrip", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if !rfcbOnBattlefield(g, other) {
		t.Error("a noncreature spell must not trigger the Hound")
	}
}

// --- Graft Surgeon ------------------------------------------------------

func TestGraftSurgeonEntersWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	id := castAndResolveCreature(t, g, "Graft Surgeon", "Creature — Human Cleric", rfcbGraftSurgeon)
	passPriorityAroundTable(t, g)
	if got := rfcbCounters(t, g, id); got != 1 {
		t.Errorf("counters = %d, want 1", got)
	}
}

func TestGraftSurgeonMovesEveryKindOfCounterWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	surgeon := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Graft Surgeon", TypeLine: "Creature — Human Cleric", OracleID: rfcbGraftSurgeon,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 1, "charge": 2},
	})
	heir := rfcbPush(g, me.ID, "Heir", "Creature — Bear", "", 2, 2)
	destroy(t, g, surgeon)
	if err := sgAnswerTargets(t, g, heir); err != nil {
		t.Fatalf("target: %v", err)
	}
	passPriorityAroundTable(t, g)
	got := allCountersOn(t, g, heir)
	if got[game.CounterPlusOne] != 1 || got["charge"] != 2 {
		t.Errorf("counters on the heir = %v, want +1/+1 x1 and charge x2", got)
	}
}

func TestGraftSurgeonWithNoOtherCreatureJustDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	surgeon := rfcbPush(g, me.ID, "Graft Surgeon", "Creature — Human Cleric", rfcbGraftSurgeon, 2, 2)
	destroy(t, g, surgeon)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(surgeon) {
		t.Error("the Surgeon should be in the graveyard")
	}
}

// --- Gallia, Tragic Host -------------------------------------------------

func TestGalliaTragicHostReturnsFromTheGraveyardWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	food := pushCatalogGraveyardCard(me, "Fodder", "Creature — Bear", "", 2, 2)
	id, _ := activateFromGraveyard(t, g, "Gallia, Tragic Host", "Legendary Creature — Zombie Satyr",
		rfcbGalliaTragicHost, 2, 1, "{B}{C}{C}{C}{C}", game.ActivateAbilityParams{ExileIDs: []uuid.UUID{food}})
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Gallia did not return to the battlefield")
	}
	if !c.Tapped {
		t.Error("Gallia returns tapped")
	}
	if got := rfcbCounters(t, g, id); got != 1 {
		t.Errorf("counters = %d, want 1", got)
	}
	if me.Graveyard.Contains(food) {
		t.Error("the creature card should have been exiled as the cost")
	}
}

func TestGalliaTragicHostCastFromHandHasNoCounter(t *testing.T) {
	g := newCatalogGame(t)
	id := castAndResolveCreature(t, g, "Gallia, Tragic Host", "Legendary Creature — Zombie Satyr", rfcbGalliaTragicHost)
	passPriorityAroundTable(t, g)
	if got := rfcbCounters(t, g, id); got != 0 {
		t.Errorf("counters = %d, want 0 for a hard cast", got)
	}
}

func TestGalliaTragicHostNeedsAnotherCreatureCardToExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
	id := pushCatalogGraveyardCard(me, "Gallia, Tragic Host", "Legendary Creature — Zombie Satyr", rfcbGalliaTragicHost, 2, 1)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{id}}); err == nil {
		t.Error("Gallia cannot pay the cost by exiling herself")
	}
}

// --- Gallia, the Merrymaker -----------------------------------------------

func TestGalliaTheMerrymakerGivesHasteToCreaturesWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gallia := rfcbPush(g, me.ID, "Gallia, the Merrymaker", "Legendary Creature — Satyr", rfcbGalliaMerrymaker, 2, 1)
	with := seedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	without := seedCreature(g, "Plain", me.ID)
	if !hasAbility(effectiveAbilities(t, g, with), "haste") {
		t.Error("a creature with a +1/+1 counter should have haste")
	}
	if hasAbility(effectiveAbilities(t, g, without), "haste") {
		t.Error("a creature with no counter must not gain haste")
	}
	if !hasAbility(effectiveAbilities(t, g, gallia), "haste") {
		t.Error("Gallia has haste as a printed keyword")
	}
}

func TestGalliaTheMerrymakerCountersOnlyCreaturesThatEnteredThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gallia := rfcbPush(g, me.ID, "Gallia, the Merrymaker", "Legendary Creature — Satyr", rfcbGalliaMerrymaker, 2, 1)
	old := rfcbPush(g, me.ID, "Old Bear", "Creature — Bear", "", 2, 2)
	fresh := castAndResolveCreature(t, g, "Fresh Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	err := g.ActivateCatalogAbility(me.ID, gallia, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: old}},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a creature that did not enter this turn: err = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, gallia, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: fresh}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := rfcbCounters(t, g, fresh); got != 1 {
		t.Errorf("counters = %d, want 1", got)
	}
}

// --- Edgar, Moonlit Sovereign ------------------------------------------------

func TestEdgarGrowsAtYourEndStepOnlyIfYouCastNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	edgar := rfcbPush(g, me.ID, "Edgar, Moonlit Sovereign", "Legendary Creature — Werewolf Noble", rfcbEdgar, 4, 4)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := rfcbCounters(t, g, edgar); got != 2 {
		t.Errorf("counters = %d, want 2 after a turn with no spell", got)
	}
}

func TestEdgarStaysPutAfterYouCastASpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	edgar := rfcbPush(g, me.ID, "Edgar, Moonlit Sovereign", "Legendary Creature — Werewolf Noble", rfcbEdgar, 4, 4)
	castCatalogSpell(t, g, "Cantrip", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := rfcbCounters(t, g, edgar); got != 0 {
		t.Errorf("counters = %d, want 0 after casting a spell", got)
	}
}

func TestEdgarPutsACounterOnEachCreatureThatHasOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	edgar := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Edgar, Moonlit Sovereign", TypeLine: "Legendary Creature — Werewolf Noble",
		OracleID: rfcbEdgar, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 2},
	})
	with := seedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	without := seedCreature(g, "Plain", me.ID)
	theirs := seedCreatureWithCounters(g, opp.ID, map[string]int{game.CounterPlusOne: 1})
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, edgar, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := rfcbCounters(t, g, edgar); got != 3 {
		t.Errorf("Edgar counters = %d, want 3", got)
	}
	if got := rfcbCounters(t, g, with); got != 2 {
		t.Errorf("my countered creature = %d, want 2", got)
	}
	if got := rfcbCounters(t, g, without); got != 0 {
		t.Errorf("my plain creature = %d, want 0", got)
	}
	if got := rfcbCounters(t, g, theirs); got != 1 {
		t.Errorf("an opponent's creature = %d, want 1 (unchanged)", got)
	}
}

// --- Desperate Futurescribe ------------------------------------------------------

func TestDesperateFuturescribePumpsAnotherCreatureAtBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	scribe := rfcbPush(g, me.ID, "Desperate Futurescribe", "Creature — Kor Scout", rfcbDesperateFuturescribe, 3, 4)
	other := rfcbPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	advanceToStepInTurn(t, g, game.StepBeginCombat)
	if err := sgAnswerTargets(t, g, scribe); err == nil {
		t.Fatal("the Futurescribe must not be able to target itself")
	}
	if err := sgAnswerTargets(t, g, other); err != nil {
		t.Fatalf("target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, th := rfPT(t, g, other); p != 3 || th != 3 {
		t.Errorf("P/T = %d/%d, want 3/3 until end of turn", p, th)
	}
	if got := rfcbCounters(t, g, other); got != 0 {
		t.Errorf("counters = %d, want 0 with no scry or surveil", got)
	}
}

func TestDesperateFuturescribePutsACounterAfterYouScried(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rfcbPush(g, me.ID, "Desperate Futurescribe", "Creature — Kor Scout", rfcbDesperateFuturescribe, 3, 4)
	other := rfcbPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventSurveil, Actor: me.ID, LookedAt: 1})
	})
	advanceToStepInTurn(t, g, game.StepBeginCombat)
	if err := sgAnswerTargets(t, g, other); err != nil {
		t.Fatalf("target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := rfcbCounters(t, g, other); got != 1 {
		t.Errorf("counters = %d, want 1 after a surveil", got)
	}
}

// --- Draconic Visitor ----------------------------------------------------------

func TestDraconicVisitorTurnsYourArtifactTokensIntoDragons(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	rfcbPush(g, me.ID, "Draconic Visitor", "Creature — Dragon", rfcbDraconicVisitor, 5, 5)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, TreasureToken(), 2); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
		if err := g.CreateTokenForEffect(me.ID, TokenCard("1/1 white Soldier"), 1); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
		if err := g.CreateTokenForEffect(opp.ID, TreasureToken(), 1); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
	})
	if got := b16CountNamed(g, "Dragon"); got != 2 {
		t.Errorf("Dragons = %d, want 2 (one per Treasure replaced)", got)
	}
	if got := b16CountNamed(g, "Treasure"); got != 1 {
		t.Errorf("Treasures = %d, want only the opponent's", got)
	}
	if got := b16CountNamed(g, "Soldier"); got != 1 {
		t.Errorf("Soldiers = %d, want 1 (not an artifact)", got)
	}
}

// --- Fblthp, Knows the Way -------------------------------------------------------

func TestFblthpKnowsTheWayPowerIsTheNumberOfBasicLandTypes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfcbPush(g, me.ID, "Fblthp, Knows the Way", "Legendary Creature — Homunculus Scout", rfcbFblthpKnowsTheWay, 0, 2)
	if got := effectivePower(t, g, id); got != 0 {
		t.Errorf("power with no lands = %d, want 0", got)
	}
	for _, l := range []struct {
		owner uuid.UUID
		sub   string
	}{{me.ID, "Forest"}, {me.ID, "Forest"}, {me.ID, "Island"}, {g.Seats[1].ID, "Swamp"}} {
		rfcbPush(g, l.owner, l.sub, "Basic Land — "+l.sub, "", 0, 0)
	}
	if got := effectivePower(t, g, id); got != 2 {
		t.Errorf("power = %d, want 2 (Forest and Island; two Forests count once, theirs doesn't)", got)
	}
}

func TestFblthpKnowsTheWaySearchesForXBasicsWithDifferentNames(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Mountain", "Basic Land — Mountain"),
	)
	toMainForCost(t, g)
	id := handCardFull(me, "Fblthp, Knows the Way", "Legendary Creature — Homunculus Scout", "{X}{G}{G}", rfcbFblthpKnowsTheWay, []string{"G"})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0], ids[1]}); err == nil {
		t.Fatal("two Islands share a name and must be refused")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0], ids[2]}); err != nil {
		t.Fatalf("Island and Forest: %v", err)
	}
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[2]) {
		t.Error("both basics should be in hand")
	}
}

// --- Fblthp, Impossibly Lost ------------------------------------------------------

func rfcbFblthpHits(t *testing.T, g *game.Game, me, opp *game.Player, fblthp uuid.UUID, combat bool) {
	t.Helper()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: fblthp, Target: opp.ID, Amount: 1, Combat: combat, Actor: me.ID})
	})
	passPriorityAroundTable(t, g)
}

func TestFblthpImpossiblyLostDrawsTwoAndShufflesItselfIn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	fblthp := rfcbPush(g, me.ID, "Fblthp, Impossibly Lost", "Legendary Creature — Homunculus", rfcbFblthpImpossiblyLost, 1, 1)
	hand, lib := me.Hand.Size(), me.Library.Size()
	rfcbFblthpHits(t, g, me, opp, fblthp, true)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand = %d, want %d", got, hand+2)
	}
	if g.State != game.StateActive {
		t.Fatalf("the game ended with %d cards left in the library", me.Library.Size())
	}
	if rfcbOnBattlefield(g, fblthp) {
		t.Error("Fblthp should have been shuffled into the library")
	}
	if got := me.Library.Size(); got != lib-2+1 {
		t.Errorf("library = %d, want %d", got, lib-1)
	}
}

func TestFblthpImpossiblyLostIgnoresNoncombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	fblthp := rfcbPush(g, me.ID, "Fblthp, Impossibly Lost", "Legendary Creature — Homunculus", rfcbFblthpImpossiblyLost, 1, 1)
	hand := me.Hand.Size()
	rfcbFblthpHits(t, g, me, opp, fblthp, false)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand = %d, want %d: noncombat damage does not trigger it", got, hand)
	}
}

func TestFblthpImpossiblyLostIgnoresDamageToYouAndOnOtherTurns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	fblthp := rfcbPush(g, me.ID, "Fblthp, Impossibly Lost", "Legendary Creature — Homunculus", rfcbFblthpImpossiblyLost, 1, 1)
	hand := me.Hand.Size()
	rfcbFblthpHits(t, g, me, me, fblthp, true) // damage to ME is not "your opponents"
	if me.Hand.Size() != hand {
		t.Error("combat damage to its own controller must not trigger it")
	}
	// An opponent's turn: the trigger needs it to be MY turn.
	g.Turn.ActiveSeat = (g.Turn.ActiveSeat + 1) % 4
	rfcbFblthpHits(t, g, me, opp, fblthp, true)
	if me.Hand.Size() != hand {
		t.Error("damage dealt on another player's turn must not trigger it")
	}
}

func TestFblthpImpossiblyLostWinsOnAnEmptyLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	fblthp := rfcbPush(g, me.ID, "Fblthp, Impossibly Lost", "Legendary Creature — Homunculus", rfcbFblthpImpossiblyLost, 1, 1)
	for me.Library.Size() > 1 {
		_, _ = me.Library.PopTop()
	}
	rfcbFblthpHits(t, g, me, opp, fblthp, true)
	if g.State == game.StateActive {
		t.Error("drawing the library out should win the game")
	}
	if me.Eliminated {
		t.Error("the controller must not lose for drawing from an empty library")
	}
}

// --- Ginger, Queen of Sweets ---------------------------------------------------------

func TestGingerMakesYouMonarchAndMakesGingerbrutesAtEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castAndResolveCreature(t, g, "Ginger, Queen of Sweets", "Legendary Artifact Creature — Food Noble", rfcbGinger)
	passPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("monarch = %v, want me", g.Monarch)
	}
	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%4)
	passPriorityAroundTable(t, g)
	if got := b16CountNamed(g, "Gingerbrute"); got != 1 {
		t.Errorf("Gingerbrutes after an opponent's upkeep = %d, want 1", got)
	}
}

func TestGingerMakesNoTokenWhenYouAreNotTheMonarch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	rfcbPush(g, me.ID, "Ginger, Queen of Sweets", "Legendary Artifact Creature — Food Noble", rfcbGinger, 6, 4)
	g.WithWriteLock(func() {
		if err := g.SetMonarchForEffect(opp.ID); err != nil {
			t.Fatalf("SetMonarchForEffect: %v", err)
		}
	})
	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%4)
	passPriorityAroundTable(t, g)
	if got := b16CountNamed(g, "Gingerbrute"); got != 0 {
		t.Errorf("Gingerbrutes = %d, want 0 without the crown", got)
	}
}

func TestGingerSacrificesForSixLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ginger := rfcbPush(g, me.ID, "Ginger, Queen of Sweets", "Legendary Artifact Creature — Food Noble", rfcbGinger, 6, 4)
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	before := me.Life
	if err := g.ActivateCatalogAbility(me.ID, ginger, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+6 {
		t.Errorf("life = %d, want %d", me.Life, before+6)
	}
	if rfcbOnBattlefield(g, ginger) {
		t.Error("Ginger is sacrificed as a cost")
	}
}

func TestGingerbruteTokenHasBothAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, GingerbruteToken(), 1); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
	})
	var tok game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Gingerbrute" {
			tok = c
		}
	}
	if tok.InstanceID == uuid.Nil {
		t.Fatal("no Gingerbrute token")
	}
	if n := len(game.ActivatedAbilitiesForCard(tok)); n != 2 {
		t.Errorf("activated abilities = %d, want 2", n)
	}
	if !tok.HasSubtype("Food") || !tok.HasSubtype("Golem") || !tok.IsArtifact() || !tok.IsCreature() {
		t.Errorf("type line %q, want an artifact creature Food Golem", tok.TypeLine)
	}
	if !hasAbility(effectiveAbilities(t, g, tok.InstanceID), "haste") {
		t.Error("the token has haste")
	}
}

// --- Emrakul, the Exigent Doom -------------------------------------------------------

func TestEmrakulUntapsYourLandsWhenCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := pushPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest", Tapped: true})
	theirs := pushPermanent(g, opp.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island", Tapped: true})
	toMainForCost(t, g)
	id := handCardFull(me, "Emrakul, the Exigent Doom", "Legendary Creature — Eldrazi", "", rfcbEmrakul, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tapped, _ := battlefieldCardTapped(g, mine); tapped {
		t.Error("my land should be untapped by the cast trigger")
	}
	if tapped, _ := battlefieldCardTapped(g, theirs); !tapped {
		t.Error("an opponent's land must stay tapped")
	}
}

func TestEmrakulHasWardSacrificeThreePermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfcbPush(g, me.ID, "Emrakul, the Exigent Doom", "Legendary Creature — Eldrazi", rfcbEmrakul, 12, 12)
	abilities := effectiveAbilities(t, g, id)
	if !hasAbility(abilities, "flying") || !hasAbility(abilities, "trample") {
		t.Errorf("abilities = %v, want flying and trample", abilities)
	}
	if len(game.TriggersForCard(game.Card{OracleID: rfcbEmrakul})) < 2 {
		t.Error("expected the cast trigger and the ward trigger")
	}
}

// --- Grim Repriser ----------------------------------------------------------------------

func TestGrimRepriserReturnsWithFinalityOnlyAfterNoncombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
	id := pushCatalogGraveyardCard(me, "Grim Repriser", "Creature — Zombie Bard", rfcbGrimRepriser, 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activating with no damage dealt this turn must be refused")
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: uuid.New(), Target: opp.ID, Amount: 2, Actor: me.ID})
	})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("after noncombat damage: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !rfcbOnBattlefield(g, id) {
		t.Fatal("the Repriser did not return")
	}
	if got := allCountersOn(t, g, id)[rfCreatureBFinalityCounter]; got != 1 {
		t.Fatalf("finality counters = %d, want 1", got)
	}
	destroy(t, g, id)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(id) {
		t.Error("a creature with a finality counter is exiled instead of dying")
	}
}

func TestGrimRepriserDoesNotCountCombatDamageOrDamageToYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
	id := pushCatalogGraveyardCard(me, "Grim Repriser", "Creature — Zombie Bard", rfcbGrimRepriser, 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: uuid.New(), Target: opp.ID, Amount: 2, Combat: true, Actor: me.ID})
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: uuid.New(), Target: me.ID, Amount: 2, Actor: opp.ID})
	})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("combat damage and damage to yourself must not enable the ability")
	}
}

func TestGrimRepriserHardCastHasNoFinalityCounter(t *testing.T) {
	g := newCatalogGame(t)
	id := castAndResolveCreature(t, g, "Grim Repriser", "Creature — Zombie Bard", rfcbGrimRepriser)
	passPriorityAroundTable(t, g)
	if got := allCountersOn(t, g, id)[rfCreatureBFinalityCounter]; got != 0 {
		t.Errorf("finality counters = %d, want 0 for a hard cast", got)
	}
}
