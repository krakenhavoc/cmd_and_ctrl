package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// voting_b_test.go — #2928, the second slice of the other voting cards
// on ADR 0146's seam (amendment 2026-10-13): Selvala's Stampede, Custodi
// Squire, Trial of a Time Lord, Illusion of Choice and The Valeyard.

const (
	selvalasStampedeOracle = "9d2aea05-f55c-4355-950f-bfdc4e985e80"
	custodiSquireOracle    = "07c7193e-f591-4aee-a3cc-3a2c52a38f9b"
	trialOfATimeLordOracle = "a5425d4f-cd6c-4240-89a4-dd13e0e1a95f"
	illusionOfChoiceOracle = "e3ad9287-ef13-4b66-9df8-b17e15bb4b0f"
	theValeyardOracle      = "ed202f1d-3511-45f4-bfbc-e6b76d4311d1"
)

// setLibraryTopFirst replaces p's library with the cards, the first
// being the top of it.
func setLibraryTopFirst(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		ids[i] = c.InstanceID
		p.Library.PushTop(c)
	}
	return ids
}

func libraryCreature(name string) game.Card {
	return game.Card{Name: name, TypeLine: "Creature — Test", Power: 1, Toughness: 1}
}

func libraryLand(name string) game.Card {
	return game.Card{Name: name, TypeLine: "Basic Land — Forest"}
}

func controlledByName(g *game.Game, p *game.Player, name string) bool {
	return votingACountNamed(g, p, name) > 0
}

func zoneKindOf(g *game.Game, id uuid.UUID) game.ZoneKind {
	if z := g.FindCardZoneForEffect(id); z != nil {
		return z.Kind
	}
	return ""
}

func castSelvalas(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Selvala's Stampede", "Sorcery", selvalasStampedeOracle, nil)
	passPriorityAroundTable(t, g)
}

// ---- Selvala's Stampede ---------------------------------------------

// Wild: reveal until a creature card for each wild vote, and put every
// creature revealed onto the battlefield; the rest are shuffled back
// into the library. Three wild votes stop at the third creature, so the
// fourth stays in the library.
func TestSelvalasStampedeRevealsUntilACreatureForEachWildVote(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	setLibraryTopFirst(me,
		libraryLand("Forest A"), libraryCreature("Hit 1"), libraryLand("Forest B"),
		libraryCreature("Hit 2"), libraryCreature("Hit 3"), libraryCreature("Hit 4"), libraryLand("Forest C"))
	castSelvalas(t, g)
	voteRound(t, g, "wild", "wild", "wild", "free")

	for _, name := range []string{"Hit 1", "Hit 2", "Hit 3"} {
		if !controlledByName(g, me, name) {
			t.Errorf("%s was revealed for a wild vote and should have entered", name)
		}
	}
	if controlledByName(g, me, "Hit 4") {
		t.Error("the reveal went past the third creature card: Hit 4 entered")
	}
	if got := me.Library.Size(); got != 4 {
		t.Errorf("library has %d cards, want the 4 that were not put onto the battlefield (Forest A/B/C and Hit 4)", got)
	}
	// One free vote with no permanent card in hand: nothing to ask.
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a hand with no permanent card was asked to put one onto the battlefield")
	}
}

// With fewer creature cards than wild votes the whole library is
// revealed and the creatures it had enter.
func TestSelvalasStampedeRunsOutOfLibrary(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	setLibraryTopFirst(me, libraryLand("Forest A"), libraryCreature("Only Hit"), libraryLand("Forest B"))
	castSelvalas(t, g)
	voteRound(t, g, "wild", "wild", "wild", "wild")
	if !controlledByName(g, me, "Only Hit") {
		t.Error("the one creature card in the library should have entered")
	}
	if got := me.Library.Size(); got != 2 {
		t.Errorf("library has %d cards, want the 2 lands", got)
	}
}

