package testsr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_fra_creature_b_more_test.go — the second half of the
// fra-creature-b card tests (see reality_fracture_fra_creature_b_test.go).

func rfcbOnBattlefield(g *game.Game, id uuid.UUID) bool {
	_, ok := effectstest.BattlefieldCard(g, id)
	return ok
}

// --- Eardrum Rattler --------------------------------------------------

func TestEardrumRattlerMakesAnotherSmallCreatureUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rattler := effectstest.RfcbPush(g, me.ID, "Eardrum Rattler", "Creature — Human Bard", effectstest.RfcbEardrumRattler, 2, 2)
	small := effectstest.PushRestrictionBear(g, me, "Small Beast")
	big := effectstest.RfcbPush(g, me.ID, "Big Beast", "Creature — Beast", "", 3, 3)
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
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
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, small, game.CantBeBlocked)
	effectstest.AssertRestrictions(t, g, big, 0)
}

// --- Divining Duelist -------------------------------------------------

func rfcbDuelistEnters(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	effectstest.CastAndResolveCreature(t, g, "Divining Duelist", "Creature — Merfolk Wizard", effectstest.RfcbDiviningDuelist)
	mode := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt from the enters trigger: %+v", g.PendingChoices)
	}
	return mode
}

func TestDiviningDuelistTapsAndUntapsATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	theirs := effectstest.SeedCreature(g, "Theirs", opp.ID)

	mode := rfcbDuelistEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("choosing tap: %v", err)
	}
	if err := effectstest.SgAnswerTargets(t, g, theirs); err != nil {
		t.Fatalf("target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, theirs); !tapped {
		t.Error("the targeted creature should be tapped")
	}

	mode = rfcbDuelistEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("choosing untap: %v", err)
	}
	if err := effectstest.SgAnswerTargets(t, g, theirs); err != nil {
		t.Fatalf("target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, theirs); tapped {
		t.Error("the targeted creature should be untapped")
	}
}

func TestDiviningDuelistLootsWithNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	mode := rfcbDuelistEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{2}); err != nil {
		t.Fatalf("choosing the loot bullet needs no target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); got == nil {
		t.Error("the loot should open a discard prompt after drawing")
	}
	if me.Hand.Size() <= before {
		t.Errorf("hand = %d, want it to have grown by the draw (was %d)", me.Hand.Size(), before)
	}
}

// --- Fateshaper Aspirant ----------------------------------------------

func rfcbAspirantEnters(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	effectstest.CastAndResolveCreature(t, g, "Fateshaper Aspirant", "Creature — Rhino Cleric", effectstest.RfcbFateshaperAspirant)
	mode := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt from the enters trigger: %+v", g.PendingChoices)
	}
	return mode
}

func TestFateshaperAspirantReturnsALegendaryCardAndRefusesOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	legend := effectstest.Batch01GraveyardCard(me, "Legend", "Legendary Artifact")
	plain := effectstest.Batch01GraveyardCard(me, "Plain", "Creature — Bear")
	mode := rfcbAspirantEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("choosing the return bullet: %v", err)
	}
	if err := effectstest.SgAnswerTargets(t, g, plain); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a nonlegendary card: err = %v, want ErrIllegalTarget", err)
	}
	if err := effectstest.SgAnswerTargets(t, g, legend); err != nil {
		t.Fatalf("a legendary card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(legend) {
		t.Error("the legendary card should be in hand")
	}
	if !me.Graveyard.Contains(plain) {
		t.Error("the other card stays in the graveyard")
	}
}

func TestFateshaperAspirantCountersAndProtectsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	theirs := effectstest.SeedCreature(g, "Theirs", opp.ID)
	mode := rfcbAspirantEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("choosing the counter bullet: %v", err)
	}
	if err := effectstest.SgAnswerTargets(t, g, theirs); err != nil {
		t.Fatalf("target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := rfcbCounters(t, g, theirs); got != 1 {
		t.Errorf("counters = %d, want 1", got)
	}
	abilities := effectstest.EffectiveAbilities(t, g, theirs)
	if !effectstest.HasAbility(abilities, "vigilance") || !effectstest.HasAbility(abilities, "indestructible") {
		t.Errorf("abilities = %v, want vigilance and indestructible", abilities)
	}
}

