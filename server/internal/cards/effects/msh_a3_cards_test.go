package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_a3_cards_test.go — Marvel Super Heroes, half A, third batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshA3GatheringPlaceOracle = "36b58705-c5a5-4547-8d8b-a7c35e1f69ae"
	mshA3FalconOracle         = "92e222c2-28e6-4d43-8b48-480ebcd88495"
	mshA3HydraInfilOracle     = "27692079-a517-4d38-b452-7409980b5912"
	mshA3HawkeyesBowOracle    = "83798eb8-b14f-4620-bce2-1a6804a24dc7"
	mshA3HawkeyeYAOracle      = "c8f8b4ed-8455-45cd-a24c-e2f42cd5e5b1"
	mshA3HerbieOracle         = "393ff8c5-4da0-4290-b498-222cc0244026"
	mshA3EvilsThrallOracle    = "d7562f89-4808-44ab-8dc9-fc4078e8a47b"
	mshA3BlackWidowOracle     = "579f9a86-a1af-4c2a-a7e5-8fea0bd0eb61"
	mshA3DeathEnemiesOracle   = "c2b2b743-dfc7-439e-8563-1ccce54422bd"
	mshA3DoomReignsOracle     = "f398a742-22c1-4da9-9d49-fdce2eb93d01"
)

func TestMshA3EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mshA3GatheringPlaceOracle: "Gathering Place",
		mshA3FalconOracle:         "Falcon, Winged Wonder",
		mshA3HydraInfilOracle:     "HYDRA Infiltration",
		mshA3HawkeyesBowOracle:    "Hawkeye's Bow",
		mshA3HawkeyeYAOracle:      "Hawkeye, Young Avenger",
		mshA3HerbieOracle:         "H.E.R.B.I.E. Scout Unit",
		mshA3EvilsThrallOracle:    "Evil's Thrall",
		mshA3BlackWidowOracle:     "Black Widow, Super Spy",
		mshA3DeathEnemiesOracle:   "Death to Our Enemies",
		mshA3DoomReignsOracle:     "Doom Reigns Supreme",
	}
	for oracle, name := range want {
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

func TestMshA3CombatKeywordsAreDeclared(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct{ name, typeLine, oracle, keyword string }{
		{"Falcon, Winged Wonder", "Legendary Creature — Human Hero", mshA3FalconOracle, "flying"},
		{"H.E.R.B.I.E. Scout Unit", "Artifact Creature — Robot Scout", mshA3HerbieOracle, "flying"},
		{"Hawkeye, Young Avenger", "Legendary Creature — Human Archer Hero", mshA3HawkeyeYAOracle, "reach"},
		{"Black Widow, Super Spy", "Legendary Creature — Human Spy Hero", mshA3BlackWidowOracle, "menace"},
	} {
		id := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
		if !mshA3HasKeyword(g, id, tc.keyword) {
			t.Errorf("%s has no %s", tc.name, tc.keyword)
		}
	}
}

// --- Gathering Place -------------------------------------------------------------------

// mshA3TapSecondRow taps Gathering Place's gated {G}/{W} row and picks `color`.
func mshA3TapSecondRow(t *testing.T, g *game.Game, who, land uuid.UUID, color string) {
	t.Helper()
	if err := g.ActivateManaAbility(who, land, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	b10ResolveAllManaPicks(t, g, who, color)
}

func TestMshA3GatheringPlaceTapsForColourOnTheTurnItEntersWithNoBasic(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := playLandFromHand(t, g, "Gathering Place", mshA3GatheringPlaceOracle)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	mshA3TapSecondRow(t, g, me.ID, land, "W")
	if got := poolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("tapped for {W} the turn it entered: pool %v", got)
	}
}

func TestMshA3GatheringPlaceNeedsABasicLandAfterTheTurnItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushCatalogPermanent(g, me.ID, "Gathering Place", "Land", mshA3GatheringPlaceOracle, false)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("the coloured ability worked with no basic land and without having entered this turn")
	}
	if b16Tapped(t, g, land) {
		t.Error("a refused activation tapped the land")
	}
	pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	mshA3TapSecondRow(t, g, me.ID, land, "G")
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("tapped for {G} with a basic land: pool %v", got)
	}
}