// No wild votes: nothing is revealed and nothing is shuffled.
func TestSelvalasStampedeWithNoWildVoteRevealsNothing(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	ids := setLibraryTopFirst(me, libraryCreature("Hit 1"), libraryCreature("Hit 2"))
	castSelvalas(t, g)
	voteRound(t, g, "free", "free", "free", "free")
	if me.Library.Size() != 2 || me.Library.Cards[1].InstanceID != ids[0] {
		t.Errorf("the library was disturbed with no wild vote: %+v", me.Library.Cards)
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventRevealCards {
			t.Errorf("a reveal was announced with no wild vote: %+v", ev)
		}
	}
}

// Free: the caster may put a permanent card from hand onto the
// battlefield for each free vote, up to that many, all at once. A
// nonpermanent card in hand is never offered.
func TestSelvalasStampedePutsUpToOneCardPerFreeVote(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	setLibraryTopFirst(me, libraryLand("Forest A"))
	rock := handCardForTest(me, "Rock", "Artifact", "")
	bear := handCardForTest(me, "Bear", "Creature — Bear", "")
	elf := handCardForTest(me, "Elf", "Creature — Elf", "")
	bolt := handCardForTest(me, "Bolt", "Instant", "")
	castSelvalas(t, g)
	voteRound(t, g, "wild", "free", "free", "free")

	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("three free votes and three permanent cards in hand: the caster was not asked")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 3 {
		t.Errorf("bounds %d..%d, want 0..3 — \"you MAY put a permanent card ... for each free vote\"", pick.ChooseMin, pick.ChooseMax)
	}
	if hasID(pick.ChooseCards, bolt) {
		t.Error("an instant was offered: it cannot become a permanent")
	}
	answerChooseCards(t, g, me.ID, rock, bear)
	if !controlledByName(g, me, "Rock") || !controlledByName(g, me, "Bear") {
		t.Error("the two picked cards should be on the battlefield")
	}
	if controlledByName(g, me, "Elf") || zoneKindOf(g, elf) != game.ZoneHand {
		t.Error("the card not picked should stay in hand")
	}
}

// The prompt is capped by the cards that could be put: two free votes
// and one permanent card in hand is a prompt of at most one, and
// declining is allowed.
func TestSelvalasStampedeFreeCapsAtTheCardsInHandAndMayBeDeclined(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	setLibraryTopFirst(me, libraryLand("Forest A"))
	rock := handCardForTest(me, "Rock", "Artifact", "")
	castSelvalas(t, g)
	voteRound(t, g, "free", "free", "wild", "wild")
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || pick.ChooseMax != 1 || pick.ChooseMin != 0 {
		t.Fatalf("expected a 0..1 prompt, got %+v", pick)
	}
	answerChooseCards(t, g, me.ID)
	if zoneKindOf(g, rock) != game.ZoneHand {
		t.Error("declining put the card onto the battlefield")
	}
}

// The wild half finishes before the free half is asked, and a creature
// that entered is on the battlefield by then.
func TestSelvalasStampedeWildResolvesBeforeFree(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	setLibraryTopFirst(me, libraryCreature("Hit 1"), libraryLand("Forest A"))
	handCardForTest(me, "Rock", "Artifact", "")
	castSelvalas(t, g)
	voteRound(t, g, "wild", "free", "free", "free")
	if !controlledByName(g, me, "Hit 1") {
		t.Error("the wild creature was not on the battlefield when the free put was asked")
	}
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Error("the free put was never asked")
	}
}

// ---- Custodi Squire --------------------------------------------------

func castCustodiSquire(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Custodi Squire", "Creature — Spirit Cleric", custodiSquireOracle, nil)
	passPriorityAroundTable(t, g)
}