// --- Flickering Hound --------------------------------------------------

func TestFlickeringHoundFlickersAnotherCreatureWhenYouCastACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hound := effectstest.RfcbPush(g, me.ID, "Flickering Hound", "Creature — Dog", effectstest.RfcbFlickeringHound, 2, 2)
	other := effectstest.RfcbPush(g, me.ID, "Flickered Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "New Bear", "Creature — Bear", "", nil)
	if err := effectstest.SgAnswerTargets(t, g, hound); err == nil {
		t.Fatal("the Hound must not be able to target itself")
	}
	if err := effectstest.SgAnswerTargets(t, g, other); err != nil {
		t.Fatalf("target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if rfcbOnBattlefield(g, other) {
		t.Error("the original object should be gone (a new object returned)")
	}
	if got := effectstest.B16CountNamed(g, "Flickered Bear"); got != 1 {
		t.Errorf("Flickered Bear copies on the battlefield = %d, want 1", got)
	}
	if !rfcbOnBattlefield(g, hound) {
		t.Error("the Hound stays")
	}
}

func TestFlickeringHoundIgnoresNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.RfcbPush(g, me.ID, "Flickering Hound", "Creature — Dog", effectstest.RfcbFlickeringHound, 2, 2)
	other := effectstest.RfcbPush(g, me.ID, "Flickered Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Cantrip", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !rfcbOnBattlefield(g, other) {
		t.Error("a noncreature spell must not trigger the Hound")
	}
}

// --- Graft Surgeon ------------------------------------------------------

func TestGraftSurgeonEntersWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastAndResolveCreature(t, g, "Graft Surgeon", "Creature — Human Cleric", effectstest.RfcbGraftSurgeon)
	effectstest.PassPriorityAroundTable(t, g)
	if got := rfcbCounters(t, g, id); got != 1 {
		t.Errorf("counters = %d, want 1", got)
	}
}

func TestGraftSurgeonMovesEveryKindOfCounterWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	surgeon := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Graft Surgeon", TypeLine: "Creature — Human Cleric", OracleID: effectstest.RfcbGraftSurgeon,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 1, "charge": 2},
	})
	heir := effectstest.RfcbPush(g, me.ID, "Heir", "Creature — Bear", "", 2, 2)
	effectstest.Destroy(t, g, surgeon)
	if err := effectstest.SgAnswerTargets(t, g, heir); err != nil {
		t.Fatalf("target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.AllCountersOn(t, g, heir)
	if got[game.CounterPlusOne] != 1 || got["charge"] != 2 {
		t.Errorf("counters on the heir = %v, want +1/+1 x1 and charge x2", got)
	}
}

func TestGraftSurgeonWithNoOtherCreatureJustDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	surgeon := effectstest.RfcbPush(g, me.ID, "Graft Surgeon", "Creature — Human Cleric", effectstest.RfcbGraftSurgeon, 2, 2)
	effectstest.Destroy(t, g, surgeon)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(surgeon) {
		t.Error("the Surgeon should be in the graveyard")
	}
}

// --- Gallia, Tragic Host -------------------------------------------------

func TestGalliaTragicHostReturnsFromTheGraveyardWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	food := effectstest.PushCatalogGraveyardCard(me, "Fodder", "Creature — Bear", "", 2, 2)
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Gallia, Tragic Host", "Legendary Creature — Zombie Satyr",
		effectstest.RfcbGalliaTragicHost, 2, 1, "{B}{C}{C}{C}{C}", game.ActivateAbilityParams{ExileIDs: []uuid.UUID{food}})
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, id)
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
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastAndResolveCreature(t, g, "Gallia, Tragic Host", "Legendary Creature — Zombie Satyr", effectstest.RfcbGalliaTragicHost)
	effectstest.PassPriorityAroundTable(t, g)
	if got := rfcbCounters(t, g, id); got != 0 {
		t.Errorf("counters = %d, want 0 for a hard cast", got)
	}
}

