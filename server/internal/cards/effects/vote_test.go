package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// vote_test.go — ADR 0146, #2143: voting (CR 701.38) and the cards that
// ship with it.

const (
	pleaForPowerOracle     = "8b886623-5c96-48dd-a3d2-fd5865c58ff5"
	councilsJudgmentOracle = "acf4d685-a718-42f7-9a7a-dee069ae8db2"
	expropriateOracle      = "a69265b2-0e37-4d86-866e-e4a923233b4d"
	coercivePortalOracle   = "eea4feef-2595-4ed2-92a0-11ccb3e7d40e"
	magisterOfWorthOracle  = "0429d146-332f-4fb0-8273-c73679ca6d47"
	galadrielElvenOracle   = "b97a11c2-7119-4399-9565-bf69520d3988"
	bragosRepOracle        = "8877b896-8458-4828-9ff6-9e0ecc180cf7"
	ballotBrokerOracle     = "e8df915f-f509-4c6c-9261-a5a16c3f06c5"
	tivitOracle            = "f4e0b557-4f80-4b7c-bfcc-6d4c468faf2e"
)

// ballotFor is the open ballot owed by voter, or nil.
func ballotFor(g *game.Game, voter uuid.UUID) *game.PendingChoice {
	c := latestOptionPickFor(g, voter)
	if c == nil || c.CouncilVote == nil {
		return nil
	}
	return c
}

// voteFor casts voter's open ballot for the option labelled label.
func voteFor(t *testing.T, g *game.Game, voter uuid.UUID, label string) {
	t.Helper()
	c := ballotFor(g, voter)
	if c == nil {
		t.Fatalf("no ballot open for %s: %+v", voter, g.PendingChoices)
	}
	for i, o := range c.PickOptions {
		if o.Label == label {
			if err := g.ResolveOptionPick(c.ID, voter, i); err != nil {
				t.Fatalf("vote %q: %v", label, err)
			}
			return
		}
	}
	t.Fatalf("ballot for %s offers no %q: %+v", voter, label, c.PickOptions)
}

// voteOrder is the table in turn order from the active player.
func voteOrder(g *game.Game) []*game.Player {
	n := len(g.Seats)
	out := make([]*game.Player, n)
	for i := range out {
		out[i] = g.Seats[(g.Turn.ActiveSeat+i)%n]
	}
	return out
}

