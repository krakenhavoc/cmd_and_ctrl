package game

import (
	"testing"

	"github.com/google/uuid"
)

// council_vote_zone_chooser_test.go — ADR 0146's amendment of
// 2026-10-13 (#2928): a vote over cards that are not permanents
// (Custodi Squire), and "you choose how each player votes this turn"
// (Illusion of Choice).

// zoneChooserBallots collects each finished vote's ballots.
var zoneChooserBallots [][]VoteBallot

var zoneChooserThen = RegisterVoteThen(testEffectKeyPrefix+"zone-chooser", func(_ *Game, r VoteResult) error {
	zoneChooserBallots = append(zoneChooserBallots, r.Ballots)
	return nil
})

func startTestVote(t *testing.T, g *Game, p VotePrompt) {
	t.Helper()
	zoneChooserBallots = nil
	p.Then = zoneChooserThen
	if p.Question == "" {
		p.Question = "test vote"
	}
	if p.Starting == uuid.Nil {
		p.Starting = g.Seats[0].ID
	}
	p.Controller = p.Starting
	if p.Source == uuid.Nil {
		// A card the vote's controller controls, so a ballot dropped
		// for a player who left still has an object to finish.
		p.Source = departureTestSource(g, p.Controller, "Test Voter")
	}
	g.WithWriteLock(func() {
		if err := g.StartVoteForEffect(p); err != nil {
			t.Fatalf("StartVoteForEffect: %v", err)
		}
	})
}

func openBallot(g *Game) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.CouncilVote != nil {
			return c
		}
	}
	return nil
}

func castTestBallot(t *testing.T, g *Game, c *PendingChoice, label string) {
	t.Helper()
	for i, o := range c.PickOptions {
		if o.Label == label {
			if err := g.ResolveOptionPick(c.ID, c.Chooser, i); err != nil {
				t.Fatalf("ResolveOptionPick %q: %v", label, err)
			}
			return
		}
	}
	t.Fatalf("ballot offers no %q: %+v", label, c.PickOptions)
}

func graveyardCardNamed(p *Player, name string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(Card{InstanceID: id, Name: name, TypeLine: "Creature — Test", Owner: p.ID, Controller: p.ID})
	return id
}

// An object vote over graveyard cards offers the cards in the
// graveyard, and drops one that has left it (CR 400.7), where the
// battlefield rule would have dropped every one of them.
func TestAVoteOverGraveyardCardsOffersWhatIsStillThere(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	a, b := graveyardCardNamed(me, "Alpha"), graveyardCardNamed(me, "Beta")
	startTestVote(t, g, VotePrompt{
		Options:    []ChoiceOption{{Label: "Alpha", Cards: []uuid.UUID{a}}, {Label: "Beta", Cards: []uuid.UUID{b}}},
		OptionZone: ZoneGraveyard,
	})
	c := openBallot(g)
	if c == nil {
		t.Fatal("no ballot opened: the graveyard cards were treated as battlefield permanents")
	}
	if len(c.PickOptions) != 2 || c.CouncilVote.Zone != ZoneGraveyard {
		t.Fatalf("first ballot offers %d options in zone %q, want both graveyard cards", len(c.PickOptions), c.CouncilVote.Zone)
	}

	// Alpha leaves the graveyard while the players vote; the ballot
	// asked next no longer offers it.
	g.WithWriteLock(func() { g.ExileCardsForEffect([]uuid.UUID{a}) })
	castTestBallot(t, g, c, "Beta")
	c = openBallot(g)
	if c == nil || len(c.PickOptions) != 1 || c.PickOptions[0].Label != "Beta" {
		t.Fatalf("second ballot should offer only Beta, got %+v", c)
	}
}

// With the zone unset a card option is a battlefield permanent, as it
// has been since the vote shipped: a card in a graveyard is not offered.
func TestAnUnsetVoteZoneStillMeansTheBattlefield(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a := graveyardCardNamed(g.Seats[0], "Alpha")
	startTestVote(t, g, VotePrompt{Options: []ChoiceOption{{Label: "Alpha", Cards: []uuid.UUID{a}}}})
	if c := openBallot(g); c != nil {
		t.Errorf("a graveyard card was offered in a battlefield vote: %+v", c.PickOptions)
	}
}