func TestGalliaTragicHostNeedsAnotherCreatureCardToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
	id := effectstest.PushCatalogGraveyardCard(me, "Gallia, Tragic Host", "Legendary Creature — Zombie Satyr", effectstest.RfcbGalliaTragicHost, 2, 1)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{id}}); err == nil {
		t.Error("Gallia cannot pay the cost by exiling herself")
	}
}

// --- Gallia, the Merrymaker -----------------------------------------------

func TestGalliaTheMerrymakerGivesHasteToCreaturesWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gallia := effectstest.RfcbPush(g, me.ID, "Gallia, the Merrymaker", "Legendary Creature — Satyr", effectstest.RfcbGalliaMerrymaker, 2, 1)
	with := effectstest.SeedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	without := effectstest.SeedCreature(g, "Plain", me.ID)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, with), "haste") {
		t.Error("a creature with a +1/+1 counter should have haste")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, without), "haste") {
		t.Error("a creature with no counter must not gain haste")
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, gallia), "haste") {
		t.Error("Gallia has haste as a printed keyword")
	}
}

func TestGalliaTheMerrymakerCountersOnlyCreaturesThatEnteredThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gallia := effectstest.RfcbPush(g, me.ID, "Gallia, the Merrymaker", "Legendary Creature — Satyr", effectstest.RfcbGalliaMerrymaker, 2, 1)
	old := effectstest.RfcbPush(g, me.ID, "Old Bear", "Creature — Bear", "", 2, 2)
	fresh := effectstest.CastAndResolveCreature(t, g, "Fresh Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.PassPriorityAroundTable(t, g)
	if got := rfcbCounters(t, g, fresh); got != 1 {
		t.Errorf("counters = %d, want 1", got)
	}
}

// --- Edgar, Moonlit Sovereign ------------------------------------------------

func TestEdgarGrowsAtYourEndStepOnlyIfYouCastNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	edgar := effectstest.RfcbPush(g, me.ID, "Edgar, Moonlit Sovereign", "Legendary Creature — Werewolf Noble", effectstest.RfcbEdgar, 4, 4)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := rfcbCounters(t, g, edgar); got != 2 {
		t.Errorf("counters = %d, want 2 after a turn with no spell", got)
	}
}

func TestEdgarStaysPutAfterYouCastASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	edgar := effectstest.RfcbPush(g, me.ID, "Edgar, Moonlit Sovereign", "Legendary Creature — Werewolf Noble", effectstest.RfcbEdgar, 4, 4)
	effectstest.CastCatalogSpell(t, g, "Cantrip", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := rfcbCounters(t, g, edgar); got != 0 {
		t.Errorf("counters = %d, want 0 after casting a spell", got)
	}
}

func TestEdgarPutsACounterOnEachCreatureThatHasOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	edgar := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Edgar, Moonlit Sovereign", TypeLine: "Legendary Creature — Werewolf Noble",
		OracleID: effectstest.RfcbEdgar, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 2},
	})
	with := effectstest.SeedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	without := effectstest.SeedCreature(g, "Plain", me.ID)
	theirs := effectstest.SeedCreatureWithCounters(g, opp.ID, map[string]int{game.CounterPlusOne: 1})
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, edgar, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	scribe := effectstest.RfcbPush(g, me.ID, "Desperate Futurescribe", "Creature — Kor Scout", effectstest.RfcbDesperateFuturescribe, 3, 4)
	other := effectstest.RfcbPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToStepInTurn(t, g, game.StepBeginCombat)
	if err := effectstest.SgAnswerTargets(t, g, scribe); err == nil {
		t.Fatal("the Futurescribe must not be able to target itself")
	}
	if err := effectstest.SgAnswerTargets(t, g, other); err != nil {
		t.Fatalf("target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.RfPT(t, g, other); p != 3 || th != 3 {
		t.Errorf("P/T = %d/%d, want 3/3 until end of turn", p, th)
	}
	if got := rfcbCounters(t, g, other); got != 0 {
		t.Errorf("counters = %d, want 0 with no scry or surveil", got)
	}
}