// toUpkeepOf walks the game to p's next upkeep and lets its triggers go
// on the stack and start resolving.
func toUpkeepOf(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	advanceToNextSeatsTurn(t, g)
	for i := 0; i < 80; i++ {
		if g.Seats[g.Turn.ActiveSeat].ID == p.ID && g.Turn.Step == game.StepUpkeep {
			passPriorityAroundTable(t, g)
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("never reached %s's upkeep", p.Name)
}

// castPlea casts Plea for Power for the active player and lets it
// start resolving.
func castPlea(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Plea for Power", "Sorcery", pleaForPowerOracle, nil)
	passPriorityAroundTable(t, g)
}

// Each player votes once, in turn order starting with the caster, and
// only the player voting now is asked (CR 701.38a). Knowledge wins 3-1:
// the caster draws three.
func TestPleaForPowerVotesInTurnOrderFromTheCaster(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	castPlea(t, g)
	hand := handSize(me)
	for i, p := range order {
		for j, q := range order {
			if (ballotFor(g, q.ID) != nil) != (i == j) {
				t.Fatalf("vote %d: seat %d has a ballot open = %v", i, j, ballotFor(g, q.ID) != nil)
			}
		}
		label := "knowledge"
		if i == 0 {
			label = "time"
		}
		voteFor(t, g, p.ID, label)
	}
	if got := handSize(me); got != hand+3 {
		t.Errorf("knowledge won: hand %d → %d, want +3", hand, got)
	}
	if len(g.ExtraTurns) != 0 {
		t.Errorf("time lost, yet %d extra turns are queued", len(g.ExtraTurns))
	}
}

// Every vote is logged as it is cast, before the next voter is asked.
func TestEachVoteIsPublicAsItIsCast(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	castPlea(t, g)
	voteFor(t, g, order[0].ID, "time")
	var cast []game.Event
	for _, ev := range g.Events {
		if ev.Kind == game.EventVoteCast {
			cast = append(cast, ev)
		}
	}
	if len(cast) != 1 || cast[0].Actor != order[0].ID || cast[0].Label != "time" {
		t.Fatalf("after one ballot the log has %+v, want the caster's vote for time", cast)
	}
	// The next ballot carries the tally so far for every seat to see.
	next := ballotFor(g, order[1].ID)
	if next == nil || len(next.CouncilVote.Ballots) != 1 || next.CouncilVote.Ballots[0].Option != 0 {
		t.Fatalf("the second ballot does not carry the first vote: %+v", next)
	}
}

// A tie goes to knowledge ("or the vote is tied"); more time votes take
// the extra turn.
func TestPleaForPowerTieDrawsAndTimeTakesATurn(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	castPlea(t, g)
	hand := handSize(me)
	for i, p := range order {
		label := "time"
		if i%2 == 1 {
			label = "knowledge"
		}
		voteFor(t, g, p.ID, label)
	}
	if handSize(me) != hand+3 || len(g.ExtraTurns) != 0 {
		t.Fatalf("2-2 is a tie and goes to knowledge: hand %d → %d, extra turns %d", hand, handSize(me), len(g.ExtraTurns))
	}

	g = newCatalogGame(t)
	order = voteOrder(g)
	castPlea(t, g)
	for i, p := range order {
		label := "time"
		if i == 3 {
			label = "knowledge"
		}
		voteFor(t, g, p.ID, label)
	}
	if len(g.ExtraTurns) != 1 {
		t.Errorf("time won 3-1: %d extra turns, want 1", len(g.ExtraTurns))
	}
}

// A voter who leaves with the ballot in front of them casts nothing,
// and the vote goes on to the next player (CR 800.4a).
func TestAVoterWhoLeavesIsSkipped(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	castPlea(t, g)
	hand := handSize(me)
	voteFor(t, g, me.ID, "time")
	if err := g.Concede(order[1].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if ballotFor(g, order[2].ID) == nil {
		t.Fatalf("the vote stopped at the seat that left: %+v", g.PendingChoices)
	}
	voteFor(t, g, order[2].ID, "time")
	voteFor(t, g, order[3].ID, "knowledge")
	if len(g.ExtraTurns) != 1 || handSize(me) != hand {
		t.Errorf("time won 2-1 among the players left: extra turns %d, hand %d → %d", len(g.ExtraTurns), hand, handSize(me))
	}
}

// A vote waiting on a ballot is a restore point, and the restored game
// finishes the vote with the ballots already cast.
func TestAnOpenVoteSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	castPlea(t, g)
	voteFor(t, g, order[0].ID, "time")
	voteFor(t, g, order[1].ID, "time")
	restored := restoreRoundTrip(t, g, true)
	voteFor(t, restored, order[2].ID, "knowledge")
	voteFor(t, restored, order[3].ID, "time")
	if len(restored.ExtraTurns) != 1 {
		t.Errorf("time won 3-1 across the restore: %d extra turns, want 1", len(restored.ExtraTurns))
	}
}

// Brago's Representative: its controller casts two ballots, both
// required. Ballot Broker: one more, which may be declined.
func TestExtraVotes(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	b43Catalog(g, me.ID, "Brago's Representative", "Creature — Human Advisor", bragosRepOracle, 1, 4)
	b43Catalog(g, order[1].ID, "Ballot Broker", "Creature — Human Advisor", ballotBrokerOracle, 2, 3)
	castPlea(t, g)
	voteFor(t, g, me.ID, "time")
	if c := ballotFor(g, me.ID); c == nil || c.CouncilVote.Optional() {
		t.Fatalf("Brago's Representative's vote is a second required ballot: %+v", c)
	}
	voteFor(t, g, me.ID, "time")
	voteFor(t, g, order[1].ID, "knowledge")
	c := ballotFor(g, order[1].ID)
	if c == nil || !c.CouncilVote.Optional() {
		t.Fatalf("Ballot Broker offers its controller a second ballot: %+v", c)
	}
	voteFor(t, g, order[1].ID, game.VoteDeclineLabel)
	voteFor(t, g, order[2].ID, "knowledge")
	voteFor(t, g, order[3].ID, "knowledge")
	if len(g.ExtraTurns) != 0 {
		t.Errorf("knowledge won 3-2 with Ballot Broker's vote declined: %d extra turns", len(g.ExtraTurns))
	}
	var votes int
	for _, ev := range g.Events {
		if ev.Kind == game.EventVoteCast {
			votes++
		}
	}
	if votes != 5 {
		t.Errorf("%d votes cast, want 5 (two from Brago's Representative's controller)", votes)
	}
}

// Council's Judgment: the options are the nonland permanents the caster
// doesn't control, nothing is targeted, and every permanent tied for
// the most votes is exiled.
func TestCouncilsJudgmentExilesEveryPermanentTiedForMostVotes(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	mine := ctrlPushCreature(g, me.ID, "My Bear")
	a := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hexproof Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: order[1].ID, Controller: order[1].ID,
		Keywords: []string{"hexproof"},
	})
	b := ctrlPushCreature(g, order[2].ID, "Other Bear")
	c := ctrlPushCreature(g, order[3].ID, "Third Bear")
	castCatalogSpell(t, g, "Council's Judgment", "Sorcery", councilsJudgmentOracle, nil)
	passPriorityAroundTable(t, g)
	for _, o := range ballotFor(g, me.ID).PickOptions {
		if len(o.Cards) == 1 && o.Cards[0] == mine {
			t.Fatal("the caster's own permanent is not on the ballot")
		}
	}
	voteFor(t, g, me.ID, "Hexproof Bear")
	voteFor(t, g, order[1].ID, "Other Bear")
	voteFor(t, g, order[2].ID, "Hexproof Bear")
	voteFor(t, g, order[3].ID, "Other Bear")
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("both permanents tied for the most votes are exiled, hexproof or not")
	}
	if !g.Battlefield.Contains(c) || !g.Battlefield.Contains(mine) {
		t.Error("a permanent with fewer votes stays")
	}
}

