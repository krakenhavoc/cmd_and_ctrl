package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// finish_voting_test.go — ADR 0146's amendment of 2026-10-13, #2927:
// "whenever players finish voting" and the three cards that wait on it.

const (
	grudgeKeeperOracle     = "ffbf8459-fda2-4c1d-9841-c001c37d2141"
	modelOfUnityOracle     = "d7b69fd7-e575-4009-8549-f576249ba641"
	erestorOfTheCouncilOID = "b65e1dff-9423-4840-a6af-78cb3e6f4f31"
)

// Plea for Power's two words, as option indexes.
const (
	pleaTime      = 0
	pleaKnowledge = 1
)

// votesFinished lists the game's EventVotesFinished events.
func votesFinished(g *game.Game) []game.Event {
	var out []game.Event
	for _, ev := range g.Events {
		if ev.Kind == game.EventVotesFinished {
			out = append(out, ev)
		}
	}
	return out
}

// castPleaAndVote casts Plea for Power for the active player and casts
// each seat's votes (labels, in the order cast), seat order from the
// caster. A seat with an extra vote lists one label per ballot.
func castPleaAndVote(t *testing.T, g *game.Game, votes [][]string) {
	t.Helper()
	castPlea(t, g)
	for i, p := range voteOrder(g) {
		for _, label := range votes[i] {
			voteFor(t, g, p.ID, label)
		}
	}
}

// emitFinishedVote announces a finished vote by hand, for the cases a
// real table cannot reach in one step (a seat that cast nothing).
func emitFinishedVote(g *game.Game, controller uuid.UUID, ballots ...game.VoteBallot) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventVotesFinished, Actor: controller, Amount: len(ballots), Ballots: ballots,
		})
	})
}

func ballot(p *game.Player, option int) game.VoteBallot {
	return game.VoteBallot{Voter: p.ID, Option: option}
}

// The event fires once per vote, after the last ballot, carrying every
// ballot in the order cast — including a player's two (CR 701.38d).
func TestVotesFinishedFiresOncePerVoteWithTheBallots(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	b43Catalog(g, order[1].ID, "Brago's Representative", "Creature — Human Advisor", bragosRepOracle, 1, 4)
	castPlea(t, g)
	if n := len(votesFinished(g)); n != 0 {
		t.Fatalf("%d finished-vote events before anyone voted", n)
	}
	voteFor(t, g, me.ID, "time")
	voteFor(t, g, order[1].ID, "knowledge")
	voteFor(t, g, order[1].ID, "time")
	voteFor(t, g, order[2].ID, "knowledge")
	if n := len(votesFinished(g)); n != 0 {
		t.Fatalf("%d finished-vote events with a ballot still open", n)
	}
	voteFor(t, g, order[3].ID, "knowledge")

	done := votesFinished(g)
	if len(done) != 1 {
		t.Fatalf("%d finished-vote events, want exactly 1", len(done))
	}
	ev := done[0]
	want := []game.VoteBallot{
		ballot(me, pleaTime),
		ballot(order[1], pleaKnowledge), ballot(order[1], pleaTime),
		ballot(order[2], pleaKnowledge),
		ballot(order[3], pleaKnowledge),
	}
	if !reflect.DeepEqual(ev.Ballots, want) {
		t.Errorf("ballots = %+v, want %+v", ev.Ballots, want)
	}
	if ev.Actor != me.ID || ev.Amount != 5 {
		t.Errorf("event %+v, want the caster as Actor and 5 ballots", ev)
	}
	var lastCast uint64
	for _, e := range g.Events {
		if e.Kind == game.EventVoteCast {
			lastCast = e.Seq
		}
	}
	if ev.Seq <= lastCast {
		t.Errorf("the finish (seq %d) comes before the last ballot (seq %d)", ev.Seq, lastCast)
	}

	// A second vote is a second event with its own ballots.
	castPlea(t, g)
	for _, p := range voteOrder(g) {
		voteFor(t, g, p.ID, "knowledge")
		if p.ID == order[1].ID {
			voteFor(t, g, p.ID, "knowledge")
		}
	}
	done = votesFinished(g)
	if len(done) != 2 || len(done[1].Ballots) != 5 || len(done[0].Ballots) != 5 {
		t.Fatalf("two votes, want two events of five ballots: %+v", done)
	}
	if done[1].Ballots[0].Option != pleaKnowledge {
		t.Errorf("the second vote's event carries the first vote's ballots: %+v", done[1].Ballots)
	}
}

// A vote with nothing to vote for never began, so nobody finished it.
func TestNoFinishedVoteEventWhenThereWasNothingToVoteFor(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Council's Judgment", "Sorcery", councilsJudgmentOracle, nil)
	passPriorityAroundTable(t, g)
	if n := len(votesFinished(g)); n != 0 {
		t.Errorf("Council's Judgment found no permanent to vote on, yet %d finished-vote events fired", n)
	}
}