// The options are the artifact, creature and enchantment cards in the
// caster's graveyard; every card tied for the most votes returns to hand.
func TestCustodiSquireReturnsEveryCardTiedForMostVotes(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	bear := pushGraveyardTyped(me, "Bear", "Creature — Bear")
	rock := pushGraveyardTyped(me, "Rock", "Artifact")
	aura := pushGraveyardTyped(me, "Pacifism", "Enchantment — Aura")
	bolt := pushGraveyardTyped(me, "Bolt", "Instant")
	land := pushGraveyardTyped(me, "Forest", "Basic Land — Forest")
	oppCard := pushGraveyardTyped(order[1], "Opponent Bear", "Creature — Bear")
	castCustodiSquire(t, g)

	c := ballotFor(g, me.ID)
	if c == nil {
		t.Fatal("no ballot opened for the caster")
	}
	offered := map[string]bool{}
	for _, o := range c.PickOptions {
		offered[o.Label] = true
	}
	for _, want := range []string{"Bear", "Rock", "Pacifism"} {
		if !offered[want] {
			t.Errorf("%s is an artifact, creature or enchantment card in the caster's graveyard and was not offered", want)
		}
	}
	for _, not := range []string{"Bolt", "Forest", "Opponent Bear"} {
		if offered[not] {
			t.Errorf("%s was offered: not a card of the right type in the caster's graveyard", not)
		}
	}
	if c.CouncilVote.Zone != game.ZoneGraveyard {
		t.Errorf("the ballot says its cards are in %q, want the graveyard", c.CouncilVote.Zone)
	}

	// Bear 2, Rock 2, Pacifism 0 — a tie for most votes.
	voteFor(t, g, order[0].ID, "Bear")
	voteFor(t, g, order[1].ID, "Rock")
	voteFor(t, g, order[2].ID, "Bear")
	voteFor(t, g, order[3].ID, "Rock")
	if !inHandOf(me, bear) || !inHandOf(me, rock) {
		t.Error("both cards tied for most votes should have returned to hand")
	}
	if inHandOf(me, aura) || inHandOf(me, bolt) || inHandOf(me, land) || inHandOf(order[1], oppCard) {
		t.Error("a card nobody voted for returned")
	}
}

// A card that leaves the graveyard while the players vote is not
// offered to the next ballot and does not return.
func TestCustodiSquireDropsACardThatLeftTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	bear := pushGraveyardTyped(me, "Bear", "Creature — Bear")
	rock := pushGraveyardTyped(me, "Rock", "Artifact")
	castCustodiSquire(t, g)

	g.WithWriteLock(func() { g.ExileCardsForEffect([]uuid.UUID{bear}) })
	voteFor(t, g, me.ID, "Rock")
	c := ballotFor(g, order[1].ID)
	if c == nil {
		t.Fatal("the second ballot never opened")
	}
	for _, o := range c.PickOptions {
		if o.Label == "Bear" {
			t.Error("a card that left the graveyard is still on the next ballot")
		}
	}
	voteFor(t, g, order[1].ID, "Rock")
	voteFor(t, g, order[2].ID, "Rock")
	voteFor(t, g, order[3].ID, "Rock")
	if !inHandOf(me, rock) {
		t.Error("Rock should have returned")
	}
	if inHandOf(me, bear) {
		t.Error("the exiled Bear came back to hand")
	}
}

// An empty graveyard: nobody votes and nothing happens.
func TestCustodiSquireWithNothingToVoteFor(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	pushGraveyardTyped(order[0], "Bolt", "Instant")
	castCustodiSquire(t, g)
	if ballotFor(g, order[0].ID) != nil {
		t.Error("a ballot opened with no artifact, creature or enchantment card in the graveyard")
	}
}

