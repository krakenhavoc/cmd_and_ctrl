package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b107_cards_test.go — Core Set 2021, batch 107 (#5511), the six
// planeswalkers. Every test plays the real catalog entry through the engine.

const (
	m21B107Basri        = "1601e553-31c2-4e04-9f0c-bc17e49eeff7"
	m21B107Chandra      = "58ad45fc-b625-4484-8d49-ebc42db9040a"
	m21B107Teferi       = "7240037f-ddea-425e-9824-866e2bb0dc84"
	m21B107LilianaMage  = "d654dd73-763d-4dfb-a30f-a058fd7f27a6"
	m21B107LilianaWaker = "e0df8e2a-96d2-4e4d-b065-aa2e162abbfe"
	m21B107Garruk       = "fc762e40-84c1-4762-82f7-a0fc4f6fd522"

	m21B107Bear = "Creature — Bear"
)

// m21B107Table is a catalog game at the active seat's main phase with a
// planeswalker of theirs on the battlefield.
func m21B107Table(t *testing.T, name, oracle string, loyalty int) (g *game.Game, me, opp *game.Player, pw uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pw = effectstest.PushWalkerForTest(g, me.ID, name, oracle, loyalty)
	return
}

func m21B107Activate(t *testing.T, g *game.Game, me *game.Player, pw uuid.UUID, row int, targets ...game.TargetRef) {
	t.Helper()
	effectstest.ApaActivate(t, g, me, pw, row, game.ActivateAbilityParams{Targets: targets})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestM21B107EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		m21B107Basri:        "Basri, Devoted Paladin",
		m21B107Chandra:      "Chandra, Heart of Fire",
		m21B107Teferi:       "Teferi, Timeless Voyager",
		m21B107LilianaMage:  "Liliana, Death Mage",
		m21B107LilianaWaker: "Liliana, Waker of the Dead",
		m21B107Garruk:       "Garruk, Savage Herald",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Basri ----------------------------------------------------------------------

func TestM21B107BasriPlusOneCountersAndGivesVigilance(t *testing.T) {
	g, me, _, pw := m21B107Table(t, "Basri, Devoted Paladin", m21B107Basri, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", m21B107Bear, 2, 2)
	m21B107Activate(t, g, me, pw, 0, game.TargetRef{Kind: game.TargetCard, ID: bear})
	if got := effectstest.PlusOnes(g, bear); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("no vigilance")
	}
	if got := effectstest.LoyaltyOf(g, pw); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestM21B107BasriMinusOneCountersEveryAttackerThisTurn(t *testing.T) {
	g, me, opp, pw := m21B107Table(t, "Basri, Devoted Paladin", m21B107Basri, 4)
	a := effectstest.B12Creature(g, me.ID, "A", m21B107Bear, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", m21B107Bear, 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", m21B107Bear, 2, 2)
	m21B107Activate(t, g, me, pw, 1)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, a) != 1 || effectstest.PlusOnes(g, b) != 1 {
		t.Errorf("attackers have %d and %d counters, want 1 each", effectstest.PlusOnes(g, a), effectstest.PlusOnes(g, b))
	}
	if effectstest.PlusOnes(g, idle) != 0 {
		t.Error("a creature that didn't attack got a counter")
	}
}

func TestM21B107BasriMinusSixPumpsAndGivesFlying(t *testing.T) {
	g, me, opp, pw := m21B107Table(t, "Basri, Devoted Paladin", m21B107Basri, 6)
	mine := effectstest.B12Creature(g, me.ID, "Mine", m21B107Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", m21B107Bear, 2, 2)
	m21B107Activate(t, g, me, pw, 2)
	if p, tough := effectstest.PtOf(t, g, mine); p != 4 || tough != 4 || !effectstest.HasEffectiveKeyword(t, g, mine, "flying") {
		t.Errorf("mine is %d/%d", p, tough)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "flying") {
		t.Error("an opponent's creature was affected")
	}
}

// --- Chandra --------------------------------------------------------------------

func TestM21B107ChandraDiscardsTheHandThenExilesThreePlayable(t *testing.T) {
	g, me, _, pw := m21B107Table(t, "Chandra, Heart of Fire", m21B107Chandra, 5)
	for me.Hand.Size() > 0 {
		_, _ = me.Hand.PopTop()
	}
	old := effectstest.EoHand(me, "Old Card", "Instant")
	var top []uuid.UUID
	for _, n := range []string{"Top One", "Top Two", "Top Three", "Fourth"} {
		id := uuid.New()
		top = append([]uuid.UUID{id}, top...)
		me.Library.PushTop(game.Card{InstanceID: id, Name: n, TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	m21B107Activate(t, g, me, pw, 0)
	if me.Hand.Contains(old) || !me.Graveyard.Contains(old) {
		t.Error("the hand was not discarded")
	}
	exiled := 0
	for _, id := range top[:3] {
		if effectstest.ZoneOfCard(g, id) == game.ZoneExile {
			exiled++
		}
	}
	if exiled != 3 {
		t.Errorf("%d of the top three are exiled, want 3", exiled)
	}
	if effectstest.ZoneOfCard(g, top[3]) != game.ZoneLibrary {
		t.Error("a fourth card was exiled")
	}
	if perm := effectstest.ExiledPermission(g, top[0]); perm.Player != me.ID || perm.CastOnly {
		t.Errorf("permission %+v: want yours and not cast-only", perm)
	}
	if err := g.CastSpell(me.ID, top[0], game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Errorf("the exiled land can't be played: %v", err)
	}
}

func TestM21B107ChandraPlusOneDealsTwo(t *testing.T) {
	g, me, opp, pw := m21B107Table(t, "Chandra, Heart of Fire", m21B107Chandra, 5)
	life := opp.Life
	m21B107Activate(t, g, me, pw, 1, effectstest.PlayerTarget(opp.ID))
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
	if got := effectstest.LoyaltyOf(g, pw); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestM21B107ChandraMinusNineExilesRedSpellsAddsSixAndLetsYouCastThem(t *testing.T) {
	g, me, _, pw := m21B107Table(t, "Chandra, Heart of Fire", m21B107Chandra, 9)
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	redSorcery := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Red Sorcery", TypeLine: "Sorcery", Colors: []string{"R"}, ManaCost: "{1}{R}"})
	blueInstant := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Blue Instant", TypeLine: "Instant", Colors: []string{"U"}, ManaCost: "{U}"})
	redCreature := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Red Creature", TypeLine: m21B107Bear, Colors: []string{"R"}, ManaCost: "{R}"})
	redInstant := effectstest.PushGraveyardCardWithTypeLineID(me, "Red Instant", "Instant")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == redInstant {
				me.Graveyard.Cards[i].Colors = []string{"R"}
				me.Graveyard.Cards[i].ManaCost = "{R}"
			}
		}
	})
	effectstest.ApaActivate(t, g, me, pw, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	offered := map[uuid.UUID]bool{}
	for _, id := range effectstest.SearchChoiceFor(g, me.ID).SearchCards {
		offered[id] = true
	}
	if !offered[redSorcery] || !offered[redInstant] || offered[blueInstant] || offered[redCreature] {
		t.Fatalf("offered %v: want only the red sorcery and the red instant", offered)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, redSorcery, redInstant)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, redSorcery) != game.ZoneExile || effectstest.ZoneOfCard(g, redInstant) != game.ZoneExile {
		t.Error("the chosen cards were not exiled")
	}
	if got := me.ManaPool.Total(); got != 6 {
		t.Errorf("pool holds %d, want 6", got)
	}
	if err := g.CastSpell(me.ID, redSorcery, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Errorf("can't cast the exiled sorcery with the six {R}: %v", err)
	}
}

// --- Teferi ---------------------------------------------------------------------

func TestM21B107TeferiDrawsAndTucks(t *testing.T) {
	g, me, opp, pw := m21B107Table(t, "Teferi, Timeless Voyager", m21B107Teferi, 4)
	effectstest.SeedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	m21B107Activate(t, g, me, pw, 0)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	victim := effectstest.B12Creature(g, opp.ID, "Victim", m21B107Bear, 2, 2)
	pw2 := effectstest.PushWalkerForTest(g, me.ID, "Teferi Second Copy", m21B107Teferi, 4)
	m21B107Activate(t, g, me, pw2, 1, game.TargetRef{Kind: game.TargetCard, ID: victim})
	if effectstest.ZoneOfCard(g, victim) != game.ZoneLibrary {
		t.Errorf("the creature is in %v, want the library", effectstest.ZoneOfCard(g, victim))
	}
	top, err := opp.Library.Top()
	if err != nil || top.InstanceID != victim {
		t.Error("it is not on top of its owner's library")
	}
}

func TestM21B107TeferiMinusEightPhasesOutUntilTheEndOfYourNextTurn(t *testing.T) {
	g, me, opp, pw := m21B107Table(t, "Teferi, Timeless Voyager", m21B107Teferi, 8)
	oppSeat, mySeat := (g.Turn.ActiveSeat+1)%len(g.Seats), g.Turn.ActiveSeat
	a := effectstest.B12Creature(g, opp.ID, "A", m21B107Bear, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", m21B107Bear, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", m21B107Bear, 2, 2)
	later := effectstest.B12Creature(g, opp.ID, "Later", m21B107Bear, 2, 2)
	_ = later
	m21B107Activate(t, g, me, pw, 2, effectstest.PlayerTarget(opp.ID))
	out := func(id uuid.UUID) bool {
		v := false
		g.ReadSnapshot(func() { v = g.IsPhasedOutForEffect(id) })
		return v
	}
	if !out(a) || !out(b) || !out(later) || out(mine) {
		t.Fatalf("phased out: a=%v b=%v later=%v mine=%v", out(a), out(b), out(later), out(mine))
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, oppSeat)
	if !out(a) || !out(b) {
		t.Error("they phased in during the opponent's next untap step")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, mySeat)
	if !out(a) || !out(b) {
		t.Error("they phased in during your next turn")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, oppSeat)
	if out(a) || out(b) {
		t.Error("they stayed out after the end of your next turn")
	}
}

// --- Liliana, Death Mage --------------------------------------------------------

func TestM21B107LilianaDeathMagePlusOneReturnsACreatureCard(t *testing.T) {
	g, me, _, pw := m21B107Table(t, "Liliana, Death Mage", m21B107LilianaMage, 4)
	dead := effectstest.PushGraveyardCardWithTypeLineID(me, "Dead Bear", m21B107Bear)
	m21B107Activate(t, g, me, pw, 0, game.TargetRef{Kind: game.TargetCard, ID: dead})
	if !me.Hand.Contains(dead) {
		t.Error("the creature card did not return to hand")
	}
}

func TestM21B107LilianaDeathMageMinusThreeDestroysAndDrains(t *testing.T) {
	for _, indestructible := range []bool{false, true} {
		g, me, opp, pw := m21B107Table(t, "Liliana, Death Mage", m21B107LilianaMage, 4)
		victim := effectstest.B12Creature(g, opp.ID, "Victim", m21B107Bear, 2, 2)
		if indestructible {
			g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, victim).Keywords = []string{"indestructible"} })
		}
		life := opp.Life
		m21B107Activate(t, g, me, pw, 1, game.TargetRef{Kind: game.TargetCard, ID: victim})
		if g.Battlefield.Contains(victim) == !indestructible {
			t.Errorf("indestructible %v: creature on the battlefield %v", indestructible, g.Battlefield.Contains(victim))
		}
		if opp.Life != life-2 {
			t.Errorf("indestructible %v: life %d, want %d", indestructible, opp.Life, life-2)
		}
	}
}

func TestM21B107LilianaDeathMageMinusSevenCountsCreatureCards(t *testing.T) {
	g, me, opp, pw := m21B107Table(t, "Liliana, Death Mage", m21B107LilianaMage, 7)
	for i := 0; i < 3; i++ {
		effectstest.PushGraveyardCardWithTypeLineID(opp, "Dead Bear", m21B107Bear)
	}
	effectstest.PushGraveyardCardWithTypeLineID(opp, "Dead Spell", "Instant")
	life := opp.Life
	m21B107Activate(t, g, me, pw, 2, effectstest.PlayerTarget(opp.ID))
	if opp.Life != life-6 {
		t.Errorf("life %d, want %d", opp.Life, life-6)
	}
}

// --- Liliana, Waker of the Dead -------------------------------------------------

func TestM21B107LilianaWakerEachPlayerDiscardsAndEmptyHandsLoseThree(t *testing.T) {
	g, me, opp, pw := m21B107Table(t, "Liliana, Waker of the Dead", m21B107LilianaWaker, 4)
	third, fourth := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)], g.Seats[(g.Turn.ActiveSeat+3)%len(g.Seats)]
	for _, p := range g.Seats {
		for p.Hand.Size() > 0 {
			_, _ = p.Hand.PopTop()
		}
	}
	mine := effectstest.EoHand(me, "Mine", "Instant")
	theirs := effectstest.EoHand(opp, "Theirs", "Instant")
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.ApaActivate(t, g, me, pw, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, mine)
	effectstest.AnswerDiscard(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(mine) || opp.Hand.Contains(theirs) {
		t.Error("a player kept their card")
	}
	for _, p := range []*game.Player{third, fourth} {
		if p.Life != lives[p.ID]-3 {
			t.Errorf("%s (empty hand) is at %d, want %d", p.Name, p.Life, lives[p.ID]-3)
		}
	}
	if opp.Life != lives[opp.ID] || me.Life != lives[me.ID] {
		t.Error("a player who could discard lost life")
	}
}