func TestMshA3GatheringPlaceAlwaysTapsForColourless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushCatalogPermanent(g, me.ID, "Gathering Place", "Land", mshA3GatheringPlaceOracle, false)
	advanceToMain(t, g)
	b28TapForMana(t, g, me.ID, land, "")
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("colourless ability: pool %v", got)
	}
}

func TestMshA3GatheringPlaceANonbasicLandDoesNotCount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushCatalogPermanent(g, me.ID, "Gathering Place", "Land", mshA3GatheringPlaceOracle, false)
	pushCatalogPermanent(g, me.ID, "Other Place", "Land", "", false)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("a nonbasic land satisfied the gate")
	}
}

// --- Falcon, Winged Wonder -----------------------------------------------------------------

func TestMshA3FalconMakesRedwingWhoSurveilsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12PlayFromHand(t, g, "Falcon, Winged Wonder", "Legendary Creature — Human Hero", mshA3FalconOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	var wing *game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == me.ID && c.Name == "Redwing" {
				cc := c
				wing = &cc
			}
		}
	})
	if wing == nil {
		t.Fatal("Falcon entering made no Redwing")
	}
	if !wing.IsToken() || !wing.IsLegendary() || wing.Power != 1 || wing.Toughness != 1 || !wing.HasSubtype("Bird") || !wing.HasSubtype("Scout") {
		t.Errorf("Redwing is %+v, want a legendary 1/1 Bird Scout token", *wing)
	}
	if !mshA3HasKeyword(g, wing.InstanceID, "flying") {
		t.Error("Redwing has no flying")
	}
	// The attack declaration itself: a token that just entered is
	// summoning-sick, and the declaration is combat's, not Redwing's.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: wing.InstanceID, Actor: me.ID, Target: opp.ID})
	})
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID) == nil {
		t.Fatalf("Redwing attacking asked for no surveil; pending %+v", g.PendingChoices)
	}
}

// --- HYDRA Infiltration ----------------------------------------------------------------------

func TestMshA3HydraInfiltrationMakesAnOpponentDiscardTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Hand.Size()
	if before < 2 {
		t.Fatalf("test setup: opponent hand %d", before)
	}
	castCatalogSpell(t, g, "HYDRA Infiltration", "Enchantment", mshA3HydraInfilOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	hand := opp.Hand.Cards
	answerDiscard(t, g, opp.ID, hand[0].InstanceID, hand[1].InstanceID)
	if got := before - opp.Hand.Size(); got != 2 {
		t.Errorf("opponent discarded %d, want 2", got)
	}
}

func TestMshA3HydraInfiltrationDrainsWhenACreatureAttacksAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "HYDRA Infiltration", "Enchantment", mshA3HydraInfilOracle, false)
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	myLife, theirLife := me.Life, opp.Life
	declareAttack(t, g, opp.ID, bear)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if me.Life != myLife+1 || opp.Life != theirLife-1 {
		t.Errorf("life %d/%d -> %d/%d, want +1 for me and -1 for them", myLife, theirLife, me.Life, opp.Life)
	}
}

func TestMshA3HydraInfiltrationDoesNothingWhenTwoCreaturesAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "HYDRA Infiltration", "Enchantment", mshA3HydraInfilOracle, false)
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wolf := b12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	myLife, theirLife := me.Life, opp.Life
	declareAttack(t, g, opp.ID, bear, wolf)
	passPriorityAroundTable(t, g)
	if me.Life != myLife || opp.Life != theirLife {
		t.Errorf("life changed to %d/%d with two attackers", me.Life, opp.Life)
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a trigger asked for a target with two attackers")
	}
}

// --- Hawkeye's Bow ------------------------------------------------------------------------------

// mshA3HasKeyword reads a permanent's effective keyword.
func mshA3HasKeyword(g *game.Game, id uuid.UUID, kw string) bool {
	var has bool
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			has = game.HasKeyword(&c, kw)
		}
	})
	return has
}

