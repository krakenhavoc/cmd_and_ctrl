package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

func controlledByName(g *game.Game, p *game.Player, name string) bool {
	return votingACountNamed(g, p, name) > 0
}

func castSelvalas(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Selvala's Stampede", "Sorcery", effectstest.SelvalasStampedeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
}

// ---- Selvala's Stampede ---------------------------------------------

// Wild: reveal until a creature card for each wild vote, and put every
// creature revealed onto the battlefield; the rest are shuffled back
// into the library. Three wild votes stop at the third creature, so the
// fourth stays in the library.
func TestSelvalasStampedeRevealsUntilACreatureForEachWildVote(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.SetLibraryTopFirst(me,
		effectstest.LibraryLand("Forest A"), effectstest.LibraryCreature("Hit 1"), effectstest.LibraryLand("Forest B"),
		effectstest.LibraryCreature("Hit 2"), effectstest.LibraryCreature("Hit 3"), effectstest.LibraryCreature("Hit 4"), effectstest.LibraryLand("Forest C"))
	castSelvalas(t, g)
	effectstest.VoteRound(t, g, "wild", "wild", "wild", "free")

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
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a hand with no permanent card was asked to put one onto the battlefield")
	}
}

// With fewer creature cards than wild votes the whole library is
// revealed and the creatures it had enter.
func TestSelvalasStampedeRunsOutOfLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.SetLibraryTopFirst(me, effectstest.LibraryLand("Forest A"), effectstest.LibraryCreature("Only Hit"), effectstest.LibraryLand("Forest B"))
	castSelvalas(t, g)
	effectstest.VoteRound(t, g, "wild", "wild", "wild", "wild")
	if !controlledByName(g, me, "Only Hit") {
		t.Error("the one creature card in the library should have entered")
	}
	if got := me.Library.Size(); got != 2 {
		t.Errorf("library has %d cards, want the 2 lands", got)
	}
}

// No wild votes: nothing is revealed and nothing is shuffled.
func TestSelvalasStampedeWithNoWildVoteRevealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	ids := effectstest.SetLibraryTopFirst(me, effectstest.LibraryCreature("Hit 1"), effectstest.LibraryCreature("Hit 2"))
	castSelvalas(t, g)
	effectstest.VoteRound(t, g, "free", "free", "free", "free")
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
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.SetLibraryTopFirst(me, effectstest.LibraryLand("Forest A"))
	rock := effectstest.HandCardForTest(me, "Rock", "Artifact", "")
	bear := effectstest.HandCardForTest(me, "Bear", "Creature — Bear", "")
	elf := effectstest.HandCardForTest(me, "Elf", "Creature — Elf", "")
	bolt := effectstest.HandCardForTest(me, "Bolt", "Instant", "")
	castSelvalas(t, g)
	effectstest.VoteRound(t, g, "wild", "free", "free", "free")

	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("three free votes and three permanent cards in hand: the caster was not asked")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 3 {
		t.Errorf("bounds %d..%d, want 0..3 — \"you MAY put a permanent card ... for each free vote\"", pick.ChooseMin, pick.ChooseMax)
	}
	if effectstest.HasID(pick.ChooseCards, bolt) {
		t.Error("an instant was offered: it cannot become a permanent")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, rock, bear)
	if !controlledByName(g, me, "Rock") || !controlledByName(g, me, "Bear") {
		t.Error("the two picked cards should be on the battlefield")
	}
	if controlledByName(g, me, "Elf") || effectstest.ZoneKindOf(g, elf) != game.ZoneHand {
		t.Error("the card not picked should stay in hand")
	}
}

// The prompt is capped by the cards that could be put: two free votes
// and one permanent card in hand is a prompt of at most one, and
// declining is allowed.
func TestSelvalasStampedeFreeCapsAtTheCardsInHandAndMayBeDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.SetLibraryTopFirst(me, effectstest.LibraryLand("Forest A"))
	rock := effectstest.HandCardForTest(me, "Rock", "Artifact", "")
	castSelvalas(t, g)
	effectstest.VoteRound(t, g, "free", "free", "wild", "wild")
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || pick.ChooseMax != 1 || pick.ChooseMin != 0 {
		t.Fatalf("expected a 0..1 prompt, got %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID)
	if effectstest.ZoneKindOf(g, rock) != game.ZoneHand {
		t.Error("declining put the card onto the battlefield")
	}
}

// The wild half finishes before the free half is asked, and a creature
// that entered is on the battlefield by then.
func TestSelvalasStampedeWildResolvesBeforeFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.SetLibraryTopFirst(me, effectstest.LibraryCreature("Hit 1"), effectstest.LibraryLand("Forest A"))
	effectstest.HandCardForTest(me, "Rock", "Artifact", "")
	castSelvalas(t, g)
	effectstest.VoteRound(t, g, "wild", "free", "free", "free")
	if !controlledByName(g, me, "Hit 1") {
		t.Error("the wild creature was not on the battlefield when the free put was asked")
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		t.Error("the free put was never asked")
	}
}

