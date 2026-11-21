package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// toUpkeepOf walks the game to p's next upkeep and lets its triggers go
// on the stack and start resolving.
func toUpkeepOf(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	effectstest.AdvanceToNextSeatsTurn(t, g)
	for i := 0; i < 80; i++ {
		if g.Seats[g.Turn.ActiveSeat].ID == p.ID && g.Turn.Step == game.StepUpkeep {
			effectstest.PassPriorityAroundTable(t, g)
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("never reached %s's upkeep", p.Name)
}

// Each player votes once, in turn order starting with the caster, and
// only the player voting now is asked (CR 701.38a). Knowledge wins 3-1:
// the caster draws three.
func TestPleaForPowerVotesInTurnOrderFromTheCaster(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.CastPlea(t, g)
	hand := effectstest.HandSize(me)
	for i, p := range order {
		for j, q := range order {
			if (effectstest.BallotFor(g, q.ID) != nil) != (i == j) {
				t.Fatalf("vote %d: seat %d has a ballot open = %v", i, j, effectstest.BallotFor(g, q.ID) != nil)
			}
		}
		label := "knowledge"
		if i == 0 {
			label = "time"
		}
		effectstest.VoteFor(t, g, p.ID, label)
	}
	if got := effectstest.HandSize(me); got != hand+3 {
		t.Errorf("knowledge won: hand %d → %d, want +3", hand, got)
	}
	if len(g.ExtraTurns) != 0 {
		t.Errorf("time lost, yet %d extra turns are queued", len(g.ExtraTurns))
	}
}

// Every vote is logged as it is cast, before the next voter is asked.
func TestEachVoteIsPublicAsItIsCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	effectstest.CastPlea(t, g)
	effectstest.VoteFor(t, g, order[0].ID, "time")
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
	next := effectstest.BallotFor(g, order[1].ID)
	if next == nil || len(next.CouncilVote.Ballots) != 1 || next.CouncilVote.Ballots[0].Option != 0 {
		t.Fatalf("the second ballot does not carry the first vote: %+v", next)
	}
}

// A tie goes to knowledge ("or the vote is tied"); more time votes take
// the extra turn.
func TestPleaForPowerTieDrawsAndTimeTakesATurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.CastPlea(t, g)
	hand := effectstest.HandSize(me)
	for i, p := range order {
		label := "time"
		if i%2 == 1 {
			label = "knowledge"
		}
		effectstest.VoteFor(t, g, p.ID, label)
	}
	if effectstest.HandSize(me) != hand+3 || len(g.ExtraTurns) != 0 {
		t.Fatalf("2-2 is a tie and goes to knowledge: hand %d → %d, extra turns %d", hand, effectstest.HandSize(me), len(g.ExtraTurns))
	}

	g = effectstest.NewCatalogGame(t)
	order = effectstest.VoteOrder(g)
	effectstest.CastPlea(t, g)
	for i, p := range order {
		label := "time"
		if i == 3 {
			label = "knowledge"
		}
		effectstest.VoteFor(t, g, p.ID, label)
	}
	if len(g.ExtraTurns) != 1 {
		t.Errorf("time won 3-1: %d extra turns, want 1", len(g.ExtraTurns))
	}
}

// A voter who leaves with the ballot in front of them casts nothing,
// and the vote goes on to the next player (CR 800.4a).
func TestAVoterWhoLeavesIsSkipped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.CastPlea(t, g)
	hand := effectstest.HandSize(me)
	effectstest.VoteFor(t, g, me.ID, "time")
	if err := g.Concede(order[1].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if effectstest.BallotFor(g, order[2].ID) == nil {
		t.Fatalf("the vote stopped at the seat that left: %+v", g.PendingChoices)
	}
	effectstest.VoteFor(t, g, order[2].ID, "time")
	effectstest.VoteFor(t, g, order[3].ID, "knowledge")
	if len(g.ExtraTurns) != 1 || effectstest.HandSize(me) != hand {
		t.Errorf("time won 2-1 among the players left: extra turns %d, hand %d → %d", len(g.ExtraTurns), hand, effectstest.HandSize(me))
	}
}

// A vote waiting on a ballot is a restore point, and the restored game
// finishes the vote with the ballots already cast.
func TestAnOpenVoteSurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	effectstest.CastPlea(t, g)
	effectstest.VoteFor(t, g, order[0].ID, "time")
	effectstest.VoteFor(t, g, order[1].ID, "time")
	restored := effectstest.RestoreRoundTrip(t, g, true)
	effectstest.VoteFor(t, restored, order[2].ID, "knowledge")
	effectstest.VoteFor(t, restored, order[3].ID, "time")
	if len(restored.ExtraTurns) != 1 {
		t.Errorf("time won 3-1 across the restore: %d extra turns, want 1", len(restored.ExtraTurns))
	}
}