func mshA3EquipBow(t *testing.T, g *game.Game, owner uuid.UUID) (bow, bear uuid.UUID) {
	t.Helper()
	bow = pushCatalogPermanent(g, owner, "Hawkeye's Bow", "Artifact — Equipment", mshA3HawkeyesBowOracle, false)
	bear = b12Creature(g, owner, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AttachForEffect(bow, game.TargetRef{Kind: game.TargetCard, ID: bear}) })
	return bow, bear
}

func TestMshA3HawkeyesBowGivesPlusOneReachAndPingsWhenTheCreatureTaps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	_, bear := mshA3EquipBow(t, g, me.ID)
	if got := hobA4Power(t, g, bear); got != 3 {
		t.Errorf("equipped power %d, want 3", got)
	}
	if !mshA3HasKeyword(g, bear, "reach") {
		t.Error("the equipped creature has no reach")
	}
	before := []int{opp.Life, g.Seats[2].Life, g.Seats[3].Life}
	mine := me.Life
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	passPriorityAroundTable(t, g)
	for i, s := range g.Seats[1:] {
		if s.Life != before[i]-1 {
			t.Errorf("opponent %d life %d -> %d, want -1", i+1, before[i], s.Life)
		}
	}
	if me.Life != mine {
		t.Errorf("my life changed %d -> %d", mine, me.Life)
	}
}

func TestMshA3HawkeyesBowPingsWhenTheCreatureAttacksButNotWithVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	_, bear := mshA3EquipBow(t, g, me.ID)
	before := opp.Life
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("attacker tapped: opponent life %d -> %d, want -1", before, opp.Life)
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	advanceToMain(t, g2)
	bow2 := pushCatalogPermanent(g2, me2.ID, "Hawkeye's Bow", "Artifact — Equipment", mshA3HawkeyesBowOracle, false)
	bear2 := pushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Vigil Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Keywords: []string{"vigilance"}, Owner: me2.ID, Controller: me2.ID,
	})
	g2.WithWriteLock(func() { _ = g2.AttachForEffect(bow2, game.TargetRef{Kind: game.TargetCard, ID: bear2}) })
	before2 := opp2.Life
	declareAttack(t, g2, opp2.ID, bear2)
	passPriorityAroundTable(t, g2)
	if opp2.Life != before2 {
		t.Errorf("a vigilance attacker did not become tapped, but opponent life %d -> %d", before2, opp2.Life)
	}
}

func TestMshA3HawkeyesBowIgnoresAnUnequippedCreatureTapping(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mshA3EquipBow(t, g, me.ID)
	other := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	before := opp.Life
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(other) })
	passPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("another creature tapped but opponent life %d -> %d", before, opp.Life)
	}
}

// --- Hawkeye, Young Avenger ---------------------------------------------------------------------

func TestMshA3HawkeyeAddsHisPowerToNoncombatDamageToAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	hawkeye := pushCatalogPermanent(g, me.ID, "Hawkeye, Young Avenger", "Legendary Creature — Human Archer Hero", mshA3HawkeyeYAOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(hawkeye, game.CounterPlusOne, 2) })
	source := b12Creature(g, me.ID, "Pinger", "Creature — Bear", 2, 2)
	before := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(source, opp.ID, 2) })
	if got := before - opp.Life; got != 5 {
		t.Errorf("2 noncombat damage became %d, want 5 (2 plus Hawkeye's power 3)", got)
	}
}

func TestMshA3HawkeyeAddsToDamageToAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Hawkeye, Young Avenger", "Legendary Creature — Human Archer Hero", mshA3HawkeyeYAOracle, false)
	source := b12Creature(g, me.ID, "Pinger", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Big Wall", "Creature — Wall", 0, 9)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(source, victim, 2) })
	if c, ok := battlefieldCard(g, victim); !ok || c.DamageMarked != 3 {
		t.Errorf("victim = %+v, want 3 damage marked (2 plus Hawkeye's power 1)", c)
	}
}

func TestMshA3HawkeyeLeavesDamageToYouAndYourCreaturesAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Hawkeye, Young Avenger", "Legendary Creature — Human Archer Hero", mshA3HawkeyeYAOracle, false)
	theirs := b12Creature(g, opp.ID, "Their Pinger", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 9)
	life := me.Life
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(theirs, me.ID, 2)
		_ = g.DealDamageToCreatureForEffect(theirs, mine, 2)
	})
	if me.Life != life-2 {
		t.Errorf("damage to me: life %d -> %d, want 2", life, me.Life)
	}
	if c, ok := battlefieldCard(g, mine); !ok || c.DamageMarked != 2 {
		t.Errorf("my creature = %+v, want 2 damage marked", c)
	}
}

