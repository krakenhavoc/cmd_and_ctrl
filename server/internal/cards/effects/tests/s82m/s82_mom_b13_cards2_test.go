package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b13_cards2_test.go — the second half of the March of the Machine
// batch 13 tests (#5477): the attack, dig and modal creatures.

// --- Scrappy Bruiser -----------------------------------------------------------

func TestS82B13ScrappyBruiserPumpsAnAttackerAndReturnsItAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bruiser := effectstest.B12Push(g, me.ID, "Scrappy Bruiser", "Creature — Raccoon Warrior", s82b13BruiserOracle, 3, 4)
	ally := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range []uuid.UUID{bruiser, ally} {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the attack trigger")
	}
	if effectstest.HasID(p.PickTargetCards, idle) {
		t.Error("a non-attacking creature is offered as a target")
	}
	effectstest.PickCard(t, g, me.ID, ally)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, ally); got != 4 {
		t.Errorf("the target's power is %d, want 4 (2 + 2)", got)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, ally, "trample") {
		t.Error("the target did not gain trample")
	}
	if got := effectstest.EffectivePower(t, g, bruiser); got != 3 {
		t.Errorf("the Bruiser itself is %d power, want 3", got)
	}
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if g.Battlefield.Contains(ally) {
		t.Error("the target is still on the battlefield after combat")
	}
	if !me.Hand.Contains(ally) || me.Hand.Size() != hand+1 {
		t.Error("the target did not return to its owner's hand at end of combat")
	}
	if !g.Battlefield.Contains(bruiser) {
		t.Error("the Bruiser was returned")
	}
}

func TestS82B13ScrappyBruiserMayTargetItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bruiser := effectstest.B12Push(g, me.ID, "Scrappy Bruiser", "Creature — Raccoon Warrior", s82b13BruiserOracle, 3, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bruiser, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PickCard(t, g, me.ID, bruiser)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bruiser); got != 5 {
		t.Errorf("power %d, want 5", got)
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if !me.Hand.Contains(bruiser) {
		t.Error("the Bruiser did not return to hand after targeting itself")
	}
}

// --- Oracle of Tragedy -----------------------------------------------------------

func TestS82B13OracleOfTragedyLootsWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreatureNoResolve(t, g, "Oracle of Tragedy", "Creature — Human Wizard", s82b13OracleOfTragedy, 1, 3)
	libraryBefore := me.Library.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1}, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := libraryBefore - me.Library.Size(); got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if got := me.Graveyard.Size(); got != 1 {
		t.Errorf("%d cards in the graveyard after the discard, want 1", got)
	}
}

func TestS82B13OracleOfTragedyShufflesBigCardsBackWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oracle := effectstest.B12Push(g, me.ID, "Oracle of Tragedy", "Creature — Human Wizard", s82b13OracleOfTragedy, 1, 3)
	big := uuid.New()
	small := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Big Wurm", TypeLine: "Creature — Wurm", ManaCost: "{4}{G}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: small, Name: "Small Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(oracle) })
	g.RunStateChecksForTest()
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1}, 1)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p != nil {
		if effectstest.HasID(p.PickTargetCards, small) {
			t.Error("a mana value 2 card is offered")
		}
		if !effectstest.HasID(p.PickTargetCards, big) {
			t.Error("a mana value 5 card is not offered")
		}
		effectstest.PickCard(t, g, me.ID, big)
	}
	libraryBefore := me.Library.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(big) {
		t.Error("the big card is still in the graveyard")
	}
	if !me.Graveyard.Contains(small) {
		t.Error("the small card left the graveyard")
	}
	if got := me.Library.Size() - libraryBefore; got != 1 {
		t.Errorf("the library grew by %d, want 1", got)
	}
}

// --- Sandstalker Moloch ------------------------------------------------------------

func s82b13LibraryTop(me *game.Player, cards ...game.Card) []uuid.UUID {
	var ids []uuid.UUID
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = me.ID, me.ID
		me.Library.PushTop(c)
		ids = append([]uuid.UUID{c.InstanceID}, ids...)
	}
	return ids
}

