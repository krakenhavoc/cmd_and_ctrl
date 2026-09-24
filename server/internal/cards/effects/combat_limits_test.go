package effects

import (
	"encoding/json"
	"errors"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// combat_limits_test.go — #1507's proof cards: the CR 508.1c / 509.1b
// whole-combat count limits (ADR 0045 amendment of 2026-09-24,
// Decisions 43-45). Silent Arbiter, Dueling Grounds and Caverns of
// Despair carry both halves; Crawlspace and Judoon Enforcers the
// per-defender attack half.
//
// Every shape is pinned the way block_rules_test.go pins #750's: the
// VERB accepts the legal declaration and refuses the illegal one with
// its reason, and the ENUMERATOR offers exactly what the verb accepts —
// checked exhaustively, by trying every candidate on a clone.

const (
	silentArbiterOracle    = "1cdf30de-d88c-421a-80df-4917bbd2f09e"
	crawlspaceOracle       = "2296370c-fe34-4df6-92a5-260f1634bede"
	duelingGroundsOracle   = "eb2df4a1-6b63-4f6a-a830-e9487afe59f8"
	cavernsOfDespairOracle = "a1034a02-36cf-4586-a001-9dc3fb76e904"
	judoonEnforcersOracle  = "ca04089c-24b6-465e-9303-ea28c0d6f3c7"
)

// clAttack is one (attacker, target) pair.
type clAttack struct{ attacker, target uuid.UUID }

func clSorted(in []clAttack) []clAttack {
	sort.Slice(in, func(i, j int) bool {
		if in[i].attacker != in[j].attacker {
			return in[i].attacker.String() < in[j].attacker.String()
		}
		return in[i].target.String() < in[j].target.String()
	})
	return in
}

// clOfferedAttacks is every attack the enumerator offers `seat`.
func clOfferedAttacks(t *testing.T, g *game.Game, seat uuid.UUID) []clAttack {
	t.Helper()
	var out []clAttack
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Kind != legal.KindAttack {
			continue
		}
		var p struct {
			Attacker string `json:"attacker"`
			Target   string `json:"target"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("attack params %s: %v", m.Params, err)
		}
		out = append(out, clAttack{uuid.MustParse(p.Attacker), uuid.MustParse(p.Target)})
	}
	return clSorted(out)
}

// clAcceptedAttacks is every (creature of seat, opponent) attack the
// VERB accepts right now, each tried on its own clone.
func clAcceptedAttacks(g *game.Game, seat uuid.UUID) []clAttack {
	var out []clAttack
	for _, c := range g.Battlefield.Cards {
		if c.Controller != seat || !c.IsCreature() || c.AttackingTarget != uuid.Nil {
			continue
		}
		for _, p := range g.Seats {
			if p.ID == seat {
				continue
			}
			if err := g.Clone().DeclareAttacker(c.InstanceID, p.ID); err == nil {
				out = append(out, clAttack{c.InstanceID, p.ID})
			}
		}
	}
	return clSorted(out)
}

// clAgreeOnAttacks asserts the enumerator offers exactly what the verb
// accepts, and returns the offer.
func clAgreeOnAttacks(t *testing.T, g *game.Game, seat uuid.UUID) []clAttack {
	t.Helper()
	offered, accepted := clOfferedAttacks(t, g, seat), clAcceptedAttacks(g, seat)
	if len(offered) != len(accepted) {
		t.Fatalf("the enumerator offers %d attacks, the verb accepts %d:\noffered  %v\naccepted %v", len(offered), len(accepted), offered, accepted)
	}
	for i := range offered {
		if offered[i] != accepted[i] {
			t.Fatalf("the enumerator and the verb disagree:\noffered  %v\naccepted %v", offered, accepted)
		}
	}
	return offered
}

// clAgreeOnBlocks asserts every single block the verb accepts is
// offered and every offered block is accepted, and returns how many
// block moves were offered.
func clAgreeOnBlocks(t *testing.T, g *game.Game, seat uuid.UUID) int {
	t.Helper()
	offered := brOffered(t, g, seat)
	for _, set := range offered {
		if err := g.Clone().DeclareBlockers(set); err != nil {
			t.Fatalf("the enumerator offers %v, which the verb refuses: %v", set, err)
		}
	}
	for _, b := range g.Battlefield.Cards {
		if b.Controller != seat || !b.IsCreature() || b.Tapped || b.BlockingTarget != uuid.Nil {
			continue
		}
		for _, a := range g.Battlefield.Cards {
			if a.AttackingTarget == uuid.Nil {
				continue
			}
			if g.Clone().DeclareBlocker(b.InstanceID, a.InstanceID) == nil && !brOffers(offered, b.InstanceID, a.InstanceID) {
				t.Fatalf("the verb accepts %s blocking %s but the enumerator withholds it", b.Name, a.Name)
			}
		}
	}
	return len(offered)
}

func clAttackRefusal(t *testing.T, err error) *game.AttackLimitError {
	t.Helper()
	var le *game.AttackLimitError
	if err == nil || !errors.Is(err, game.ErrAttackLimit) || !errors.As(err, &le) {
		t.Fatalf("want an attack-limit refusal, got %v", err)
	}
	return le
}

func clTargets(offered []clAttack) map[uuid.UUID]int {
	out := map[uuid.UUID]int{}
	for _, a := range offered {
		out[a.target]++
	}
	return out
}

// --- Silent Arbiter ---------------------------------------------

// Attack side: every creature is offered until one attacks, then none
// is, and the verb refuses the second with the Arbiter named.
func TestSilentArbiterAllowsOneAttackerEachCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arbiter := b12Push(g, opp.ID, "Silent Arbiter", "Artifact Creature — Construct", silentArbiterOracle, 1, 5)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)

	if n := len(clAgreeOnAttacks(t, g, me.ID)); n != 6 {
		t.Fatalf("before any attack, %d attacks are offered; want two creatures at three opponents", n)
	}
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("the first attacker: %v", err)
	}
	if n := len(clAgreeOnAttacks(t, g, me.ID)); n != 0 {
		t.Errorf("after one attacker, %d more attacks are offered", n)
	}
	le := clAttackRefusal(t, g.DeclareAttacker(b, g.Seats[2].ID))
	if le.Source != arbiter || le.Max != 1 {
		t.Errorf("refusal = %+v", le)
	}
}

// Block side: one blocker in the whole combat, across attackers — and
// the #328 signal lets the defender go once it is used.
func TestSilentArbiterAllowsOneBlockerEachCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Silent Arbiter", "Artifact Creature — Construct", silentArbiterOracle, 1, 5)
	atk := b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	blkA := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	blkB := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	brAttack(t, g, atk)

	if n := clAgreeOnBlocks(t, g, opp.ID); n != 2 {
		t.Fatalf("before any block, %d blocks are offered, want 2", n)
	}
	refused := brRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: blkA, Attacker: atk}, {Blocker: blkB, Attacker: atk},
	}), game.BlockReasonDeclarationLimit)
	if refused.SourceName != "Silent Arbiter" || refused.N != 1 {
		t.Errorf("refusal = %+v", refused)
	}
	if err := g.DeclareBlocker(blkA, atk); err != nil {
		t.Fatalf("one block is legal: %v", err)
	}
	if n := clAgreeOnBlocks(t, g, opp.ID); n != 0 {
		t.Errorf("after one block, %d more are offered", n)
	}
	if g.SeatOwesBlockDecision(opp.ID) || brViewOwes(g) {
		t.Error("block_decision_seats holds the defender for a block it cannot make")
	}
	brRefusal(t, g.DeclareBlocker(blkB, atk), game.BlockReasonDeclarationLimit)
}

// CR 613.1f: an Arbiter that has lost its abilities limits nothing.
func TestSilentArbiterWithNoAbilitiesLimitsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arbiter := b12Push(g, opp.ID, "Silent Arbiter", "Artifact Creature — Construct", silentArbiterOracle, 1, 5)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		g.RegisterScopedStaticForEffect(game.StaticAbility{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.InstanceID == arbiter
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.AbilitiesRemoved = true
			},
		}, uuid.Nil, "loses all abilities", g.UntilEndOfTurnDuration())
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("an Arbiter with no abilities refused an attacker: %v", err)
		}
	}
}

// --- Dueling Grounds ---------------------------------------------

func TestDuelingGroundsAllowsOneAttackerAndOneBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Dueling Grounds", "Enchantment", duelingGroundsOracle, 0, 0)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	blkA := b12Creature(g, opp.ID, "Wall A", "Creature — Wall", 0, 4)
	blkB := b12Creature(g, opp.ID, "Wall B", "Creature — Wall", 0, 4)
	advanceTo(t, g, game.StepDeclareAttackers)

	// Its own controller is bound too: the line has no "you".
	_, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: a, Target: opp.ID}, {Attacker: b, Target: opp.ID}})
	clAttackRefusal(t, err)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: a, Target: opp.ID}}); err != nil {
		t.Fatalf("a one-creature swing: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blkA, a); err != nil {
		t.Fatalf("the one blocker: %v", err)
	}
	brRefusal(t, g.DeclareBlocker(blkB, a), game.BlockReasonDeclarationLimit)
}

// --- Caverns of Despair -------------------------------------------

func TestCavernsOfDespairAllowsTwoAttackersAndTwoBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, opp.ID, "Caverns of Despair", "World Enchantment", cavernsOfDespairOracle, 0, 0)
	var atk []uuid.UUID
	for i := 0; i < 3; i++ {
		atk = append(atk, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	var blk []uuid.UUID
	for i := 0; i < 3; i++ {
		blk = append(blk, b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4))
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range atk[:2] {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("two attackers are legal: %v", err)
		}
	}
	if n := len(clAgreeOnAttacks(t, g, me.ID)); n != 0 {
		t.Errorf("%d attacks offered past two", n)
	}
	if le := clAttackRefusal(t, g.DeclareAttacker(atk[2], opp.ID)); le.Max != 2 {
		t.Errorf("refusal names bound %d, want 2", le.Max)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: blk[0], Attacker: atk[0]}, {Blocker: blk[1], Attacker: atk[1]},
	}); err != nil {
		t.Fatalf("two blockers are legal: %v", err)
	}
	if n := clAgreeOnBlocks(t, g, opp.ID); n != 0 {
		t.Errorf("%d blocks offered past two", n)
	}
	brRefusal(t, g.DeclareBlocker(blk[2], atk[0]), game.BlockReasonDeclarationLimit)
}

// --- Crawlspace ---------------------------------------------------

// Multiplayer: two creatures may attack the Crawlspace player; the
// enumerator stops offering that seat at two and keeps offering the
// others, and the verb agrees.
func TestCrawlspaceLimitsAttackersPerDefendingPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, crawl, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b12Push(g, crawl.ID, "Crawlspace", "Artifact", crawlspaceOracle, 0, 0)
	var bears []uuid.UUID
	for i := 0; i < 4; i++ {
		bears = append(bears, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	advanceTo(t, g, game.StepDeclareAttackers)

	if at := clTargets(clAgreeOnAttacks(t, g, me.ID)); at[crawl.ID] != 4 || at[other.ID] != 4 {
		t.Fatalf("before any attack every creature is offered at every seat: %v", at)
	}
	for _, id := range bears[:2] {
		if err := g.DeclareAttacker(id, crawl.ID); err != nil {
			t.Fatalf("two attackers at the Crawlspace player: %v", err)
		}
	}
	at := clTargets(clAgreeOnAttacks(t, g, me.ID))
	if at[crawl.ID] != 0 {
		t.Errorf("%d attacks at the Crawlspace player offered past two", at[crawl.ID])
	}
	if at[other.ID] != 2 || at[g.Seats[3].ID] != 2 {
		t.Errorf("the other opponents stopped being offered: %v", at)
	}
	le := clAttackRefusal(t, g.DeclareAttacker(bears[2], crawl.ID))
	if le.Defender != crawl.ID || le.SourceName != "Crawlspace" {
		t.Errorf("refusal = %+v", le)
	}
	for _, id := range bears[2:] {
		if err := g.DeclareAttacker(id, other.ID); err != nil {
			t.Fatalf("another opponent may still be attacked: %v", err)
		}
	}
}

// "You" is the Crawlspace's controller: it does not limit attacks on
// its controller's opponents when its controller is the one attacking.
func TestCrawlspaceDoesNotLimitItsControllersAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Crawlspace", "Artifact", crawlspaceOracle, 0, 0)
	var bears []uuid.UUID
	for i := 0; i < 3; i++ {
		bears = append(bears, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range bears {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("my own Crawlspace limited my attack: %v", err)
		}
	}
}

// --- Judoon Enforcers ---------------------------------------------

func TestJudoonEnforcersAllowsOneAttackerAtItsController(t *testing.T) {
	g := newCatalogGame(t)
	me, judoon, other := g.Seats[0], g.Seats[1], g.Seats[2]
	enforcers := b12Push(g, judoon.ID, "Judoon Enforcers", "Creature — Alien Rhino Soldier", judoonEnforcersOracle, 8, 8)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	if !game.HasKeyword(mustBattlefieldCard(t, g, enforcers), "trample") {
		t.Error("Judoon Enforcers has trample")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, judoon.ID); err != nil {
		t.Fatalf("one attacker: %v", err)
	}
	if at := clTargets(clAgreeOnAttacks(t, g, me.ID)); at[judoon.ID] != 0 || at[other.ID] != 1 {
		t.Errorf("offers after one attacker: %v", at)
	}
	if le := clAttackRefusal(t, g.DeclareAttacker(b, judoon.ID)); le.Max != 1 {
		t.Errorf("refusal = %+v", le)
	}
	if err := g.DeclareAttacker(b, other.ID); err != nil {
		t.Fatalf("another opponent: %v", err)
	}
}

func TestJudoonEnforcersCanBeSuspended(t *testing.T) {
	spec, ok := Lookup(judoonEnforcersOracle)
	if !ok {
		t.Fatal("Judoon Enforcers is not registered")
	}
	if len(spec.SpecialActions) != 1 || spec.SpecialActions[0].Kind != game.SpecialActionSuspend ||
		spec.SpecialActions[0].Counters != 6 || spec.SpecialActions[0].Cost != "{1}{R}{W}" {
		t.Errorf("suspend = %+v, want Suspend 6—{1}{R}{W}", spec.SpecialActions)
	}
}