// ---- Custodi Squire --------------------------------------------------

func castCustodiSquire(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Custodi Squire", "Creature — Spirit Cleric", effectstest.CustodiSquireOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
}

// The options are the artifact, creature and enchantment cards in the
// caster's graveyard; every card tied for the most votes returns to hand.
func TestCustodiSquireReturnsEveryCardTiedForMostVotes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	bear := effectstest.PushGraveyardTyped(me, "Bear", "Creature — Bear")
	rock := effectstest.PushGraveyardTyped(me, "Rock", "Artifact")
	aura := effectstest.PushGraveyardTyped(me, "Pacifism", "Enchantment — Aura")
	bolt := effectstest.PushGraveyardTyped(me, "Bolt", "Instant")
	land := effectstest.PushGraveyardTyped(me, "Forest", "Basic Land — Forest")
	oppCard := effectstest.PushGraveyardTyped(order[1], "Opponent Bear", "Creature — Bear")
	castCustodiSquire(t, g)

	c := effectstest.BallotFor(g, me.ID)
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
	effectstest.VoteFor(t, g, order[0].ID, "Bear")
	effectstest.VoteFor(t, g, order[1].ID, "Rock")
	effectstest.VoteFor(t, g, order[2].ID, "Bear")
	effectstest.VoteFor(t, g, order[3].ID, "Rock")
	if !effectstest.InHandOf(me, bear) || !effectstest.InHandOf(me, rock) {
		t.Error("both cards tied for most votes should have returned to hand")
	}
	if effectstest.InHandOf(me, aura) || effectstest.InHandOf(me, bolt) || effectstest.InHandOf(me, land) || effectstest.InHandOf(order[1], oppCard) {
		t.Error("a card nobody voted for returned")
	}
}

// A card that leaves the graveyard while the players vote is not
// offered to the next ballot and does not return.
func TestCustodiSquireDropsACardThatLeftTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	bear := effectstest.PushGraveyardTyped(me, "Bear", "Creature — Bear")
	rock := effectstest.PushGraveyardTyped(me, "Rock", "Artifact")
	castCustodiSquire(t, g)

	g.WithWriteLock(func() { g.ExileCardsForEffect([]uuid.UUID{bear}) })
	effectstest.VoteFor(t, g, me.ID, "Rock")
	c := effectstest.BallotFor(g, order[1].ID)
	if c == nil {
		t.Fatal("the second ballot never opened")
	}
	for _, o := range c.PickOptions {
		if o.Label == "Bear" {
			t.Error("a card that left the graveyard is still on the next ballot")
		}
	}
	effectstest.VoteFor(t, g, order[1].ID, "Rock")
	effectstest.VoteFor(t, g, order[2].ID, "Rock")
	effectstest.VoteFor(t, g, order[3].ID, "Rock")
	if !effectstest.InHandOf(me, rock) {
		t.Error("Rock should have returned")
	}
	if effectstest.InHandOf(me, bear) {
		t.Error("the exiled Bear came back to hand")
	}
}

// An empty graveyard: nobody votes and nothing happens.
func TestCustodiSquireWithNothingToVoteFor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	effectstest.PushGraveyardTyped(order[0], "Bolt", "Instant")
	castCustodiSquire(t, g)
	if effectstest.BallotFor(g, order[0].ID) != nil {
		t.Error("a ballot opened with no artifact, creature or enchantment card in the graveyard")
	}
}

// An open graveyard vote is a restore point, and the restored vote
// still knows its cards are in a graveyard.
func TestCustodiSquireVoteSurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	bear := effectstest.PushGraveyardTyped(me, "Bear", "Creature — Bear")
	effectstest.PushGraveyardTyped(me, "Rock", "Artifact")
	castCustodiSquire(t, g)
	effectstest.VoteFor(t, g, order[0].ID, "Bear")
	restored := effectstest.RestoreRoundTrip(t, g, true)
	c := effectstest.BallotFor(restored, order[1].ID)
	if c == nil || c.CouncilVote.Zone != game.ZoneGraveyard {
		t.Fatalf("the restored ballot lost its zone: %+v", c)
	}
	effectstest.VoteFor(t, restored, order[1].ID, "Bear")
	effectstest.VoteFor(t, restored, order[2].ID, "Bear")
	effectstest.VoteFor(t, restored, order[3].ID, "Bear")
	if !effectstest.InHandOf(restored.PlayerByIDForEffect(me.ID), bear) {
		t.Error("the restored vote did not return the card")
	}
}