// Illusion of Choice: every ballot is put to the chooser, in the
// voters' own turn order, and each is recorded as the voter's.
func TestAChosenVoteIsAnsweredByTheChooserForEachVoter(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser := g.Seats[2]
	g.WithWriteLock(func() {
		if !g.ChooseHowEachPlayerVotesThisTurnForEffect(uuid.Nil, chooser.ID, "Illusion of Choice") {
			t.Fatal("the record was not written")
		}
	})
	startTestVote(t, g, VotePrompt{Options: []ChoiceOption{{Label: "yes"}, {Label: "no"}}})
	for i, seat := range g.Seats {
		c := openBallot(g)
		if c == nil {
			t.Fatalf("ballot %d: none open", i)
		}
		if c.Chooser != chooser.ID || c.CouncilVote.Voter != seat.ID {
			t.Fatalf("ballot %d is owed by %s for voter %s, want the chooser for seat %d", i, c.Chooser, c.CouncilVote.Voter, i)
		}
		castTestBallot(t, g, c, map[bool]string{true: "yes", false: "no"}[i%2 == 0])
	}
	if len(zoneChooserBallots) != 1 || len(zoneChooserBallots[0]) != 4 {
		t.Fatalf("the vote did not finish with four ballots: %+v", zoneChooserBallots)
	}
	for i, b := range zoneChooserBallots[0] {
		if b.Voter != g.Seats[i].ID {
			t.Errorf("ballot %d names %s as the voter, want seat %d: the chooser's vote is the voter's", i, b.Voter, i)
		}
	}
	// The log names the voter, not the chooser, on every ballot.
	logged := 0
	for _, ev := range g.Events {
		if ev.Kind == EventVoteCast {
			if ev.Actor != g.Seats[logged].ID {
				t.Errorf("vote %d is logged for %s, want seat %d", logged, ev.Actor, logged)
			}
			logged++
		}
	}
}

// Two records stand: the newer one chooses.
func TestTheNewestChooserWins(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.WithWriteLock(func() {
		g.ChooseHowEachPlayerVotesThisTurnForEffect(uuid.Nil, g.Seats[1].ID, "first")
		g.ChooseHowEachPlayerVotesThisTurnForEffect(uuid.Nil, g.Seats[3].ID, "second")
	})
	startTestVote(t, g, VotePrompt{Options: []ChoiceOption{{Label: "yes"}, {Label: "no"}}})
	if c := openBallot(g); c == nil || c.Chooser != g.Seats[3].ID {
		t.Errorf("the newer record should choose: %+v", c)
	}
}

// A chooser who leaves the game stops choosing: their open ballot goes
// back to the voter and the rest of the vote is cast by the voters.
func TestAChooserWhoLeavesStopsChoosing(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser := g.Seats[2]
	g.WithWriteLock(func() { g.ChooseHowEachPlayerVotesThisTurnForEffect(uuid.Nil, chooser.ID, "Illusion of Choice") })
	startTestVote(t, g, VotePrompt{Options: []ChoiceOption{{Label: "yes"}, {Label: "no"}}})
	castTestBallot(t, g, openBallot(g), "yes") // seat 0's ballot, chosen
	if err := g.Concede(chooser.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	c := openBallot(g)
	if c == nil {
		t.Fatalf("the vote stalled when its chooser left: %+v", g.PendingChoices)
	}
	if c.CouncilVote.Voter != g.Seats[1].ID || c.Chooser != g.Seats[1].ID {
		t.Fatalf("seat 1's ballot is owed by %s (voter %s), want seat 1 themself", c.Chooser, c.CouncilVote.Voter)
	}
}

// A voter who leaves while the chooser holds their ballot casts nothing
// (CR 800.4a): the chooser's answer is dropped, and the vote moves on.
func TestAVotersChosenBallotIsDroppedWhenTheyLeave(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser, leaver := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { g.ChooseHowEachPlayerVotesThisTurnForEffect(uuid.Nil, chooser.ID, "Illusion of Choice") })
	startTestVote(t, g, VotePrompt{Options: []ChoiceOption{{Label: "yes"}, {Label: "no"}}})
	castTestBallot(t, g, openBallot(g), "yes") // seat 0 (the chooser) for themself
	if c := openBallot(g); c == nil || c.CouncilVote.Voter != leaver.ID {
		t.Fatalf("seat 1's ballot should be open: %+v", c)
	}
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	c := openBallot(g)
	if c == nil || c.CouncilVote.Voter != g.Seats[2].ID {
		t.Fatalf("the ballot for the player who left should be withdrawn and seat 2's asked: %+v", c)
	}
	for c := openBallot(g); c != nil; c = openBallot(g) {
		castTestBallot(t, g, c, "no")
	}
	if len(zoneChooserBallots) != 1 {
		t.Fatalf("the vote did not finish: %+v", zoneChooserBallots)
	}
	for _, b := range zoneChooserBallots[0] {
		if b.Voter == leaver.ID {
			t.Errorf("a ballot was recorded for the player who left: %+v", b)
		}
	}
}