func TestMshA3HawkeyeIgnoresAnOpponentsSourceAndCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Hawkeye, Young Avenger", "Legendary Creature — Human Archer Hero", mshA3HawkeyeYAOracle, false)
	// A source another player controls hitting a third player is not "a
	// source you control".
	theirs := b12Creature(g, opp.ID, "Their Pinger", "Creature — Bear", 2, 2)
	third := g.Seats[2]
	life := third.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirs, third.ID, 2) })
	if got := life - third.Life; got != 2 {
		t.Errorf("an opponent's damage to another opponent was %d, want 2", got)
	}
	// Combat damage from my creature is untouched.
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	before := opp.Life
	attackWith(t, g, opp.ID, bear)
	if got := before - opp.Life; got != 2 {
		t.Errorf("combat damage was %d, want exactly 2: only noncombat damage grows", got)
	}
}

// --- H.E.R.B.I.E. Scout Unit --------------------------------------------------------------------------

func TestMshA3HerbieDrawsThenPutsALandFromHandTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()
	b12PlayFromHand(t, g, "H.E.R.B.I.E. Scout Unit", "Artifact Creature — Robot Scout", mshA3HerbieOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	// +1 for the draw; the played card itself left the hand again.
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand grew by %d, want 1 (the draw, before the pick)", got)
	}
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no land pick after the draw; pending %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, land)
	if !ok || !card.Tapped {
		t.Errorf("the land is %+v, want on the battlefield tapped", card)
	}
}

func TestMshA3HerbieMayDeclineTheLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	b12PlayFromHand(t, g, "H.E.R.B.I.E. Scout Unit", "Artifact Creature — Robot Scout", mshA3HerbieOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no land pick")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatalf("declining the land: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, land); ok {
		t.Error("the land entered although the pick was declined")
	}
}

// --- Evil's Thrall ------------------------------------------------------------------------------------------

// mshA3ControllerOf reads who controls a permanent now.
func mshA3ControllerOf(g *game.Game, id uuid.UUID) uuid.UUID {
	var who uuid.UUID
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			who = c.Controller
		}
	})
	return who
}

// mshA3PassThroughTheTurnEnd advances until the current turn has ended.
func mshA3PassThroughTheTurnEnd(t *testing.T, g *game.Game) {
	t.Helper()
	seq := g.Turn.Seq
	for i := 0; i < 80 && g.Turn.Seq == seq; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.Seq == seq {
		t.Fatal("the turn never ended")
	}
}