// Brago's Representative: its controller casts two ballots, both
// required. Ballot Broker: one more, which may be declined.
func TestExtraVotes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.B43Catalog(g, me.ID, "Brago's Representative", "Creature — Human Advisor", effectstest.BragosRepOracle, 1, 4)
	effectstest.B43Catalog(g, order[1].ID, "Ballot Broker", "Creature — Human Advisor", effectstest.BallotBrokerOracle, 2, 3)
	effectstest.CastPlea(t, g)
	effectstest.VoteFor(t, g, me.ID, "time")
	if c := effectstest.BallotFor(g, me.ID); c == nil || c.CouncilVote.Optional() {
		t.Fatalf("Brago's Representative's vote is a second required ballot: %+v", c)
	}
	effectstest.VoteFor(t, g, me.ID, "time")
	effectstest.VoteFor(t, g, order[1].ID, "knowledge")
	c := effectstest.BallotFor(g, order[1].ID)
	if c == nil || !c.CouncilVote.Optional() {
		t.Fatalf("Ballot Broker offers its controller a second ballot: %+v", c)
	}
	effectstest.VoteFor(t, g, order[1].ID, game.VoteDeclineLabel)
	effectstest.VoteFor(t, g, order[2].ID, "knowledge")
	effectstest.VoteFor(t, g, order[3].ID, "knowledge")
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
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	mine := effectstest.CtrlPushCreature(g, me.ID, "My Bear")
	a := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hexproof Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: order[1].ID, Controller: order[1].ID,
		Keywords: []string{"hexproof"},
	})
	b := effectstest.CtrlPushCreature(g, order[2].ID, "Other Bear")
	c := effectstest.CtrlPushCreature(g, order[3].ID, "Third Bear")
	effectstest.CastCatalogSpell(t, g, "Council's Judgment", "Sorcery", effectstest.CouncilsJudgmentOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, o := range effectstest.BallotFor(g, me.ID).PickOptions {
		if len(o.Cards) == 1 && o.Cards[0] == mine {
			t.Fatal("the caster's own permanent is not on the ballot")
		}
	}
	effectstest.VoteFor(t, g, me.ID, "Hexproof Bear")
	effectstest.VoteFor(t, g, order[1].ID, "Other Bear")
	effectstest.VoteFor(t, g, order[2].ID, "Hexproof Bear")
	effectstest.VoteFor(t, g, order[3].ID, "Other Bear")
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
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	theirs := effectstest.CtrlPushCreature(g, order[1].ID, "Their Bear")
	spell := effectstest.CastCatalogSpell(t, g, "Expropriate", "Sorcery", effectstest.ExpropriateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteFor(t, g, me.ID, "time")
	effectstest.VoteFor(t, g, order[1].ID, "money")
	effectstest.VoteFor(t, g, order[2].ID, "time")
	effectstest.VoteFor(t, g, order[3].ID, "money")
	if len(g.ExtraTurns) != 2 {
		t.Errorf("two time votes: %d extra turns, want 2", len(g.ExtraTurns))
	}
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceTheirPermanents, me.ID)
	if pick == nil {
		t.Fatalf("no permanent to choose from the money voter: %+v", g.PendingChoices)
	}
	if err := g.ResolveTheirPermanents(pick.ID, me.ID, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c, ok := effectstest.BattlefieldCard(g, theirs); !ok || c.Controller != me.ID {
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
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	galadriel := effectstest.B43Catalog(g, me.ID, "Galadriel, Elven-Queen", "Legendary Creature — Elf Noble", effectstest.GaladrielElvenOracle, 4, 5)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BallotFor(g, me.ID) != nil {
		t.Fatal("no other Elf entered this turn, so there is no vote")
	}
	for range order {
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	effectstest.CastCatalogSpell(t, g, "Elvish Mystic", "Creature — Elf Druid", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range order {
		label := "dominion"
		if i == 3 {
			label = "guidance"
		}
		effectstest.VoteFor(t, g, p.ID, label)
	}
	effectstest.AnswerRing(t, g, me.ID, galadriel)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RingBearerOf(g, me.ID) != galadriel || effectstest.CounterOn(g, galadriel, game.CounterPlusOne) != 1 {
		t.Errorf("dominion: bearer %v (want Galadriel), counters %d (want 1)",
			effectstest.RingBearerOf(g, me.ID), effectstest.CounterOn(g, galadriel, game.CounterPlusOne))
	}
}

// Magister of Worth: condemnation (or a tie) destroys every other
// creature; grace returns every creature card from every graveyard.
func TestMagisterOfWorth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	bear := effectstest.CtrlPushCreature(g, order[1].ID, "Bear")
	effectstest.CastCatalogSpell(t, g, "Magister of Worth", "Creature — Angel", effectstest.MagisterOfWorthOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	magister, ok := effectstest.B43BattlefieldByName(g, "Magister of Worth")
	if !ok {
		t.Fatal("Magister of Worth did not enter")
	}
	for i, p := range order {
		label := "grace"
		if i%2 == 1 {
			label = "condemnation"
		}
		effectstest.VoteFor(t, g, p.ID, label)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !g.Battlefield.Contains(magister) {
		t.Fatalf("a tie is condemnation: bear kept %v, Magister kept %v",
			g.Battlefield.Contains(bear), g.Battlefield.Contains(magister))
	}

	effectstest.CastCatalogSpell(t, g, "Magister of Worth", "Creature — Angel", effectstest.MagisterOfWorthOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range order {
		effectstest.VoteFor(t, g, p.ID, "grace")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.B43BattlefieldByName(g, "Bear"); !ok {
		t.Error("grace returns the Bear from its owner's graveyard")
	}
}

// Tivit: a Clue per evidence vote, a Treasure per bribery vote, and its
// controller may vote a second time.
func TestTivitSellerOfSecrets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.CastCatalogSpell(t, g, "Tivit, Seller of Secrets", "Legendary Creature — Sphinx Rogue", effectstest.TivitOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteFor(t, g, me.ID, "bribery")
	effectstest.VoteFor(t, g, me.ID, "bribery")
	for _, p := range order[1:] {
		effectstest.VoteFor(t, g, p.ID, "evidence")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B43TokensNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("%d Treasures, want 2", n)
	}
	if n := effectstest.B43TokensNamed(g, me.ID, "Clue"); n != 3 {
		t.Errorf("%d Clues, want 3", n)
	}
}

// Coercive Portal: carnage sacrifices the Portal and destroys every
// nonland permanent; a tie is homage and draws.
func TestCoercivePortal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	portal := effectstest.B43Catalog(g, me.ID, "Coercive Portal", "Artifact", effectstest.CoercivePortalOracle, 0, 0)
	toUpkeepOf(t, g, me)
	hand := effectstest.HandSize(me)
	for i, p := range order {
		label := "carnage"
		if i%2 == 1 {
			label = "homage"
		}
		effectstest.VoteFor(t, g, p.ID, label)
	}
	if effectstest.HandSize(me) != hand+1 || !g.Battlefield.Contains(portal) {
		t.Fatalf("a tie is homage: hand %d → %d, portal kept %v", hand, effectstest.HandSize(me), g.Battlefield.Contains(portal))
	}

	bear := effectstest.CtrlPushCreature(g, order[1].ID, "Bear")
	toUpkeepOf(t, g, me)
	for i, p := range order {
		label := "carnage"
		if i == 3 {
			label = "homage"
		}
		effectstest.VoteFor(t, g, p.ID, label)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(portal) || g.Battlefield.Contains(bear) {
		t.Error("carnage sacrifices the Portal and destroys every nonland permanent")
	}
}

// #3669: a ballot has no abstain, so the enumerator's first option is
// the answer a stalled bot or the MCP seat falls back on (AlwaysLegal,
// aiseat.SafeIndex). That flag has to stay on the wire even though the
// text seat no longer calls the option "always legal": every ballot is
// equally legal, and a mandatory vote must never leave a seat with no
// way out.
func TestCouncilsJudgmentBallotKeepsItsEscapeAnswer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.CtrlPushCreature(g, order[1].ID, "Alpha Bear")
	effectstest.CtrlPushCreature(g, order[2].ID, "Beta Bear")
	effectstest.CastCatalogSpell(t, g, "Council's Judgment", "Sorcery", effectstest.CouncilsJudgmentOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BallotFor(g, me.ID) == nil {
		t.Fatalf("no ballot open for the caster: %+v", g.PendingChoices)
	}

	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) < 2 {
		t.Fatalf("a ballot over two permanents offered %d moves", len(moves))
	}
	safe := aiseat.SafeIndex(moves)
	if safe < 0 {
		t.Fatal("a mandatory vote has no always-legal answer: a bot that cannot place a ballot would hold the table")
	}
	if !moves[0].AlwaysLegal || safe != 0 {
		t.Errorf("the escape is move %d (always_legal on 0: %v), want the first option", safe, moves[0].AlwaysLegal)
	}
	for i, m := range moves[1:] {
		if m.AlwaysLegal {
			t.Errorf("ballot option %d is also always_legal; only the first is the escape", i+1)
		}
	}
}