// Expropriate: an extra turn per time vote, a permanent the voter owns
// per money vote, and the spell exiles itself.
func TestExpropriateTakesTurnsAndPermanents(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	theirs := ctrlPushCreature(g, order[1].ID, "Their Bear")
	spell := castCatalogSpell(t, g, "Expropriate", "Sorcery", expropriateOracle, nil)
	passPriorityAroundTable(t, g)
	voteFor(t, g, me.ID, "time")
	voteFor(t, g, order[1].ID, "money")
	voteFor(t, g, order[2].ID, "time")
	voteFor(t, g, order[3].ID, "money")
	if len(g.ExtraTurns) != 2 {
		t.Errorf("two time votes: %d extra turns, want 2", len(g.ExtraTurns))
	}
	pick := latestChoiceOfKindFor(g, game.PendingChoiceTheirPermanents, me.ID)
	if pick == nil {
		t.Fatalf("no permanent to choose from the money voter: %+v", g.PendingChoices)
	}
	if err := g.ResolveTheirPermanents(pick.ID, me.ID, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c, ok := battlefieldCard(g, theirs); !ok || c.Controller != me.ID {
		t.Errorf("the caster gains control of the money voter's permanent")
	}
	if z := g.FindCardZoneForEffect(spell); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("Expropriate exiles itself, found in %v", z)
	}
}