// Grudge Keeper: each opponent who voted for a choice you didn't loses 2.
func TestGrudgeKeeper(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	b43Catalog(g, order[0].ID, "Grudge Keeper", "Creature — Zombie Wizard", grudgeKeeperOracle, 2, 1)
	start := lifeOf(g, order[0].ID)
	castPleaAndVote(t, g, [][]string{{"time"}, {"time"}, {"knowledge"}, {"knowledge"}})
	if n := len(g.PendingTriggers) + len(g.StackMeta); n == 0 {
		t.Fatal("no trigger went on the stack when the vote finished")
	}
	passPriorityAroundTable(t, g)
	got := []int{lifeOf(g, order[0].ID), lifeOf(g, order[1].ID), lifeOf(g, order[2].ID), lifeOf(g, order[3].ID)}
	want := []int{start, start, start - 2, start - 2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("life after the vote = %v, want %v (only the opponents who voted apart lose 2)", got, want)
	}
}

// The trigger is not "your" votes: it fires for a vote an opponent
// called, and "you" is the Grudge Keeper's controller.
func TestGrudgeKeeperTriggersOnAnotherPlayersVote(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	keeper := order[2]
	b43Catalog(g, keeper.ID, "Grudge Keeper", "Creature — Zombie Wizard", grudgeKeeperOracle, 2, 1)
	start := lifeOf(g, keeper.ID)
	// The caster (order[0]) and order[1] vote time; the controller votes
	// knowledge, as does order[3].
	castPleaAndVote(t, g, [][]string{{"time"}, {"time"}, {"knowledge"}, {"time"}})
	passPriorityAroundTable(t, g)
	got := []int{lifeOf(g, order[0].ID), lifeOf(g, order[1].ID), lifeOf(g, keeper.ID), lifeOf(g, order[3].ID)}
	want := []int{start - 2, start - 2, start, start - 2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("life = %v, want %v: everyone who did not vote knowledge loses 2", got, want)
	}
}

// CR 701.38d: a controller with two votes who split them took both
// choices, so nobody voted for one they didn't; an opponent with two
// votes who split them voted for a choice the controller didn't.
func TestGrudgeKeeperWithExtraVotes(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	b43Catalog(g, me.ID, "Grudge Keeper", "Creature — Zombie Wizard", grudgeKeeperOracle, 2, 1)
	b43Catalog(g, me.ID, "Brago's Representative", "Creature — Human Advisor", bragosRepOracle, 1, 4)
	start := lifeOf(g, me.ID)
	castPleaAndVote(t, g, [][]string{{"time", "knowledge"}, {"time"}, {"knowledge"}, {"time"}})
	passPriorityAroundTable(t, g)
	for _, p := range order[1:] {
		if got := lifeOf(g, p.ID); got != start {
			t.Errorf("%s lost life (%d → %d) though I voted for both choices", p.Name, start, got)
		}
	}

	// The same table with the votes the other way round: I voted time
	// twice, and the opponent with two votes took both choices.
	g = newCatalogGame(t)
	order = voteOrder(g)
	me = order[0]
	b43Catalog(g, me.ID, "Grudge Keeper", "Creature — Zombie Wizard", grudgeKeeperOracle, 2, 1)
	b43Catalog(g, order[1].ID, "Brago's Representative", "Creature — Human Advisor", bragosRepOracle, 1, 4)
	castPleaAndVote(t, g, [][]string{{"time"}, {"time", "knowledge"}, {"time"}, {"time"}})
	passPriorityAroundTable(t, g)
	got := []int{lifeOf(g, order[1].ID), lifeOf(g, order[2].ID), lifeOf(g, order[3].ID)}
	want := []int{start - 2, start, start}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("life = %v, want %v: only the player who split their votes voted for a choice I didn't", got, want)
	}
}

// A controller who cast no ballot voted for no choice, so every choice
// an opponent voted for is one they didn't vote for; an opponent who cast
// none voted for nothing and loses nothing.
func TestGrudgeKeeperControllerWhoDidNotVote(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	b43Catalog(g, me.ID, "Grudge Keeper", "Creature — Zombie Wizard", grudgeKeeperOracle, 2, 1)
	start := lifeOf(g, me.ID)
	emitFinishedVote(g, me.ID, ballot(order[1], pleaTime), ballot(order[2], pleaKnowledge))
	passPriorityAroundTable(t, g)
	got := []int{lifeOf(g, me.ID), lifeOf(g, order[1].ID), lifeOf(g, order[2].ID), lifeOf(g, order[3].ID)}
	want := []int{start, start - 2, start - 2, start}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("life = %v, want %v", got, want)
	}
}