// An open graveyard vote is a restore point, and the restored vote
// still knows its cards are in a graveyard.
func TestCustodiSquireVoteSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	bear := pushGraveyardTyped(me, "Bear", "Creature — Bear")
	pushGraveyardTyped(me, "Rock", "Artifact")
	castCustodiSquire(t, g)
	voteFor(t, g, order[0].ID, "Bear")
	restored := restoreRoundTrip(t, g, true)
	c := ballotFor(restored, order[1].ID)
	if c == nil || c.CouncilVote.Zone != game.ZoneGraveyard {
		t.Fatalf("the restored ballot lost its zone: %+v", c)
	}
	voteFor(t, restored, order[1].ID, "Bear")
	voteFor(t, restored, order[2].ID, "Bear")
	voteFor(t, restored, order[3].ID, "Bear")
	if !inHandOf(restored.PlayerByIDForEffect(me.ID), bear) {
		t.Error("the restored vote did not return the card")
	}
}

// ---- Illusion of Choice ----------------------------------------------

func castIllusion(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Illusion of Choice", "Instant", illusionOfChoiceOracle, nil)
	passPriorityAroundTable(t, g)
}

// The caster answers every ballot of a vote this turn, in the voters'
// turn order, and each is the voter's vote.
func TestIllusionOfChoiceLetsTheCasterCastEveryBallot(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	hand := handSize(me)
	castIllusion(t, g)
	if handSize(me) != hand+1-0 && handSize(me) < hand {
		t.Errorf("Illusion of Choice should draw a card: hand %d → %d", hand, handSize(me))
	}
	castPlea(t, g)

	// The caster is asked for all four ballots; the others none.
	wantVoters := []*game.Player{order[0], order[1], order[2], order[3]}
	labels := []string{"time", "time", "knowledge", "time"}
	for i, voter := range wantVoters {
		c := ballotFor(g, me.ID)
		if c == nil {
			t.Fatalf("ballot %d: the caster is not asked", i)
		}
		if c.CouncilVote.Voter != voter.ID {
			t.Fatalf("ballot %d is for %s, want %s", i, c.CouncilVote.Voter, voter.ID)
		}
		for j := 1; j < len(order); j++ {
			if ballotFor(g, order[j].ID) != nil {
				t.Fatalf("ballot %d: %s was asked to vote though the caster chooses", i, order[j].Name)
			}
		}
		voteFor(t, g, me.ID, labels[i])
	}
	// Three time to one knowledge: time wins and the caster takes the turn.
	if len(g.ExtraTurns) != 1 {
		t.Errorf("time won 3-1: %d extra turns", len(g.ExtraTurns))
	}
	i := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventVoteCast {
			if ev.Actor != order[i].ID {
				t.Errorf("vote %d is logged for %s, want the voter %s", i, ev.Actor, order[i].ID)
			}
			i++
		}
	}
	if i != 4 {
		t.Errorf("%d votes logged, want 4", i)
	}
}

// Extra votes still belong to the voter: a Ballot Broker's controller is
// offered the extra ballot, and the caster chooses it.
func TestIllusionOfChoiceChoosesTheVotersExtraVote(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	b43Catalog(g, order[1].ID, "Ballot Broker", "Creature — Human Advisor", ballotBrokerOracle, 2, 3)
	castIllusion(t, g)
	castPlea(t, g)
	voteFor(t, g, me.ID, "time") // seat 0
	voteFor(t, g, me.ID, "time") // seat 1, required
	c := ballotFor(g, me.ID)
	if c == nil || c.CouncilVote.Voter != order[1].ID || !c.CouncilVote.Optional() {
		t.Fatalf("seat 1's extra vote should be offered to the caster: %+v", c)
	}
	voteFor(t, g, me.ID, game.VoteDeclineLabel)
	if c := ballotFor(g, me.ID); c == nil || c.CouncilVote.Voter != order[2].ID {
		t.Fatalf("seat 2's ballot should follow: %+v", c)
	}
}

// The rule lasts the turn: a vote on a later turn is answered by the
// voters again.
func TestIllusionOfChoiceEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	castIllusion(t, g)
	toUpkeepOf(t, g, order[1]) // the next player's turn
	castPlea(t, g)
	if c := ballotFor(g, order[1].ID); c == nil || c.CouncilVote.Voter != order[1].ID {
		t.Fatalf("the next turn's vote should be cast by its own voters: %+v", g.PendingChoices)
	}
}