func TestDesperateFuturescribePutsACounterAfterYouScried(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.RfcbPush(g, me.ID, "Desperate Futurescribe", "Creature — Kor Scout", effectstest.RfcbDesperateFuturescribe, 3, 4)
	other := effectstest.RfcbPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventSurveil, Actor: me.ID, LookedAt: 1})
	})
	effectstest.AdvanceToStepInTurn(t, g, game.StepBeginCombat)
	if err := effectstest.SgAnswerTargets(t, g, other); err != nil {
		t.Fatalf("target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := rfcbCounters(t, g, other); got != 1 {
		t.Errorf("counters = %d, want 1 after a surveil", got)
	}
}

// --- Draconic Visitor ----------------------------------------------------------

func TestDraconicVisitorTurnsYourArtifactTokensIntoDragons(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.RfcbPush(g, me.ID, "Draconic Visitor", "Creature — Dragon", effectstest.RfcbDraconicVisitor, 5, 5)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 2); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
		if err := g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Soldier"), 1); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
		if err := g.CreateTokenForEffect(opp.ID, effects.TreasureToken(), 1); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
	})
	if got := effectstest.B16CountNamed(g, "Dragon"); got != 2 {
		t.Errorf("Dragons = %d, want 2 (one per Treasure replaced)", got)
	}
	if got := effectstest.B16CountNamed(g, "Treasure"); got != 1 {
		t.Errorf("Treasures = %d, want only the opponent's", got)
	}
	if got := effectstest.B16CountNamed(g, "Soldier"); got != 1 {
		t.Errorf("Soldiers = %d, want 1 (not an artifact)", got)
	}
}

// --- Fblthp, Knows the Way -------------------------------------------------------

func TestFblthpKnowsTheWayPowerIsTheNumberOfBasicLandTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.RfcbPush(g, me.ID, "Fblthp, Knows the Way", "Legendary Creature — Homunculus Scout", effectstest.RfcbFblthpKnowsTheWay, 0, 2)
	if got := effectstest.EffectivePower(t, g, id); got != 0 {
		t.Errorf("power with no lands = %d, want 0", got)
	}
	for _, l := range []struct {
		owner uuid.UUID
		sub   string
	}{{me.ID, "Forest"}, {me.ID, "Forest"}, {me.ID, "Island"}, {g.Seats[1].ID, "Swamp"}} {
		effectstest.RfcbPush(g, l.owner, l.sub, "Basic Land — "+l.sub, "", 0, 0)
	}
	if got := effectstest.EffectivePower(t, g, id); got != 2 {
		t.Errorf("power = %d, want 2 (Forest and Island; two Forests count once, theirs doesn't)", got)
	}
}

func TestFblthpKnowsTheWaySearchesForXBasicsWithDifferentNames(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Mountain", "Basic Land — Mountain"),
	)
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandCardFull(me, "Fblthp, Knows the Way", "Legendary Creature — Homunculus Scout", "{X}{G}{G}", effectstest.RfcbFblthpKnowsTheWay, []string{"G"})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
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
	effectstest.PassPriorityAroundTable(t, g)
}

func TestFblthpImpossiblyLostDrawsTwoAndShufflesItselfIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	fblthp := effectstest.RfcbPush(g, me.ID, "Fblthp, Impossibly Lost", "Legendary Creature — Homunculus", effectstest.RfcbFblthpImpossiblyLost, 1, 1)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	fblthp := effectstest.RfcbPush(g, me.ID, "Fblthp, Impossibly Lost", "Legendary Creature — Homunculus", effectstest.RfcbFblthpImpossiblyLost, 1, 1)
	hand := me.Hand.Size()
	rfcbFblthpHits(t, g, me, opp, fblthp, false)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand = %d, want %d: noncombat damage does not trigger it", got, hand)
	}
}

func TestFblthpImpossiblyLostIgnoresDamageToYouAndOnOtherTurns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	fblthp := effectstest.RfcbPush(g, me.ID, "Fblthp, Impossibly Lost", "Legendary Creature — Homunculus", effectstest.RfcbFblthpImpossiblyLost, 1, 1)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	fblthp := effectstest.RfcbPush(g, me.ID, "Fblthp, Impossibly Lost", "Legendary Creature — Homunculus", effectstest.RfcbFblthpImpossiblyLost, 1, 1)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastAndResolveCreature(t, g, "Ginger, Queen of Sweets", "Legendary Artifact Creature — Food Noble", effectstest.RfcbGinger)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("monarch = %v, want me", g.Monarch)
	}
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B16CountNamed(g, "Gingerbrute"); got != 1 {
		t.Errorf("Gingerbrutes after an opponent's upkeep = %d, want 1", got)
	}
}