func TestMshA3EvilsThrallStealsUntilEndOfTurnUntapsAndGivesHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b27Push(g, opp.ID, "Their Bear", "Creature — Bear", "", "{2}", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	castCatalogSpell(t, g, "Evil's Thrall", "Sorcery", mshA3EvilsThrallOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if mshA3ControllerOf(g, bear) != me.ID {
		t.Fatal("the creature was not stolen")
	}
	if c, _ := battlefieldCard(g, bear); c.Tapped {
		t.Error("the stolen creature is still tapped")
	}
	if !mshA3HasKeyword(g, bear, "haste") {
		t.Error("the stolen creature has no haste")
	}
	mshA3PassThroughTheTurnEnd(t, g)
	if mshA3ControllerOf(g, bear) != opp.ID {
		t.Error("with no Villain the creature did not go home at end of turn")
	}
}

func TestMshA3EvilsThrallWithAGreaterVillainKeepsItUntilTheEndOfYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b27Push(g, me.ID, "Big Villain", "Creature — Human Villain", "", "{4}", 4, 4)
	bear := b27Push(g, opp.ID, "Their Bear", "Creature — Bear", "", "{2}", 2, 2)
	castCatalogSpell(t, g, "Evil's Thrall", "Sorcery", mshA3EvilsThrallOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	mshA3PassThroughTheTurnEnd(t, g)
	if mshA3ControllerOf(g, bear) != me.ID {
		t.Fatal("the creature went home at the end of this turn despite the greater Villain")
	}
	// Through the three other turns to my next turn, and past its end.
	for g.Turn.ActiveSeat != me.Seat {
		mshA3PassThroughTheTurnEnd(t, g)
	}
	if mshA3ControllerOf(g, bear) != me.ID {
		t.Fatal("control lapsed before my next turn ended")
	}
	mshA3PassThroughTheTurnEnd(t, g)
	if mshA3ControllerOf(g, bear) != opp.ID {
		t.Error("control did not end with my next turn")
	}
}

func TestMshA3EvilsThrallNeedsTheVillainsManaValueToBeGreater(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b27Push(g, me.ID, "Equal Villain", "Creature — Human Villain", "", "{2}", 2, 2)
	b27Push(g, me.ID, "Big Hero", "Creature — Human Hero", "", "{6}", 6, 6)
	bear := b27Push(g, opp.ID, "Their Bear", "Creature — Bear", "", "{2}", 2, 2)
	castCatalogSpell(t, g, "Evil's Thrall", "Sorcery", mshA3EvilsThrallOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	mshA3PassThroughTheTurnEnd(t, g)
	if mshA3ControllerOf(g, bear) != opp.ID {
		t.Error("an equal-value Villain and a greater Hero extended the theft")
	}
}

// --- Black Widow, Super Spy ---------------------------------------------------------------------------------

// mshA3WidowConnects puts Black Widow on the table, stacks the first
// opponent's library with two lands over a costly nonland card, and lets her
// connect. It stops at the counter-or-cast question.
func mshA3WidowConnects(t *testing.T, g *game.Game) (widow, spell uuid.UUID, lands []uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	widow = pushCatalogPermanent(g, me.ID, "Black Widow, Super Spy", "Legendary Creature — Human Spy Hero", mshA3BlackWidowOracle, false)
	opp.Library.Cards = nil
	ids := sunbirdLibrary(opp,
		game.Card{Name: "First Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Second Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Costly Wurm", TypeLine: "Creature — Wurm", ManaCost: "{3}{U}", Power: 4, Toughness: 4},
		game.Card{Name: "Below", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
	)
	attackWith(t, g, opp.ID, widow)
	passPriorityAroundTable(t, g)
	return widow, ids[2], ids[:2]
}

func TestMshA3BlackWidowExilesUntilANonlandAndCountersWhenYouSayYes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	widow, spell, lands := mshA3WidowConnects(t, g)
	for _, id := range append([]uuid.UUID{spell}, lands...) {
		if !inExile(g, id) {
			t.Fatalf("card %s was not exiled", id)
		}
	}
	if opp.Library.Size() != 1 {
		t.Errorf("their library has %d cards left, want the one below the nonland card", opp.Library.Size())
	}
	mshA1AnswerConfirm(t, g, me.ID, true)
	if c, ok := battlefieldCard(g, widow); !ok || c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("Black Widow = %+v, want one +1/+1 counter", c)
	}
	if perm := g.CastPermissionOnCardByIDForEffect(spell); perm != nil {
		t.Errorf("the exiled card got a cast permission although the counter was chosen: %+v", perm)
	}
}

func TestMshA3BlackWidowDecliningTheCounterMayCastTheCardWithAnyTypeOfMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	widow, spell, _ := mshA3WidowConnects(t, g)
	mshA1AnswerConfirm(t, g, me.ID, false)
	if c, _ := battlefieldCard(g, widow); c.Counters[game.CounterPlusOne] != 0 {
		t.Error("a counter was put on Black Widow although it was declined")
	}
	perm := exiledPermission(g, spell)
	if perm.Player != me.ID || !perm.CastOnly || !perm.AnyType || !permissionLive(g, perm, me.ID) {
		t.Fatalf("%+v: its controller may cast it with mana of any type", perm)
	}
	// Post-combat main: pay {3}{U} with four mana of the wrong types.
	advanceTo(t, g, game.StepPostcombatMain)
	b06AddMana(me, "R", "R", "G", "C")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("casting the exiled card with off-type mana: %v", err)
	}
	if !g.Stack.Contains(spell) {
		t.Error("the card is not on the stack")
	}
}