// A vote that began and ended with no ballots (every voter left) is a
// finished vote nobody voted in: nothing happens.
func TestGrudgeKeeperEmptyVote(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	b43Catalog(g, order[0].ID, "Grudge Keeper", "Creature — Zombie Wizard", grudgeKeeperOracle, 2, 1)
	start := lifeOf(g, order[1].ID)
	emitFinishedVote(g, order[0].ID)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, order[1].ID); got != start {
		t.Errorf("an empty vote cost an opponent %d life", start-got)
	}
}

// The trigger goes on the stack with the vote's ballots on it, and a
// restore in between keeps them.
func TestFinishVotingTriggerSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	b43Catalog(g, order[0].ID, "Grudge Keeper", "Creature — Zombie Wizard", grudgeKeeperOracle, 2, 1)
	start := lifeOf(g, order[0].ID)
	castPleaAndVote(t, g, [][]string{{"time"}, {"knowledge"}, {"time"}, {"time"}})
	restored := restoreRoundTrip(t, g, true)
	if n := len(votesFinished(restored)); n != 1 {
		t.Fatalf("the restored game has %d finished-vote events, want 1", n)
	}
	passPriorityAroundTable(t, restored)
	order = voteOrder(restored)
	if got := lifeOf(restored, order[1].ID); got != start-2 {
		t.Errorf("order[1] (the one opponent who voted apart) has %d life, want %d", got, start-2)
	}
	for _, p := range []*game.Player{order[0], order[2], order[3]} {
		if got := lifeOf(restored, p.ID); got != start {
			t.Errorf("%s has %d life, want %d", p.Name, got, start)
		}
	}
}

// answerMayScry answers a "scry 2?" confirm, and the scry that follows a
// yes by keeping everything on top.
func answerMayScry(t *testing.T, g *game.Game, p *game.Player, yes bool) {
	t.Helper()
	c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, p.ID)
	if c == nil {
		t.Fatalf("%s has no scry confirm open: %+v", p.Name, g.PendingChoices)
	}
	if err := g.ResolveConfirm(c.ID, p.ID, yes); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if !yes {
		return
	}
	s := latestChoiceOfKindFor(g, game.PendingChoiceScry, p.ID)
	if s == nil || len(s.ScryCards) != 2 {
		t.Fatalf("%s has no scry 2 open after saying yes: %+v", p.Name, g.PendingChoices)
	}
	if err := g.ResolveScry(s.ID, p.ID, nil, s.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
}

// Model of Unity: you, and each opponent who voted for a choice you
// voted for, may scry 2 — in turn, and only those players are asked.
func TestModelOfUnity(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	b43Catalog(g, me.ID, "Model of Unity", "Artifact", modelOfUnityOracle, 0, 0)
	castPleaAndVote(t, g, [][]string{{"time"}, {"time"}, {"knowledge"}, {"time"}})
	passPriorityAroundTable(t, g)

	if latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID) == nil {
		t.Fatalf("the controller is asked first: %+v", g.PendingChoices)
	}
	for _, p := range order[1:] {
		if latestChoiceOfKindFor(g, game.PendingChoiceConfirm, p.ID) != nil {
			t.Fatalf("%s is asked before the controller has answered", p.Name)
		}
	}
	answerMayScry(t, g, me, true)

	// The next player is asked; a bot seat's moves include both answers.
	next := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, order[1].ID)
	if next == nil {
		t.Fatalf("the opponent who voted time is asked next: %+v", g.PendingChoices)
	}
	moves := legal.EnumerateFor(g, order[1].ID)
	if len(moves) < 2 {
		t.Errorf("the enumerator offers the opponent %d moves for the scry confirm, want both answers", len(moves))
	}
	answerMayScry(t, g, order[1], false)

	// order[2] voted knowledge: never asked. order[3] voted time.
	if latestChoiceOfKindFor(g, game.PendingChoiceConfirm, order[2].ID) != nil {
		t.Fatal("an opponent who voted for the other choice is asked to scry")
	}
	answerMayScry(t, g, order[3], true)
	if len(g.PendingChoices) != 0 {
		t.Errorf("choices left open after every answer: %+v", g.PendingChoices)
	}
}

// If the controller cast no ballot, only the controller scries.
func TestModelOfUnityControllerWhoDidNotVote(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	b43Catalog(g, me.ID, "Model of Unity", "Artifact", modelOfUnityOracle, 0, 0)
	emitFinishedVote(g, me.ID, ballot(order[1], pleaTime), ballot(order[2], pleaTime))
	passPriorityAroundTable(t, g)
	answerMayScry(t, g, me, true)
	if len(g.PendingChoices) != 0 {
		t.Errorf("opponents were asked though the controller voted for nothing: %+v", g.PendingChoices)
	}
}