// ---- Illusion of Choice ----------------------------------------------

func castIllusion(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Illusion of Choice", "Instant", effectstest.IllusionOfChoiceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
}

// The caster answers every ballot of a vote this turn, in the voters'
// turn order, and each is the voter's vote.
func TestIllusionOfChoiceLetsTheCasterCastEveryBallot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	hand := effectstest.HandSize(me)
	castIllusion(t, g)
	if effectstest.HandSize(me) != hand+1-0 && effectstest.HandSize(me) < hand {
		t.Errorf("Illusion of Choice should draw a card: hand %d → %d", hand, effectstest.HandSize(me))
	}
	effectstest.CastPlea(t, g)

	// The caster is asked for all four ballots; the others none.
	wantVoters := []*game.Player{order[0], order[1], order[2], order[3]}
	labels := []string{"time", "time", "knowledge", "time"}
	for i, voter := range wantVoters {
		c := effectstest.BallotFor(g, me.ID)
		if c == nil {
			t.Fatalf("ballot %d: the caster is not asked", i)
		}
		if c.CouncilVote.Voter != voter.ID {
			t.Fatalf("ballot %d is for %s, want %s", i, c.CouncilVote.Voter, voter.ID)
		}
		for j := 1; j < len(order); j++ {
			if effectstest.BallotFor(g, order[j].ID) != nil {
				t.Fatalf("ballot %d: %s was asked to vote though the caster chooses", i, order[j].Name)
			}
		}
		effectstest.VoteFor(t, g, me.ID, labels[i])
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
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.B43Catalog(g, order[1].ID, "Ballot Broker", "Creature — Human Advisor", effectstest.BallotBrokerOracle, 2, 3)
	castIllusion(t, g)
	effectstest.CastPlea(t, g)
	effectstest.VoteFor(t, g, me.ID, "time") // seat 0
	effectstest.VoteFor(t, g, me.ID, "time") // seat 1, required
	c := effectstest.BallotFor(g, me.ID)
	if c == nil || c.CouncilVote.Voter != order[1].ID || !c.CouncilVote.Optional() {
		t.Fatalf("seat 1's extra vote should be offered to the caster: %+v", c)
	}
	effectstest.VoteFor(t, g, me.ID, game.VoteDeclineLabel)
	if c := effectstest.BallotFor(g, me.ID); c == nil || c.CouncilVote.Voter != order[2].ID {
		t.Fatalf("seat 2's ballot should follow: %+v", c)
	}
}

// The rule lasts the turn: a vote on a later turn is answered by the
// voters again.
func TestIllusionOfChoiceEndsWithTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	castIllusion(t, g)
	toUpkeepOf(t, g, order[1]) // the next player's turn
	effectstest.CastPlea(t, g)
	if c := effectstest.BallotFor(g, order[1].ID); c == nil || c.CouncilVote.Voter != order[1].ID {
		t.Fatalf("the next turn's vote should be cast by its own voters: %+v", g.PendingChoices)
	}
}

// A vote that begins later the same turn is covered too, and the
// record is a restore point.
func TestIllusionOfChoiceSurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	castIllusion(t, g)
	restored := effectstest.RestoreRoundTrip(t, g, true)
	effectstest.CastPlea(t, restored)
	c := effectstest.BallotFor(restored, order[0].ID)
	if c == nil {
		t.Fatalf("the restored game lost the chooser: %+v", restored.PendingChoices)
	}
}

// Two casts: the newer chooses.
func TestIllusionOfChoiceTheLastCasterChooses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	g.WithWriteLock(func() {
		g.ChooseHowEachPlayerVotesThisTurnForEffect(uuid.Nil, order[1].ID, "Illusion of Choice")
	})
	castIllusion(t, g)
	effectstest.CastPlea(t, g)
	if effectstest.BallotFor(g, order[0].ID) == nil || effectstest.BallotFor(g, order[1].ID) != nil {
		t.Errorf("the later cast (seat 0) should choose, not the earlier one: %+v", g.PendingChoices)
	}
}

// ---- The Valeyard -----------------------------------------------------