func TestMshA3BlackWidowOnlyOffersTheNonlandCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_, _, lands := mshA3WidowConnects(t, g)
	mshA1AnswerConfirm(t, g, me.ID, false)
	for _, id := range lands {
		if perm := g.CastPermissionOnCardByIDForEffect(id); perm != nil {
			t.Errorf("an exiled land was given a permission: %+v", perm)
		}
	}
}

func TestMshA3BlackWidowCastingIsOnlyForTheRestOfTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_, spell, _ := mshA3WidowConnects(t, g)
	mshA1AnswerConfirm(t, g, me.ID, false)
	mshA3PassThroughTheTurnEnd(t, g)
	if perm := g.CastPermissionOnCardByIDForEffect(spell); perm != nil && permissionLive(g, perm, me.ID) {
		t.Errorf("the permission outlived the turn: %+v", perm)
	}
}

// --- Death to Our Enemies -------------------------------------------------------------------------------------

func TestMshA3DeathToOurEnemiesMakesATappedTreasureAndAPlanCounterPerNoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	plan := pushCatalogPermanent(g, me.ID, "Death to Our Enemies", "Enchantment — Plan", mshA3DeathEnemiesOracle, false)
	castNoncreature(t, g)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, plan); !ok || c.Counters["plan"] != 1 {
		t.Errorf("plan = %+v, want one plan counter", c)
	}
	var treasure *game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Treasure" {
			cc := c
			treasure = &cc
		}
	}
	if treasure == nil || !treasure.Tapped {
		t.Errorf("Treasure = %+v, want one tapped Treasure", treasure)
	}
}

func TestMshA3DeathToOurEnemiesIgnoresACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	plan := pushCatalogPermanent(g, me.ID, "Death to Our Enemies", "Enchantment — Plan", mshA3DeathEnemiesOracle, false)
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, plan); c.Counters["plan"] != 0 {
		t.Errorf("a creature spell put %d plan counters", c.Counters["plan"])
	}
}

func TestMshA3DeathToOurEnemiesFourthCounterSacrificesAndDealsSevenDivided(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	plan := pushCatalogPermanent(g, me.ID, "Death to Our Enemies", "Enchantment — Plan", mshA3DeathEnemiesOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(plan, "plan", 3) })
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(plan) {
		t.Fatal("sacrificed at the third counter")
	}
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 5)
	life := opp.Life
	castNoncreature(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(plan) {
		t.Fatal("the enchantment was not sacrificed at the fourth plan counter")
	}
	if latestPickTarget(g, me.ID) == nil {
		t.Fatalf("no target prompt for the 7 damage; pending %+v", g.PendingChoices)
	}
	pickTargetRefs(t, g, me.ID, map[uuid.UUID]int{wall: 4, opp.ID: 3},
		game.TargetRef{Kind: game.TargetCard, ID: wall}, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("the player took %d, want 3", got)
	}
	if c, ok := battlefieldCard(g, wall); ok && c.DamageMarked != 4 {
		t.Errorf("the wall has %d damage, want 4", c.DamageMarked)
	}
}

func TestMshA3DeathToOurEnemiesFiveCountersAtOnceStillPutTheFourth(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	plan := pushCatalogPermanent(g, me.ID, "Death to Our Enemies", "Enchantment — Plan", mshA3DeathEnemiesOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(plan, "plan", 5) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(plan) {
		t.Error("five counters at once put the fourth, so it should have been sacrificed")
	}
}

// --- Doom Reigns Supreme ---------------------------------------------------------------------------------------------

func TestMshA3DoomReignsDrainsAndCountsWhenAVillainEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	doom := pushCatalogPermanent(g, me.ID, "Doom Reigns Supreme", "Enchantment — Plan", mshA3DoomReignsOracle, false)
	myLife := me.Life
	lives := []int{opp.Life, g.Seats[2].Life, g.Seats[3].Life}
	b12PlayFromHand(t, g, "Test Villain", "Creature — Human Villain", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if me.Life != myLife+1 {
		t.Errorf("my life %d -> %d, want +1", myLife, me.Life)
	}
	for i, s := range g.Seats[1:] {
		if s.Life != lives[i]-1 {
			t.Errorf("opponent %d life %d -> %d, want -1", i+1, lives[i], s.Life)
		}
	}
	if c, _ := battlefieldCard(g, doom); c.Counters["plan"] != 1 {
		t.Errorf("plan counters = %d, want 1", c.Counters["plan"])
	}
}

func TestMshA3DoomReignsIgnoresANonVillainAndAnOpponentsVillain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	doom := pushCatalogPermanent(g, me.ID, "Doom Reigns Supreme", "Enchantment — Plan", mshA3DoomReignsOracle, false)
	b12PlayFromHand(t, g, "Test Hero", "Creature — Human Hero", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	theirs := b27Push(g, opp.ID, "Their Villain", "Creature — Human Villain", "", "{2}", 2, 2)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: theirs}) })
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, doom); c.Counters["plan"] != 0 {
		t.Errorf("plan counters = %d, want 0", c.Counters["plan"])
	}
}