func TestS82B13SandstalkerMolochDigsOnlyAfterAnOpponentCastBlueOrBlack(t *testing.T) {
	for _, tc := range []struct {
		name   string
		colors string
		by     int
		want   bool
	}{
		{"blue spell by an opponent", "U", 1, true},
		{"black spell by another opponent", "B", 2, true},
		{"red spell by an opponent", "R", 1, false},
		{"blue spell by me", "U", 0, false},
		{"no spell", "", 1, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		if tc.colors != "" {
			if g.SpellsCastThisTurn == nil {
				g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{}
			}
			g.SpellsCastThisTurn[g.Seats[tc.by].ID] = game.CastTally{Total: 1, Colors: tc.colors}
		}
		top := s82b13LibraryTop(me,
			game.Card{Name: "Top Instant", TypeLine: "Instant"},
			game.Card{Name: "Top Wurm", TypeLine: "Creature — Wurm", Power: 3, Toughness: 3},
			game.Card{Name: "Top Land", TypeLine: "Basic Land — Forest"},
			game.Card{Name: "Top Sorcery", TypeLine: "Sorcery"},
		)
		effectstest.CastHoldCreatureNoResolve(t, g, "Sandstalker Moloch", "Creature — Lizard", s82b13MolochOracle, 4, 2)
		effectstest.PassPriorityAroundTable(t, g)
		prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
		if tc.want != (prompt != nil) {
			t.Errorf("%s: dig prompt open = %v, want %v", tc.name, prompt != nil, tc.want)
			continue
		}
		if !tc.want {
			if got := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; got != top[0] {
				t.Errorf("%s: the library was touched", tc.name)
			}
			continue
		}
		for _, id := range prompt.ChooseCards {
			if id == top[0] || id == top[3] {
				t.Errorf("%s: an instant or sorcery is offered", tc.name)
			}
		}
		hand := me.Hand.Size()
		effectstest.AnswerChooseCards(t, g, me.ID, top[1])
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(top[1]) || me.Hand.Size() != hand+1 {
			t.Errorf("%s: the picked permanent did not reach the hand", tc.name)
		}
		for _, id := range []uuid.UUID{top[0], top[2], top[3]} {
			if me.Hand.Contains(id) {
				t.Errorf("%s: an unpicked card reached the hand", tc.name)
			}
		}
		lib := me.Library.Cards
		bottom := map[uuid.UUID]bool{}
		for _, c := range lib[:3] {
			bottom[c.InstanceID] = true
		}
		for _, id := range []uuid.UUID{top[0], top[2], top[3]} {
			if !bottom[id] {
				t.Errorf("%s: an unpicked card is not on the bottom of the library", tc.name)
			}
		}
	}
}

// --- Deeproot Wayfinder -------------------------------------------------------------

func TestS82B13DeeprootWayfinderSurveilsThenReturnsALandTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wayfinder := effectstest.B12Push(g, me.ID, "Deeproot Wayfinder", "Creature — Merfolk Scout", s82b13WayfinderOracle, 2, 3)
	land := effectstest.PushGraveyardCardTyped(me, "Forest", "Basic Land — Forest")
	bear := effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	effectstest.AttackWith(t, g, opp.ID, wayfinder)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	// The surveil comes first: leave the top card where it is.
	s82b13KeepSurveil(t, g, me)
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no land prompt after the surveil")
	}
	if !effectstest.HasID(prompt.ChooseCards, land) || effectstest.HasID(prompt.ChooseCards, bear) {
		t.Errorf("land prompt offers %v, want only the land card", prompt.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardForTest(g, land)
	if c == nil || !c.Tapped {
		t.Errorf("the land is on the battlefield tapped? %v", c)
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the creature card left the graveyard")
	}
}

func TestS82B13DeeprootWayfinderDoesNothingWhenBlockedAndTheLandIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wayfinder := effectstest.B12Push(g, me.ID, "Deeproot Wayfinder", "Creature — Merfolk Scout", s82b13WayfinderOracle, 2, 3)
	land := effectstest.PushGraveyardCardTyped(me, "Forest", "Basic Land — Forest")
	effectstest.AttackWith(t, g, opp.ID, wayfinder)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	s82b13KeepSurveil(t, g, me)
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no land prompt")
	}
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !me.Graveyard.Contains(land) {
		t.Error("declining the land still moved it")
	}
}

// s82b13KeepSurveil answers the open surveil prompt by keeping every card on
// top, then lets the rest of the trigger run.
func s82b13KeepSurveil(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSurveil && c.Chooser == me.ID {
			if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
				t.Fatalf("ResolveSurveil: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			return
		}
	}
	t.Fatal("no surveil prompt is open")
}