// "While voting, you may vote an additional time": an optional second
// ballot for its controller, which may be declined.
func TestTheValeyardOffersAnExtraVote(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.B43Catalog(g, me.ID, "The Valeyard", "Legendary Creature — Time Lord Noble", effectstest.TheValeyardOracle, 4, 5)
	effectstest.CastPlea(t, g)
	effectstest.VoteFor(t, g, me.ID, "time")
	c := effectstest.BallotFor(g, me.ID)
	if c == nil || !c.CouncilVote.Optional() {
		t.Fatalf("The Valeyard's controller should be offered one more ballot: %+v", c)
	}
	effectstest.VoteFor(t, g, me.ID, "time")
	effectstest.VoteFor(t, g, order[1].ID, "knowledge")
	effectstest.VoteFor(t, g, order[2].ID, "knowledge")
	effectstest.VoteFor(t, g, order[3].ID, "knowledge")
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

func TestTheValeyardIsFullNow(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.TheValeyardOracle)
	if !ok {
		t.Fatal("The Valeyard is not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 || !spec.OpponentsFaceVillainousChoicesAgain {
		t.Errorf("both abilities are implemented (#3140): %v %v %v", spec.Completeness, spec.Caveats, spec.OpponentsFaceVillainousChoicesAgain)
	}
}

// ---- Trial of a Time Lord ---------------------------------------------

// trialThroughChapterThree casts the Saga for the active player and
// runs chapters I to III, each exiling one of the opponents' nontoken
// creatures, then advances into chapter IV and lets its vote begin.
// Returns the Saga and the three creatures.
func trialThroughChapterThree(t *testing.T, g *game.Game) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	order := effectstest.VoteOrder(g)
	seat := g.Turn.ActiveSeat
	ids := []uuid.UUID{
		effectstest.PushVanillaCreature(g, order[1].ID, "Alpha", 2, 2),
		effectstest.PushVanillaCreature(g, order[1].ID, "Beta", 2, 2),
		effectstest.PushVanillaCreature(g, order[2].ID, "Gamma", 2, 2),
	}
	sagaID := effectstest.CastCatalogSpell(t, g, "Trial of a Time Lord", "Enchantment — Saga", effectstest.TrialOfATimeLordOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAnyPendingTargetPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 2; i++ {
		effectstest.AdvanceToPrecombatMainOf(t, g, seat)
		effectstest.AnswerAnyPendingTargetPrompts(t, g)
		effectstest.PassPriorityAroundTable(t, g)
	}
	for _, id := range ids {
		if effectstest.ZoneKindOf(g, id) != game.ZoneExile {
			t.Fatalf("chapters I-III should have exiled all three creatures, %s is in %q", id, effectstest.ZoneKindOf(g, id))
		}
	}
	if got := effectstest.LoreCountersOn(g, sagaID); got != 3 {
		t.Fatalf("lore counters = %d, want 3", got)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat) // chapter IV
	effectstest.PassPriorityAroundTable(t, g)
	return sagaID, ids
}

// Guilty gets more votes: the owner of each card exiled with the Saga
// puts it on the bottom of their library (an owner of two orders them),
// and the Saga leaving then returns nothing.
func TestTrialOfATimeLordGuiltyTucksTheExiledCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	sagaID, ids := trialThroughChapterThree(t, g)
	if effectstest.BallotFor(g, order[0].ID) == nil {
		t.Fatalf("chapter IV did not start a vote: %+v", g.PendingChoices)
	}
	if effectstest.LoreCountersOn(g, sagaID) < 0 {
		t.Fatal("the Saga was sacrificed while its final chapter's vote was still open")
	}
	effectstest.VoteRound(t, g, "guilty", "guilty", "guilty", "innocent")

	// Seat 1 owns Alpha and Beta and orders them; seat 2's Gamma is alone.
	pick := effectstest.PutInLibraryChoiceFor(g, order[1].ID)
	if pick == nil || len(pick.ScryCards) != 2 {
		t.Fatalf("an owner of two cards should order them on the bottom: %+v", pick)
	}
	if effectstest.PutInLibraryChoiceFor(g, order[2].ID) != nil {
		t.Error("an owner of one card was asked to order it")
	}
	if err := g.ResolvePutInLibrary(pick.ID, order[1].ID, []uuid.UUID{ids[1], ids[0]}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoreCountersOn(g, sagaID) >= 0 {
		t.Fatal("the Saga was not sacrificed after chapter IV")
	}
	for _, id := range ids {
		if z := effectstest.ZoneKindOf(g, id); z != game.ZoneLibrary {
			t.Errorf("guilty: a creature is in %q, want the library", z)
		}
	}
	if got := effectstest.LibraryBottomIDs(order[2], 1); !effectstest.SameIDs(got, []uuid.UUID{ids[2]}) {
		t.Errorf("Gamma is not the bottom card of its owner's library: %v", got)
	}
	if got := effectstest.CountBattlefieldByName(g, "Alpha") + effectstest.CountBattlefieldByName(g, "Beta") + effectstest.CountBattlefieldByName(g, "Gamma"); got != 0 {
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
		g := effectstest.NewCatalogGame(t)
		sagaID, _ := trialThroughChapterThree(t, g)
		effectstest.VoteRound(t, g, labels...)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LoreCountersOn(g, sagaID) >= 0 {
			t.Fatalf("%s: the Saga was not sacrificed after chapter IV", name)
		}
		// Each returns as a new object (CR 610.3c, 400.7).
		for _, n := range []string{"Alpha", "Beta", "Gamma"} {
			if got := effectstest.CountBattlefieldByName(g, n); got != 1 {
				t.Errorf("%s: %d %s on the battlefield, want it back", name, got, n)
			}
		}
	}
}

// A card that has left exile before chapter IV is not "exiled with the
// Saga" any more: guilty does not reach for it.
func TestTrialOfATimeLordIgnoresACardThatLeftExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	sagaID, ids := trialThroughChapterThree(t, g)
	var back uuid.UUID
	g.WithWriteLock(func() {
		var err error
		if back, err = g.ReturnFromExileToBattlefieldForEffect(ids[0], order[1].ID, false); err != nil || back == uuid.Nil {
			t.Fatalf("setup: %v", err)
		}
	})
	effectstest.VoteRound(t, g, "guilty", "guilty", "guilty", "guilty")
	if pick := effectstest.PutInLibraryChoiceFor(g, order[1].ID); pick != nil {
		t.Errorf("Alpha left exile, so seat 1 has only Beta to tuck: %+v", pick)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoreCountersOn(g, sagaID) >= 0 {
		t.Fatal("the Saga was not sacrificed")
	}
	if z := effectstest.ZoneKindOf(g, back); z != game.ZoneBattlefield {
		t.Errorf("the card that left exile is in %q", z)
	}
	if z := effectstest.ZoneKindOf(g, ids[1]); z != game.ZoneLibrary {
		t.Errorf("the card still exiled with the Saga is in %q, want the library", z)
	}
}

// A ballot the caster casts for another voter (Illusion of Choice) is
// in the CASTER's move list, one answer per option, and the voter is
// offered nothing: a bot seat at that table is never left holding a live
// table with an empty list, and never offered a ballot that is not its
// own to answer.
func TestAChosenBallotIsEnumeratedForTheChooser(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	castIllusion(t, g)
	effectstest.CastPlea(t, g)
	effectstest.VoteFor(t, g, order[0].ID, "time") // the caster's own ballot
	// The second voter's ballot is the caster's to cast.
	c := effectstest.BallotFor(g, order[0].ID)
	if c == nil || c.CouncilVote.Voter != order[1].ID {
		t.Fatalf("setup: the second voter's ballot is not owed by the caster: %+v", c)
	}
	moves := effectstest.ResolveChoiceMoves(legal.EnumerateFor(g, order[0].ID))
	if len(moves) != 2 {
		t.Fatalf("the chooser is offered %d answers, want one per option (time, knowledge): %v", len(moves), moves)
	}
	effectstest.DispatchEveryMove(t, g, order[0].ID, moves)
	if got := legal.EnumerateFor(g, order[1].ID); len(effectstest.ResolveChoiceMoves(got)) != 0 {
		t.Errorf("the voter was offered a ballot the caster is casting for them: %v", got)
	}
}

// Selvala's free put: every set of up to N permanent cards in hand is an
// answer, the empty one included, and the engine accepts each.
func TestSelvalasStampedePutIsEnumeratedAsEverySetOfUpToN(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.SetLibraryTopFirst(me, effectstest.LibraryLand("Forest A"))
	effectstest.HandCardForTest(me, "Rock", "Artifact", "")
	effectstest.HandCardForTest(me, "Bear", "Creature — Bear", "")
	effectstest.HandCardForTest(me, "Elf", "Creature — Elf", "")
	effectstest.HandCardForTest(me, "Bolt", "Instant", "")
	castSelvalas(t, g)
	effectstest.VoteRound(t, g, "free", "free", "wild", "wild")
	moves := effectstest.ResolveChoiceMoves(legal.EnumerateFor(g, me.ID))
	// Subsets of three permanent cards of size 0..2: 1 + 3 + 3.
	if len(moves) != 7 {
		t.Fatalf("%d answers enumerated, want 7 (none, three singles, three pairs): %v", len(moves), moves)
	}
	effectstest.DispatchEveryMove(t, g, me.ID, moves)
}
