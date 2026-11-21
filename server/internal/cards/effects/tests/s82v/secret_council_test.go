package testss82v

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func TestSecretCouncilCardsAreRegisteredFull(t *testing.T) {
	for _, oracle := range []string{effectstest.TruthOrConsequencesOracle, effectstest.MobVerdictOracle, effectstest.CirdanOracle, effectstest.TrapTheTrespassersOracle, effectstest.Vault11Oracle, effectstest.ElrondOracle} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// secretVoteSeen is the council vote on `viewer`'s wire, and that
// viewer's log.
func secretVoteSeen(t *testing.T, g *game.Game, viewer string) (*protocol.CouncilVoteView, []protocol.LogEvent) {
	t.Helper()
	v := protocol.ViewOfGameFor(g, viewer)
	var cv *protocol.CouncilVoteView
	for _, c := range v.PendingChoices {
		if c.CouncilVote != nil {
			cv = c.CouncilVote
		}
	}
	return cv, v.Log
}

func logKinds(log []protocol.LogEvent, kind protocol.LogKind) []protocol.LogEvent {
	var out []protocol.LogEvent
	for _, e := range log {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}

// Two of four ballots are in. Each voter's own wire carries their own
// ballot and nobody else's; the other seats, a spectator and the admin
// see who has voted and nothing more; the log says "cast a secret vote"
// with no option; and the engine's events carry no option either.
func TestASecretBallotReachesNoOtherSeatSpectatorAdminOrLog(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	effectstest.CastTruthOrConsequences(t, g)
	effectstest.VoteFor(t, g, order[0].ID, "consequences")
	effectstest.VoteFor(t, g, order[1].ID, "truth")

	for _, ev := range g.Events {
		if ev.Kind == game.EventVoteCast {
			t.Fatalf("a secret ballot was logged as an open vote: %+v", ev)
		}
		if ev.Kind == game.EventSecretVoteCast && (ev.Label != "" || ev.Amount != 0 || ev.Target != uuid.Nil) {
			t.Fatalf("a secret ballot's event says what it was: %+v", ev)
		}
	}

	viewers := map[string][]int{
		order[0].ID.String():       {1}, // consequences
		order[1].ID.String():       {0}, // truth
		order[2].ID.String():       nil,
		order[3].ID.String():       nil,
		protocol.SpectatorViewerID: nil,
		"":                         nil, // the admin's omniscient view
	}
	for viewer, own := range viewers {
		cv, log := secretVoteSeen(t, g, viewer)
		if cv == nil || !cv.Secret {
			t.Fatalf("viewer %q: no secret vote on the wire: %+v", viewer, cv)
		}
		for i, n := range cv.Tally {
			if n != 0 {
				t.Errorf("viewer %q: tally[%d] = %d, want a zero tally", viewer, i, n)
			}
		}
		var got []int
		for _, b := range cv.Ballots {
			if b.Voter != viewer {
				t.Errorf("viewer %q sees %s's ballot", viewer, b.Voter)
			}
			got = append(got, b.Option)
		}
		if len(got) != len(own) || (len(own) == 1 && got[0] != own[0]) {
			t.Errorf("viewer %q sees ballots %v, want their own %v", viewer, got, own)
		}
		if len(cv.Voted) != 2 {
			t.Errorf("viewer %q: voted = %v, want the two who have voted", viewer, cv.Voted)
		}
		if n := len(logKinds(log, protocol.LogVote)); n != 0 {
			t.Errorf("viewer %q: %d open vote lines in the log", viewer, n)
		}
		if n := len(logKinds(log, protocol.LogVoteSecret)); n != 2 {
			t.Errorf("viewer %q: %d secret vote lines, want 2", viewer, n)
		}
		raw, _ := json.Marshal(log)
		if s := string(raw); strings.Contains(s, "voted for") || strings.Contains(s, `"choice"`) {
			t.Errorf("viewer %q: the log names a choice before the reveal: %s", viewer, s)
		}
	}

	// The unfiltered view never serialises a secret ballot either.
	for _, c := range protocol.ViewOfGame(g).PendingChoices {
		if c.CouncilVote == nil {
			continue
		}
		raw, _ := json.Marshal(c.CouncilVote)
		if strings.Contains(string(raw), `"ballots"`) {
			t.Errorf("the shared view carries ballots: %s", raw)
		}
	}
}

// The votes are revealed together, once, as one log line naming every
// vote, after the last ballot and before the card acts on the result.
func TestASecretVoteIsRevealedTogetherOnOneLine(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.CastTruthOrConsequences(t, g)
	hand := effectstest.HandSize(me)
	life := map[uuid.UUID]int{}
	for _, p := range order[1:] {
		life[p.ID] = p.Life
	}
	labels := []string{"consequences", "truth", "consequences", "truth"}
	for i, p := range order {
		effectstest.VoteFor(t, g, p.ID, labels[i])
		_, log := secretVoteSeen(t, g, protocol.SpectatorViewerID)
		revealed := logKinds(log, protocol.LogVotesRevealed)
		if i < len(order)-1 && len(revealed) != 0 {
			t.Fatalf("revealed after %d of %d ballots", i+1, len(order))
		}
	}
	_, log := secretVoteSeen(t, g, protocol.SpectatorViewerID)
	revealed := logKinds(log, protocol.LogVotesRevealed)
	if len(revealed) != 1 {
		t.Fatalf("%d reveal lines, want exactly 1", len(revealed))
	}
	r := revealed[0]
	if strings.Join(r.Votes, ",") != strings.Join(labels, ",") || len(r.Seats) != 4 {
		t.Errorf("reveal = seats %v votes %v, want %v", r.Seats, r.Votes, labels)
	}
	if !strings.HasPrefix(r.Text, "Votes revealed (Truth or Consequences): ") || !strings.Contains(r.Text, me.Name+" for consequences") {
		t.Errorf("reveal text = %q", r.Text)
	}
	revealAt, finishAt := -1, -1
	for i, ev := range g.Events {
		switch ev.Kind {
		case game.EventVotesRevealed:
			revealAt = i
		case game.EventVotesFinished:
			finishAt = i
		}
	}
	if revealAt < 0 || finishAt < revealAt {
		t.Errorf("reveal at %d, finish at %d: the reveal comes first", revealAt, finishAt)
	}
	// Two truth votes draw two; two consequences votes burn one random
	// opponent for 6.
	if got := effectstest.HandSize(me); got != hand+2 {
		t.Errorf("hand %d -> %d, want +2", hand, got)
	}
	burnt := 0
	for _, p := range order[1:] {
		switch life[p.ID] - p.Life {
		case 0:
		case 6:
			burnt++
		default:
			t.Errorf("%s lost %d life", p.Name, life[p.ID]-p.Life)
		}
	}
	if burnt != 1 {
		t.Errorf("%d opponents took the 6, want one at random", burnt)
	}
}

// A table mid-way through a secret vote is a restore point. The
// restored game still hides the ballots cast before it, and reveals all
// of them when the vote ends.
func TestASecretVoteSurvivesARestoreWithItsBallotsHidden(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.CastTruthOrConsequences(t, g)
	effectstest.VoteFor(t, g, order[0].ID, "truth")
	effectstest.VoteFor(t, g, order[1].ID, "truth")
	restored := effectstest.RestoreRoundTrip(t, g, true)

	if cv, _ := secretVoteSeen(t, restored, order[2].ID.String()); cv == nil || !cv.Secret || len(cv.Ballots) != 0 {
		t.Fatalf("after the restore seat 2 sees %+v, want a secret vote with none of its ballots", cv)
	}
	if cv, _ := secretVoteSeen(t, restored, me.ID.String()); cv == nil || len(cv.Ballots) != 1 || cv.Ballots[0].Option != 0 {
		t.Fatalf("after the restore the caster sees %+v, want their own truth ballot", cv)
	}
	hand := effectstest.HandSize(playerIn(restored, me.ID))
	effectstest.VoteFor(t, restored, order[2].ID, "truth")
	effectstest.VoteFor(t, restored, order[3].ID, "consequences")
	_, log := secretVoteSeen(t, restored, protocol.SpectatorViewerID)
	if r := logKinds(log, protocol.LogVotesRevealed); len(r) != 1 || strings.Join(r[0].Votes, ",") != "truth,truth,truth,consequences" {
		t.Fatalf("reveal after the restore = %+v", r)
	}
	if got := effectstest.HandSize(playerIn(restored, me.ID)); got != hand+3 {
		t.Errorf("three truth votes across the restore: hand %d -> %d, want +3", hand, got)
	}
}

func playerIn(g *game.Game, id uuid.UUID) *game.Player {
	for _, p := range g.Seats {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func ballotOffers(c *game.PendingChoice, label string) bool {
	for _, o := range c.PickOptions {
		if o.Label == label {
			return true
		}
	}
	return false
}

// Mob Verdict: nobody may vote for themselves. Each vote an opponent
// received is 2 damage to them and their creatures, vote by vote; each
// vote the caster received draws a card.
func TestMobVerdictDamagesPerVoteAndDrawsForTheCaster(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me, a, b, c := order[0], order[1], order[2], order[3]
	bear := effectstest.PushVanillaCreature(g, a.ID, "A's Bear", 3, 3)
	ox := effectstest.PushVanillaCreature(g, b.ID, "B's Ox", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Mob Verdict", "Sorcery", effectstest.MobVerdictOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hand, lifeA, lifeB := effectstest.HandSize(me), a.Life, b.Life

	for _, p := range order {
		ballot := effectstest.BallotFor(g, p.ID)
		if ballot == nil {
			t.Fatalf("no ballot for %s", p.Name)
		}
		if ballotOffers(ballot, p.Name) {
			t.Errorf("%s may vote for themselves: %+v", p.Name, ballot.PickOptions)
		}
		target := map[uuid.UUID]string{me.ID: a.Name, a.ID: me.Name, b.ID: a.Name, c.ID: b.Name}[p.ID]
		effectstest.VoteFor(t, g, p.ID, target)
	}
	// A got two votes: 4 to A, and A's 3/3 dies to the pair.
	if a.Life != lifeA-4 || g.Battlefield.Contains(bear) {
		t.Errorf("A: life %d -> %d, bear alive %v; want -4 and dead", lifeA, a.Life, g.Battlefield.Contains(bear))
	}
	// B got one: 2 to B and B's 3/3 survives.
	if b.Life != lifeB-2 || !g.Battlefield.Contains(ox) {
		t.Errorf("B: life %d -> %d, ox alive %v; want -2 and alive", lifeB, b.Life, g.Battlefield.Contains(ox))
	}
	if got := effectstest.HandSize(me); got != hand+1 {
		t.Errorf("the caster received one vote: hand %d -> %d, want +1", hand, got)
	}
}

// Círdan the Shipwright: each player draws per vote received; those
// with none choose, in turn, a permanent card from hand, and the
// chosen cards enter together.
func TestCirdanDrawsPerVoteThenTheUnvotedPutPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me, a, b, c := order[0], order[1], order[2], order[3]
	bGolem := effectstest.HandCardForTest(b, "B's Golem", "Artifact Creature — Golem", "")
	cRock := effectstest.HandCardForTest(c, "C's Rock", "Artifact", "")
	effectstest.CastCreatureSpellForCopy(t, g, "Círdan the Shipwright", "Legendary Creature — Elf Noble", effectstest.CirdanOracle, 3, 4)
	effectstest.PassPriorityAroundTable(t, g)
	handMe, handA := effectstest.HandSize(me), effectstest.HandSize(a)
	if ballot := effectstest.BallotFor(g, me.ID); ballot == nil || !ballotOffers(ballot, me.Name) {
		t.Fatalf("a player may vote for themselves: %+v", ballot)
	}
	effectstest.VoteFor(t, g, me.ID, me.Name)
	effectstest.VoteFor(t, g, a.ID, me.Name)
	effectstest.VoteFor(t, g, b.ID, a.Name)
	effectstest.VoteFor(t, g, c.ID, me.Name)
	if effectstest.HandSize(me) != handMe+3 || effectstest.HandSize(a) != handA+1 {
		t.Errorf("draws: me %d -> %d (want +3), A %d -> %d (want +1)", handMe, effectstest.HandSize(me), handA, effectstest.HandSize(a))
	}
	effectstest.AnswerChooseCards(t, g, b.ID, bGolem)
	if g.Battlefield.Contains(bGolem) {
		t.Fatal("B's pick entered before C had chosen: the puts are simultaneous")
	}
	effectstest.AnswerChooseCards(t, g, c.ID, cRock)
	if !g.Battlefield.Contains(bGolem) || !g.Battlefield.Contains(cRock) {
		t.Errorf("both chosen permanents should have entered")
	}
	if card, ok := effectstest.BattlefieldCard(g, bGolem); !ok || card.Controller != b.ID {
		t.Errorf("B's Golem enters under its owner's control: %+v", card)
	}
}

// Trap the Trespassers: per creature with votes, that many stun
// counters and a tap; the caster's own creatures are not options.
func TestTrapTheTrespassersStunsByVotes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	x := effectstest.PushVanillaCreature(g, order[1].ID, "X", 2, 2)
	y := effectstest.PushVanillaCreature(g, order[2].ID, "Y", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Trap the Trespassers", "Instant", effectstest.TrapTheTrespassersOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if ballotOffers(effectstest.BallotFor(g, me.ID), "My Bear") {
		t.Fatal("the caster's own creature is on the ballot")
	}
	effectstest.VoteRound(t, g, "X", "X", "Y", "X")
	for id, want := range map[uuid.UUID]int{x: 3, y: 1, mine: 0} {
		card, _ := effectstest.BattlefieldCard(g, id)
		if got := effectstest.CounterCount(g, id, game.CounterStun); got != want || card.Tapped != (want > 0) {
			t.Errorf("%s: %d stun counters, tapped %v; want %d", card.Name, got, card.Tapped, want)
		}
	}
}

// vault11ToChapterTwo casts the Saga (chapter I) and walks to chapter
// II's vote.
func vault11ToChapterTwo(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	seat := g.Turn.ActiveSeat
	saga := effectstest.CastCatalogSpell(t, g, "Vault 11: Voter's Dilemma", "Enchantment — Saga", effectstest.Vault11Oracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	return saga
}

// Vault 11: chapter I makes a Soldier per opponent; chapter II's
// secret vote may be for no creature, and a vote nobody cast for a
// creature draws everyone a card.
func TestVault11NoCreatureVotesDrawsEachPlayerACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	vault11ToChapterTwo(t, g)
	if got := effectstest.CountBattlefieldByName(g, "Human Soldier"); got != 3 {
		t.Errorf("chapter I: %d Human Soldiers, want one per opponent (3)", got)
	}
	hands := map[uuid.UUID]int{}
	for _, p := range order {
		hands[p.ID] = effectstest.HandSize(p)
	}
	effectstest.VoteRound(t, g, "No creature", "No creature", "No creature", "No creature")
	for _, p := range order {
		if got := effectstest.HandSize(p); got != hands[p.ID]+1 {
			t.Errorf("%s: hand %d -> %d, want +1", p.Name, hands[p.ID], got)
		}
	}
	_, log := secretVoteSeen(t, g, protocol.SpectatorViewerID)
	if r := logKinds(log, protocol.LogVotesRevealed); len(r) != 1 || strings.Join(r[0].Votes, ",") != "nothing,nothing,nothing,nothing" {
		t.Errorf("reveal = %+v, want four votes for nothing", r)
	}
}

// ... and otherwise every creature tied for the most votes is destroyed.
func TestVault11DestroysEachCreatureTiedForMostVotes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	x := effectstest.PushVanillaCreature(g, order[1].ID, "X", 2, 2)
	y := effectstest.PushVanillaCreature(g, order[2].ID, "Y", 2, 2)
	z := effectstest.PushVanillaCreature(g, order[3].ID, "Z", 2, 2)
	vault11ToChapterTwo(t, g)
	effectstest.VoteRound(t, g, "X", "Y", "No creature", "Z")
	if g.Battlefield.Contains(x) || g.Battlefield.Contains(y) || g.Battlefield.Contains(z) {
		t.Errorf("X, Y and Z tied at one vote each: all are destroyed")
	}
}

// Elrond of the White Council: each fellowship voter chooses a creature
// they control; Elrond's controller takes them all, they can't attack
// their owners, and each aid vote is a +1/+1 counter on every creature
// Elrond's controller controls, the new ones included.
func TestElrondTakesFellowshipCreaturesAndCountsAid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me, a, b := order[0], order[1], order[2]
	keepA := effectstest.PushVanillaCreature(g, a.ID, "A's Keeper", 2, 2)
	giftA := effectstest.PushVanillaCreature(g, a.ID, "A's Gift", 1, 1)
	giftB := effectstest.PushVanillaCreature(g, b.ID, "B's Gift", 1, 1)
	effectstest.CastCreatureSpellForCopy(t, g, "Elrond of the White Council", "Legendary Creature — Elf Noble", effectstest.ElrondOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	elrond, ok := effectstest.B43BattlefieldByName(g, "Elrond of the White Council")
	if !ok {
		t.Fatal("Elrond did not enter")
	}
	effectstest.VoteRound(t, g, "aid", "fellowship", "fellowship", "aid")
	effectstest.AnswerKeep(t, g, a.ID, giftA)
	effectstest.AnswerKeep(t, g, b.ID, giftB)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })

	for _, id := range []uuid.UUID{giftA, giftB} {
		card, ok := effectstest.BattlefieldCard(g, id)
		if !ok || card.Controller != me.ID {
			t.Fatalf("%s: Elrond's controller gains control of each chosen creature: %+v", card.Name, card)
		}
		found := false
		for _, r := range card.Effective().AttackTargetRestrictions {
			if r.NotOwner {
				found = true
			}
		}
		if !found {
			t.Errorf("%s did not gain \"can't attack its owner\"", card.Name)
		}
	}
	if card, _ := effectstest.BattlefieldCard(g, keepA); card.Controller != a.ID {
		t.Errorf("the creature A did not choose stayed with A")
	}
	for _, id := range []uuid.UUID{elrond, giftA, giftB} {
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 2 {
			t.Errorf("two aid votes: %d +1/+1 counters on %s, want 2", got, id)
		}
	}
	if got := effectstest.CounterCount(g, keepA, game.CounterPlusOne); got != 0 {
		t.Errorf("A's creature got %d counters", got)
	}
}