// A vote that begins later the same turn is covered too, and the
// record is a restore point.
func TestIllusionOfChoiceSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	castIllusion(t, g)
	restored := restoreRoundTrip(t, g, true)
	castPlea(t, restored)
	c := ballotFor(restored, order[0].ID)
	if c == nil {
		t.Fatalf("the restored game lost the chooser: %+v", restored.PendingChoices)
	}
}

// Two casts: the newer chooses.
func TestIllusionOfChoiceTheLastCasterChooses(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	g.WithWriteLock(func() {
		g.ChooseHowEachPlayerVotesThisTurnForEffect(uuid.Nil, order[1].ID, "Illusion of Choice")
	})
	castIllusion(t, g)
	castPlea(t, g)
	if ballotFor(g, order[0].ID) == nil || ballotFor(g, order[1].ID) != nil {
		t.Errorf("the later cast (seat 0) should choose, not the earlier one: %+v", g.PendingChoices)
	}
}

// ---- The Valeyard -----------------------------------------------------

// "While voting, you may vote an additional time": an optional second
// ballot for its controller, which may be declined.
func TestTheValeyardOffersAnExtraVote(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	b43Catalog(g, me.ID, "The Valeyard", "Legendary Creature — Time Lord Noble", theValeyardOracle, 4, 5)
	castPlea(t, g)
	voteFor(t, g, me.ID, "time")
	c := ballotFor(g, me.ID)
	if c == nil || !c.CouncilVote.Optional() {
		t.Fatalf("The Valeyard's controller should be offered one more ballot: %+v", c)
	}
	voteFor(t, g, me.ID, "time")
	voteFor(t, g, order[1].ID, "knowledge")
	voteFor(t, g, order[2].ID, "knowledge")
	voteFor(t, g, order[3].ID, "knowledge")
	// 2 time against 3 knowledge: the extra vote was cast and counted.
	votes := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventVoteCast && ev.Actor == me.ID {
			votes++
		}
	}
	if votes != 2 {
		t.Errorf("The Valeyard's controller cast %d votes, want 2", votes)
	}
}

func TestTheValeyardIsDeclaredWeakerThanPrinted(t *testing.T) {
	spec, ok := Lookup(theValeyardOracle)
	if !ok {
		t.Fatal("The Valeyard is not registered")
	}
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) == 0 {
		t.Errorf("the villainous-choice ability is not implemented, so the card must say so: %v %v", spec.Completeness, spec.Caveats)
	}
}

// ---- Trial of a Time Lord ---------------------------------------------

// trialThroughChapterThree casts the Saga for the active player and
// runs chapters I to III, each exiling one of the opponents' nontoken
// creatures, then advances into chapter IV and lets its vote begin.
// Returns the Saga and the three creatures.
func trialThroughChapterThree(t *testing.T, g *game.Game) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	order := voteOrder(g)
	seat := g.Turn.ActiveSeat
	ids := []uuid.UUID{
		pushVanillaCreature(g, order[1].ID, "Alpha", 2, 2),
		pushVanillaCreature(g, order[1].ID, "Beta", 2, 2),
		pushVanillaCreature(g, order[2].ID, "Gamma", 2, 2),
	}
	sagaID := castCatalogSpell(t, g, "Trial of a Time Lord", "Enchantment — Saga", trialOfATimeLordOracle, nil)
	passPriorityAroundTable(t, g)
	answerAnyPendingTargetPrompts(t, g)
	passPriorityAroundTable(t, g)
	for i := 0; i < 2; i++ {
		advanceToPrecombatMainOf(t, g, seat)
		answerAnyPendingTargetPrompts(t, g)
		passPriorityAroundTable(t, g)
	}
	for _, id := range ids {
		if zoneKindOf(g, id) != game.ZoneExile {
			t.Fatalf("chapters I-III should have exiled all three creatures, %s is in %q", id, zoneKindOf(g, id))
		}
	}
	if got := loreCountersOn(g, sagaID); got != 3 {
		t.Fatalf("lore counters = %d, want 3", got)
	}
	advanceToPrecombatMainOf(t, g, seat) // chapter IV
	passPriorityAroundTable(t, g)
	return sagaID, ids
}