func TestGingerMakesNoTokenWhenYouAreNotTheMonarch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.RfcbPush(g, me.ID, "Ginger, Queen of Sweets", "Legendary Artifact Creature — Food Noble", effectstest.RfcbGinger, 6, 4)
	g.WithWriteLock(func() {
		if err := g.SetMonarchForEffect(opp.ID); err != nil {
			t.Fatalf("SetMonarchForEffect: %v", err)
		}
	})
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B16CountNamed(g, "Gingerbrute"); got != 0 {
		t.Errorf("Gingerbrutes = %d, want 0 without the crown", got)
	}
}

func TestGingerSacrificesForSixLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ginger := effectstest.RfcbPush(g, me.ID, "Ginger, Queen of Sweets", "Legendary Artifact Creature — Food Noble", effectstest.RfcbGinger, 6, 4)
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	before := me.Life
	if err := g.ActivateCatalogAbility(me.ID, ginger, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+6 {
		t.Errorf("life = %d, want %d", me.Life, before+6)
	}
	if rfcbOnBattlefield(g, ginger) {
		t.Error("Ginger is sacrificed as a cost")
	}
}

func TestGingerbruteTokenHasBothAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.GingerbruteToken(), 1); err != nil {
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
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, tok.InstanceID), "haste") {
		t.Error("the token has haste")
	}
}

// --- Emrakul, the Exigent Doom -------------------------------------------------------

func TestEmrakulUntapsYourLandsWhenCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest", Tapped: true})
	theirs := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island", Tapped: true})
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandCardFull(me, "Emrakul, the Exigent Doom", "Legendary Creature — Eldrazi", "", effectstest.RfcbEmrakul, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, mine); tapped {
		t.Error("my land should be untapped by the cast trigger")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, theirs); !tapped {
		t.Error("an opponent's land must stay tapped")
	}
}

func TestEmrakulHasWardSacrificeThreePermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.RfcbPush(g, me.ID, "Emrakul, the Exigent Doom", "Legendary Creature — Eldrazi", effectstest.RfcbEmrakul, 12, 12)
	abilities := effectstest.EffectiveAbilities(t, g, id)
	if !effectstest.HasAbility(abilities, "flying") || !effectstest.HasAbility(abilities, "trample") {
		t.Errorf("abilities = %v, want flying and trample", abilities)
	}
	if len(game.TriggersForCard(game.Card{OracleID: effectstest.RfcbEmrakul})) < 2 {
		t.Error("expected the cast trigger and the ward trigger")
	}
}

// --- Grim Repriser ----------------------------------------------------------------------

func TestGrimRepriserReturnsWithFinalityOnlyAfterNoncombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
	id := effectstest.PushCatalogGraveyardCard(me, "Grim Repriser", "Creature — Zombie Bard", effectstest.RfcbGrimRepriser, 2, 2)
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
	effectstest.PassPriorityAroundTable(t, g)
	if !rfcbOnBattlefield(g, id) {
		t.Fatal("the Repriser did not return")
	}
	if got := effectstest.AllCountersOn(t, g, id)[effects.RfCreatureBFinalityCounter]; got != 1 {
		t.Fatalf("finality counters = %d, want 1", got)
	}
	effectstest.Destroy(t, g, id)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(id) {
		t.Error("a creature with a finality counter is exiled instead of dying")
	}
}

func TestGrimRepriserDoesNotCountCombatDamageOrDamageToYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
	id := effectstest.PushCatalogGraveyardCard(me, "Grim Repriser", "Creature — Zombie Bard", effectstest.RfcbGrimRepriser, 2, 2)
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
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastAndResolveCreature(t, g, "Grim Repriser", "Creature — Zombie Bard", effectstest.RfcbGrimRepriser)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AllCountersOn(t, g, id)[effects.RfCreatureBFinalityCounter]; got != 0 {
		t.Errorf("finality counters = %d, want 0 for a hard cast", got)
	}
}