// Galadriel, Elven-Queen: no vote unless another Elf entered under your
// control this turn; dominion tempts you and puts a +1/+1 counter on
// your Ring-bearer.
func TestGaladrielElvenQueenVotesOnlyAfterAnotherElfEntered(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	galadriel := b43Catalog(g, me.ID, "Galadriel, Elven-Queen", "Legendary Creature — Elf Noble", galadrielElvenOracle, 4, 5)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if ballotFor(g, me.ID) != nil {
		t.Fatal("no other Elf entered this turn, so there is no vote")
	}
	for range order {
		advanceToNextSeatsTurn(t, g)
	}
	castCatalogSpell(t, g, "Elvish Mystic", "Creature — Elf Druid", "", nil)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	for i, p := range order {
		label := "dominion"
		if i == 3 {
			label = "guidance"
		}
		voteFor(t, g, p.ID, label)
	}
	answerRing(t, g, me.ID, galadriel)
	passPriorityAroundTable(t, g)
	if ringBearerOf(g, me.ID) != galadriel || counterOn(g, galadriel, game.CounterPlusOne) != 1 {
		t.Errorf("dominion: bearer %v (want Galadriel), counters %d (want 1)",
			ringBearerOf(g, me.ID), counterOn(g, galadriel, game.CounterPlusOne))
	}
}

// Magister of Worth: condemnation (or a tie) destroys every other
// creature; grace returns every creature card from every graveyard.
func TestMagisterOfWorth(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	bear := ctrlPushCreature(g, order[1].ID, "Bear")
	castCatalogSpell(t, g, "Magister of Worth", "Creature — Angel", magisterOfWorthOracle, nil)
	passPriorityAroundTable(t, g)
	magister, ok := b43BattlefieldByName(g, "Magister of Worth")
	if !ok {
		t.Fatal("Magister of Worth did not enter")
	}
	for i, p := range order {
		label := "grace"
		if i%2 == 1 {
			label = "condemnation"
		}
		voteFor(t, g, p.ID, label)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !g.Battlefield.Contains(magister) {
		t.Fatalf("a tie is condemnation: bear kept %v, Magister kept %v",
			g.Battlefield.Contains(bear), g.Battlefield.Contains(magister))
	}

	castCatalogSpell(t, g, "Magister of Worth", "Creature — Angel", magisterOfWorthOracle, nil)
	passPriorityAroundTable(t, g)
	for _, p := range order {
		voteFor(t, g, p.ID, "grace")
	}
	passPriorityAroundTable(t, g)
	if _, ok := b43BattlefieldByName(g, "Bear"); !ok {
		t.Error("grace returns the Bear from its owner's graveyard")
	}
}

// Tivit: a Clue per evidence vote, a Treasure per bribery vote, and its
// controller may vote a second time.
func TestTivitSellerOfSecrets(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	castCatalogSpell(t, g, "Tivit, Seller of Secrets", "Legendary Creature — Sphinx Rogue", tivitOracle, nil)
	passPriorityAroundTable(t, g)
	voteFor(t, g, me.ID, "bribery")
	voteFor(t, g, me.ID, "bribery")
	for _, p := range order[1:] {
		voteFor(t, g, p.ID, "evidence")
	}
	passPriorityAroundTable(t, g)
	if n := b43TokensNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("%d Treasures, want 2", n)
	}
	if n := b43TokensNamed(g, me.ID, "Clue"); n != 3 {
		t.Errorf("%d Clues, want 3", n)
	}
}

// Coercive Portal: carnage sacrifices the Portal and destroys every
// nonland permanent; a tie is homage and draws.
func TestCoercivePortal(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	portal := b43Catalog(g, me.ID, "Coercive Portal", "Artifact", coercivePortalOracle, 0, 0)
	toUpkeepOf(t, g, me)
	hand := handSize(me)
	for i, p := range order {
		label := "carnage"
		if i%2 == 1 {
			label = "homage"
		}
		voteFor(t, g, p.ID, label)
	}
	if handSize(me) != hand+1 || !g.Battlefield.Contains(portal) {
		t.Fatalf("a tie is homage: hand %d → %d, portal kept %v", hand, handSize(me), g.Battlefield.Contains(portal))
	}

	bear := ctrlPushCreature(g, order[1].ID, "Bear")
	toUpkeepOf(t, g, me)
	for i, p := range order {
		label := "carnage"
		if i == 3 {
			label = "homage"
		}
		voteFor(t, g, p.ID, label)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(portal) || g.Battlefield.Contains(bear) {
		t.Error("carnage sacrifices the Portal and destroys every nonland permanent")
	}
}