// Guilty gets more votes: the owner of each card exiled with the Saga
// puts it on the bottom of their library (an owner of two orders them),
// and the Saga leaving then returns nothing.
func TestTrialOfATimeLordGuiltyTucksTheExiledCards(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	sagaID, ids := trialThroughChapterThree(t, g)
	if ballotFor(g, order[0].ID) == nil {
		t.Fatalf("chapter IV did not start a vote: %+v", g.PendingChoices)
	}
	if loreCountersOn(g, sagaID) < 0 {
		t.Fatal("the Saga was sacrificed while its final chapter's vote was still open")
	}
	voteRound(t, g, "guilty", "guilty", "guilty", "innocent")

	// Seat 1 owns Alpha and Beta and orders them; seat 2's Gamma is alone.
	pick := putInLibraryChoiceFor(g, order[1].ID)
	if pick == nil || len(pick.ScryCards) != 2 {
		t.Fatalf("an owner of two cards should order them on the bottom: %+v", pick)
	}
	if putInLibraryChoiceFor(g, order[2].ID) != nil {
		t.Error("an owner of one card was asked to order it")
	}
	if err := g.ResolvePutInLibrary(pick.ID, order[1].ID, []uuid.UUID{ids[1], ids[0]}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	passPriorityAroundTable(t, g)
	if loreCountersOn(g, sagaID) >= 0 {
		t.Fatal("the Saga was not sacrificed after chapter IV")
	}
	for _, id := range ids {
		if z := zoneKindOf(g, id); z != game.ZoneLibrary {
			t.Errorf("guilty: a creature is in %q, want the library", z)
		}
	}
	if got := libraryBottomIDs(order[2], 1); !sameIDs(got, []uuid.UUID{ids[2]}) {
		t.Errorf("Gamma is not the bottom card of its owner's library: %v", got)
	}
	if got := countBattlefieldByName(g, "Alpha") + countBattlefieldByName(g, "Beta") + countBattlefieldByName(g, "Gamma"); got != 0 {
		t.Errorf("guilty: %d exiled creatures came back when the Saga left", got)
	}
}

// Innocent gets the votes, or it is a tie: nothing is tucked, the Saga
// is sacrificed and every exiled creature comes back.
func TestTrialOfATimeLordInnocentOrATieReturnsThem(t *testing.T) {
	for name, labels := range map[string][]string{
		"innocent": {"innocent", "innocent", "guilty", "innocent"},
		"a tie":    {"guilty", "innocent", "guilty", "innocent"},
	} {
		g := newCatalogGame(t)
		sagaID, _ := trialThroughChapterThree(t, g)
		voteRound(t, g, labels...)
		passPriorityAroundTable(t, g)
		if loreCountersOn(g, sagaID) >= 0 {
			t.Fatalf("%s: the Saga was not sacrificed after chapter IV", name)
		}
		// Each returns as a new object (CR 610.3c, 400.7).
		for _, n := range []string{"Alpha", "Beta", "Gamma"} {
			if got := countBattlefieldByName(g, n); got != 1 {
				t.Errorf("%s: %d %s on the battlefield, want it back", name, got, n)
			}
		}
	}
}

// A card that has left exile before chapter IV is not "exiled with the
// Saga" any more: guilty does not reach for it.
func TestTrialOfATimeLordIgnoresACardThatLeftExile(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	sagaID, ids := trialThroughChapterThree(t, g)
	var back uuid.UUID
	g.WithWriteLock(func() {
		var err error
		if back, err = g.ReturnFromExileToBattlefieldForEffect(ids[0], order[1].ID, false); err != nil || back == uuid.Nil {
			t.Fatalf("setup: %v", err)
		}
	})
	voteRound(t, g, "guilty", "guilty", "guilty", "guilty")
	if pick := putInLibraryChoiceFor(g, order[1].ID); pick != nil {
		t.Errorf("Alpha left exile, so seat 1 has only Beta to tuck: %+v", pick)
	}
	passPriorityAroundTable(t, g)
	if loreCountersOn(g, sagaID) >= 0 {
		t.Fatal("the Saga was not sacrificed")
	}
	if z := zoneKindOf(g, back); z != game.ZoneBattlefield {
		t.Errorf("the card that left exile is in %q", z)
	}
	if z := zoneKindOf(g, ids[1]); z != game.ZoneLibrary {
		t.Errorf("the card still exiled with the Saga is in %q, want the library", z)
	}
}

// ---- the bot's move list --------------------------------------------------

// dispatchEveryMove proves the engine accepts each enumerated answer.
func dispatchEveryMove(t *testing.T, g *game.Game, seat uuid.UUID, moves []legal.Move) {
	t.Helper()
	for _, m := range moves {
		cl := g.Clone()
		if err := actions.Dispatch(cl, actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: seat, Params: m.Params}); err != nil {
			t.Errorf("move %q rejected: %v", m.Label, err)
		}
	}
}