// Erestor of the Council: a Treasure for each opponent who voted with
// you, scry X for the opponents who voted apart, then draw.
func TestErestorOfTheCouncil(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	b43Catalog(g, me.ID, "Erestor of the Council", "Legendary Creature — Elf Noble", erestorOfTheCouncilOID, 3, 4)
	castPleaAndVote(t, g, [][]string{{"time"}, {"time"}, {"knowledge"}, {"knowledge"}})
	hand := handSize(me)
	passPriorityAroundTable(t, g)

	for i, p := range order {
		want := 0
		if i == 1 {
			want = 1
		}
		if got := b43TokensNamed(g, p.ID, "Treasure"); got != want {
			t.Errorf("%s has %d Treasures, want %d", p.Name, got, want)
		}
	}
	s := latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID)
	if s == nil || len(s.ScryCards) != 2 {
		t.Fatalf("two opponents voted apart: want scry 2, got %+v", g.PendingChoices)
	}
	if got := handSize(me); got != hand {
		t.Errorf("the card was drawn (%d → %d) before the scry was answered", hand, got)
	}
	if err := g.ResolveScry(s.ID, me.ID, nil, s.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if got := handSize(me); got != hand+1 {
		t.Errorf("hand %d → %d after the scry, want +1", hand, got)
	}
}

// Everyone voted with Erestor's controller: Treasures for all, scry 0
// (no prompt), and the draw still happens.
func TestErestorWhenEveryOpponentVotedWithYou(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	b43Catalog(g, me.ID, "Erestor of the Council", "Legendary Creature — Elf Noble", erestorOfTheCouncilOID, 3, 4)
	castPleaAndVote(t, g, [][]string{{"time"}, {"time"}, {"time"}, {"time"}})
	hand := handSize(me)
	passPriorityAroundTable(t, g)
	for _, p := range order[1:] {
		if got := b43TokensNamed(g, p.ID, "Treasure"); got != 1 {
			t.Errorf("%s has %d Treasures, want 1", p.Name, got)
		}
	}
	if b43TokensNamed(g, me.ID, "Treasure") != 0 {
		t.Error("the controller made a Treasure")
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) != nil {
		t.Error("X is 0 yet a scry prompt opened")
	}
	if got := handSize(me); got != hand+1 {
		t.Errorf("hand %d → %d, want the draw", hand, got)
	}
}

// An opponent who voted for a choice the controller took AND one they
// didn't is in both groups: a Treasure, and counted in X. A seat that
// cast nothing is in neither.
func TestErestorOpponentInBothGroupsAndNonVoterInNeither(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	b43Catalog(g, me.ID, "Erestor of the Council", "Legendary Creature — Elf Noble", erestorOfTheCouncilOID, 3, 4)
	emitFinishedVote(g, me.ID,
		ballot(me, pleaTime),
		ballot(order[1], pleaTime), ballot(order[1], pleaKnowledge))
	passPriorityAroundTable(t, g)
	if got := b43TokensNamed(g, order[1].ID, "Treasure"); got != 1 {
		t.Errorf("the opponent who voted with me too has %d Treasures, want 1", got)
	}
	for _, p := range order[2:] {
		if got := b43TokensNamed(g, p.ID, "Treasure"); got != 0 {
			t.Errorf("%s cast no ballot yet has %d Treasures", p.Name, got)
		}
	}
	s := latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID)
	if s == nil || len(s.ScryCards) != 1 {
		t.Fatalf("only the split voter voted apart: want scry 1, got %+v", g.PendingChoices)
	}
}

// Undo clones the game: the clone's waiting trigger reads the same
// ballots, and resolving it there leaves the original untouched.
func TestFinishVotingTriggerSurvivesAClone(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	b43Catalog(g, order[0].ID, "Grudge Keeper", "Creature — Zombie Wizard", grudgeKeeperOracle, 2, 1)
	start := lifeOf(g, order[0].ID)
	castPleaAndVote(t, g, [][]string{{"time"}, {"knowledge"}, {"time"}, {"time"}})
	clone := g.Clone()
	passPriorityAroundTable(t, clone)
	if got := lifeOf(clone, order[1].ID); got != start-2 {
		t.Errorf("clone: the opponent who voted apart has %d life, want %d", got, start-2)
	}
	if got := lifeOf(g, order[1].ID); got != start {
		t.Errorf("resolving the clone's trigger changed the original: %d life, want %d", got, start)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, order[1].ID); got != start-2 {
		t.Errorf("original: the opponent who voted apart has %d life, want %d", got, start-2)
	}
}