// mshA3DoomFifthCounter sets up Doom Reigns Supreme at four counters, an
// opponent library of seven known cards on top, and puts the fifth counter on.
func mshA3DoomFifthCounter(t *testing.T, g *game.Game) (doom uuid.UUID, exiled []uuid.UUID, land uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	doom = pushCatalogPermanent(g, me.ID, "Doom Reigns Supreme", "Enchantment — Plan", mshA3DoomReignsOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(doom, "plan", 4) })
	opp.Library.Cards = nil
	ids := sunbirdLibrary(opp,
		game.Card{Name: "Exiled Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Exiled Wurm", TypeLine: "Creature — Wurm", ManaCost: "{4}{G}", Power: 5, Toughness: 5},
		game.Card{Name: "Exiled Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Exiled Ogre", TypeLine: "Creature — Ogre", ManaCost: "{2}{R}", Power: 3, Toughness: 3},
		game.Card{Name: "Exiled Imp", TypeLine: "Creature — Imp", ManaCost: "{B}", Power: 1, Toughness: 1},
		game.Card{Name: "Stays Put", TypeLine: "Creature — Bear", ManaCost: "{G}", Power: 1, Toughness: 1},
	)
	b12PlayFromHand(t, g, "Test Villain", "Creature — Human Villain", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickPlayer(t, g, me.ID, opp.ID)
		passPriorityAroundTable(t, g)
	}
	return doom, []uuid.UUID{ids[0], ids[1], ids[3], ids[4]}, ids[2]
}

func TestMshA3DoomReignsFifthCounterExilesFiveAndCastsUpToTwoFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doom, exiled, land := mshA3DoomFifthCounter(t, g)
	if g.Battlefield.Contains(doom) {
		t.Fatal("the enchantment was not sacrificed at the fifth plan counter")
	}
	for _, id := range append([]uuid.UUID{land}, exiled...) {
		if !inExile(g, id) {
			t.Errorf("card %s was not exiled from the opponent's library", id)
		}
	}
	if opp.Library.Size() != 1 {
		t.Errorf("their library has %d cards, want the one beyond the top five", opp.Library.Size())
	}
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, exiled[0], exiled[1])
	if err := castThroughResolutionCast(t, g, me.ID, exiled[1], game.CastSpellParams{}); err != nil {
		t.Fatalf("first free cast: %v", err)
	}
	if err := g.CastSpell(me.ID, exiled[0], game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("second free cast: %v", err)
	}
	if err := g.CastSpell(me.ID, exiled[2], game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("a third spell was cast from among the exiled cards")
	}
}

func TestMshA3DoomReignsDecliningWithdrawsTheCasts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_, exiled, _ := mshA3DoomFifthCounter(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, exiled[0])
	declineResolutionCast(t, g, me.ID)
	if err := g.CastSpell(me.ID, exiled[0], game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("a card was cast after the offer was declined")
	}
	if !inExile(g, exiled[0]) {
		t.Error("the declined card left exile")
	}
}

func TestMshA3DoomReignsNeverCastsTheExiledLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_, _, land := mshA3DoomFifthCounter(t, g)
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("the exiled land was cast")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(land); perm != nil {
		t.Errorf("the exiled land has a permission: %+v", perm)
	}
}