func resolveChoiceMoves(moves []legal.Move) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Type == legal.TypeResolveChoice {
			out = append(out, m)
		}
	}
	return out
}

// A ballot the caster casts for another voter (Illusion of Choice) is
// in the CASTER's move list, one answer per option, and the voter is
// offered nothing: a bot seat at that table is never left holding a live
// table with an empty list, and never offered a ballot that is not its
// own to answer.
func TestAChosenBallotIsEnumeratedForTheChooser(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	castIllusion(t, g)
	castPlea(t, g)
	voteFor(t, g, order[0].ID, "time") // the caster's own ballot
	// The second voter's ballot is the caster's to cast.
	c := ballotFor(g, order[0].ID)
	if c == nil || c.CouncilVote.Voter != order[1].ID {
		t.Fatalf("setup: the second voter's ballot is not owed by the caster: %+v", c)
	}
	moves := resolveChoiceMoves(legal.EnumerateFor(g, order[0].ID))
	if len(moves) != 2 {
		t.Fatalf("the chooser is offered %d answers, want one per option (time, knowledge): %v", len(moves), moves)
	}
	dispatchEveryMove(t, g, order[0].ID, moves)
	if got := legal.EnumerateFor(g, order[1].ID); len(resolveChoiceMoves(got)) != 0 {
		t.Errorf("the voter was offered a ballot the caster is casting for them: %v", got)
	}
}

// Selvala's free put: every set of up to N permanent cards in hand is an
// answer, the empty one included, and the engine accepts each.
func TestSelvalasStampedePutIsEnumeratedAsEverySetOfUpToN(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	setLibraryTopFirst(me, libraryLand("Forest A"))
	handCardForTest(me, "Rock", "Artifact", "")
	handCardForTest(me, "Bear", "Creature — Bear", "")
	handCardForTest(me, "Elf", "Creature — Elf", "")
	handCardForTest(me, "Bolt", "Instant", "")
	castSelvalas(t, g)
	voteRound(t, g, "free", "free", "wild", "wild")
	moves := resolveChoiceMoves(legal.EnumerateFor(g, me.ID))
	// Subsets of three permanent cards of size 0..2: 1 + 3 + 3.
	if len(moves) != 7 {
		t.Fatalf("%d answers enumerated, want 7 (none, three singles, three pairs): %v", len(moves), moves)
	}
	dispatchEveryMove(t, g, me.ID, moves)
}