func TestM21B107LilianaWakerMinusThreeShrinksByYourGraveyard(t *testing.T) {
	g, me, opp, pw := m21B107Table(t, "Liliana, Waker of the Dead", m21B107LilianaWaker, 4)
	for i := 0; i < 3; i++ {
		effectstest.PushGraveyardCardWithTypeLineID(me, "Filler", "Instant")
	}
	victim := effectstest.B12Creature(g, opp.ID, "Victim", m21B107Bear, 4, 4)
	m21B107Activate(t, g, me, pw, 1, game.TargetRef{Kind: game.TargetCard, ID: victim})
	if p, tough := effectstest.PtOf(t, g, victim); p != 1 || tough != 1 {
		t.Errorf("victim is %d/%d, want 1/1", p, tough)
	}
}

func TestM21B107LilianaWakerEmblemReanimatesWithHaste(t *testing.T) {
	g, me, opp, pw := m21B107Table(t, "Liliana, Waker of the Dead", m21B107LilianaWaker, 7)
	theirs := effectstest.PushGraveyardCardWithTypeLineID(opp, "Their Dead Bear", m21B107Bear)
	m21B107Activate(t, g, me, pw, 2)
	if len(effectstest.EmblemsOf(g, me.ID)) != 1 {
		t.Fatalf("emblems: %v", effectstest.EmblemsOf(g, me.ID))
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.HobAnswerTargets(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Fatal("the creature card was not put onto the battlefield")
	}
	if c := effectstest.B12Card(t, g, theirs); c.Controller != me.ID {
		t.Error("it is not under your control")
	}
	if !effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("no haste")
	}
}

// --- Garruk ---------------------------------------------------------------------

func TestM21B107GarrukPlusOneKeepsACreatureAndBottomsTheRest(t *testing.T) {
	for _, creature := range []bool{true, false} {
		g, me, _, pw := m21B107Table(t, "Garruk, Savage Herald", m21B107Garruk, 5)
		typeLine := "Sorcery"
		if creature {
			typeLine = m21B107Bear
		}
		top := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: top, Name: "Top", TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		m21B107Activate(t, g, me, pw, 0)
		if creature != me.Hand.Contains(top) {
			t.Errorf("creature %v: in hand %v", creature, me.Hand.Contains(top))
		}
		if !creature {
			if effectstest.ZoneOfCard(g, top) != game.ZoneLibrary {
				t.Fatal("the card left the library")
			}
			if next, err := me.Library.Top(); err == nil && next.InstanceID == top {
				t.Error("it stayed on top instead of going to the bottom")
			}
		}
	}
}

func TestM21B107GarrukMinusTwoBitesAnotherCreature(t *testing.T) {
	g, me, opp, pw := m21B107Table(t, "Garruk, Savage Herald", m21B107Garruk, 5)
	biter := effectstest.B12Creature(g, me.ID, "Biter", m21B107Bear, 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", m21B107Bear, 2, 3)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: biter}, {Kind: game.TargetCard, ID: biter}}}); err == nil {
		t.Fatal("the same creature was both targets")
	}
	m21B107Activate(t, g, me, pw, 1, game.TargetRef{Kind: game.TargetCard, ID: biter}, game.TargetRef{Kind: game.TargetCard, ID: victim})
	if g.Battlefield.Contains(victim) {
		t.Error("the creature took no damage")
	}
	if !g.Battlefield.Contains(biter) {
		t.Error("the biter was hurt (it is a one-sided bite)")
	}
}

func TestM21B107GarrukMinusSevenLetsCreaturesAssignAsUnblocked(t *testing.T) {
	g, me, opp, pw := m21B107Table(t, "Garruk, Savage Herald", m21B107Garruk, 7)
	mine := effectstest.B12Creature(g, me.ID, "Mine", m21B107Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", m21B107Bear, 2, 2)
	m21B107Activate(t, g, me, pw, 2)
	if !effectstest.LayeredCard(t, g, mine).Effective().AssignsAsThoughUnblocked {
		t.Error("my creature has no option")
	}
	if effectstest.LayeredCard(t, g, theirs).Effective().AssignsAsThoughUnblocked {
		t.Error("an opponent's creature has the option")
	}
}
