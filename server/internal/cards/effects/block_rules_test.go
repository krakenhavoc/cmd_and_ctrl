package effects

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// block_rules_test.go — #750's card half: Spec.BlockRules and a token
// template's BlockRules slot, built from block_rules.go's constructors
// and read by the engine through CatalogBlockRules (ADR 0045 addendum,
// Decisions 11-13 and the 2026-09-24 amendment, Decisions 40-42).
//
// Every shape is pinned the same three ways, because ADR 0045 §3's
// invariant is that the three never disagree:
//
//   - the VERB: a legal block is accepted, an illegal one is refused
//     with its reason, the card's printed clause and the permanent
//     that carries the rule;
//   - the ENUMERATOR: legal.EnumerateFor offers exactly the blocks the
//     verb accepts;
//   - the VIEW: block_decision_seats (the #328 signal) says the seat
//     owes a decision only when some block is legal.

const (
	prowlersHelmOracle       = "603ef53f-6191-4229-b85f-adad6f4503ce"
	legolasGreenleafOracle   = "beacbb51-51fe-47f7-a612-02cd93fbffdd"
	championOfLambholtOracle = "c549b0fd-1e08-4873-952e-a14dc45a0fd2"
	gingerbruteOracle        = "10b8d4c7-7553-4d76-b643-d98b80701e13"
	thievesToolsOracle       = "7fe361ef-a168-4847-92a2-21c1661aac06"
	vorracBattlehornsOracle  = "89750d72-d1c0-4c8a-aa72-fdd48570aa92"
	rampagingCeratopsOracle  = "91d01119-9b3e-4629-8cf0-bbca2c19a7cc"
	canopyCoverOracle        = "5b84101e-7e23-437d-835c-409bc061ecbb"
)

// brCreature seeds a non-catalog creature with keywords, able to
// attack and block this turn.
func brCreature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
		Keywords: keywords,
	})
}

// brAttach attaches an Equipment or Aura without paying for the equip.
func brAttach(t *testing.T, g *game.Game, attachment, host uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(attachment, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
}

// brAttack declares `attackers` against seat 1 and parks the cursor on
// declare blockers.
func brAttack(t *testing.T, g *game.Game, attackers ...uuid.UUID) {
	t.Helper()
	declareAttack(t, g, g.Seats[1].ID, attackers...)
	advanceTo(t, g, game.StepDeclareBlockers)
}

// brRefusal asserts `err` is an illegal-block refusal with `reason`
// and returns it.
func brRefusal(t *testing.T, err error, reason game.BlockReason) *game.BlockRefusedError {
	t.Helper()
	if err == nil {
		t.Fatalf("the block was accepted; want a %q refusal", reason)
	}
	if !errors.Is(err, game.ErrIllegalBlock) {
		t.Fatalf("refusal does not wrap ErrIllegalBlock: %v", err)
	}
	var refused *game.BlockRefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("refusal is %T, want *BlockRefusedError: %v", err, err)
	}
	if refused.Reason != reason {
		t.Fatalf("reason %q, want %q (%s)", refused.Reason, reason, refused.Sentence(uuid.Nil))
	}
	return refused
}

// brOffered is every (blocker, attacker) pair the enumerator offers
// `seat`, one entry per move: a single block is one pair, a grouped
// block (menace, a minimum) is all of its pairs in one entry.
func brOffered(t *testing.T, g *game.Game, seat uuid.UUID) [][]game.BlockDeclaration {
	t.Helper()
	var out [][]game.BlockDeclaration
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Kind != legal.KindBlock {
			continue
		}
		var one struct {
			Blocker  string `json:"blocker"`
			Attacker string `json:"attacker"`
			Blocks   []struct {
				Blocker  string `json:"blocker"`
				Attacker string `json:"attacker"`
			} `json:"blocks"`
		}
		if err := json.Unmarshal(m.Params, &one); err != nil {
			t.Fatalf("block params %s: %v", m.Params, err)
		}
		var set []game.BlockDeclaration
		if one.Blocker != "" {
			set = append(set, game.BlockDeclaration{Blocker: uuid.MustParse(one.Blocker), Attacker: uuid.MustParse(one.Attacker)})
		}
		for _, b := range one.Blocks {
			set = append(set, game.BlockDeclaration{Blocker: uuid.MustParse(b.Blocker), Attacker: uuid.MustParse(b.Attacker)})
		}
		out = append(out, set)
	}
	return out
}

// brOffers reports whether some offered move blocks `attacker` with
// `blocker`.
func brOffers(offered [][]game.BlockDeclaration, blocker, attacker uuid.UUID) bool {
	for _, set := range offered {
		for _, d := range set {
			if d.Blocker == blocker && d.Attacker == attacker {
				return true
			}
		}
	}
	return false
}

// brViewOwes is the wire's block_decision_seats for seat 1.
func brViewOwes(g *game.Game) bool {
	for _, s := range protocol.ViewOfGame(g).Turn.BlockDecisionSeats {
		if s == 1 {
			return true
		}
	}
	return false
}

// --- CantBeBlockedExceptBy on OnAttached: Prowler's Helm ---------

func TestProwlersHelmLetsOnlyWallsBlockTheEquippedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	helm := b12Push(g, me.ID, "Prowler's Helm", "Artifact — Equipment", prowlersHelmOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	wall := b12Creature(g, opp.ID, "Wall of Stone", "Creature — Wall", 0, 8)
	elf := b12Creature(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	brAttach(t, g, helm, bear)
	brAttack(t, g, bear, other)

	offered := brOffered(t, g, opp.ID)
	if brOffers(offered, elf, bear) {
		t.Error("the enumerator offers the Elf on the Helm's creature, which the verb refuses")
	}
	if !brOffers(offered, wall, bear) {
		t.Error("the enumerator does not offer the Wall on the Helm's creature")
	}
	if !brOffers(offered, elf, other) {
		t.Error("the rule bound the creature that is NOT wearing the Helm")
	}

	refused := brRefusal(t, g.DeclareBlocker(elf, bear), game.BlockReasonCantBeBlockedExceptBy)
	if refused.Source != helm {
		t.Errorf("the refusal names %s, want the Helm %s", refused.Source, helm)
	}
	if refused.Label != "Walls" || !strings.Contains(refused.Sentence(uuid.Nil), "except by Walls") {
		t.Errorf("the refusal does not name the clause: %q", refused.Sentence(uuid.Nil))
	}
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("a Wall blocks the Helm's creature: %v", err)
	}
	if err := g.DeclareBlocker(elf, other); err != nil {
		t.Fatalf("the Elf blocks the creature without the Helm: %v", err)
	}
}

// With nothing but a non-Wall to block with, the defender owes no
// decision — the view and the enumerator agree with the verb.
func TestProwlersHelmLeavesANonWallDefenderNothingToDecide(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	helm := b12Push(g, me.ID, "Prowler's Helm", "Artifact — Equipment", prowlersHelmOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	brAttach(t, g, helm, bear)
	brAttack(t, g, bear)

	if n := len(brOffered(t, g, opp.ID)); n != 0 {
		t.Errorf("the enumerator offers %d blocks the verb refuses", n)
	}
	if g.SeatOwesBlockDecision(opp.ID) || brViewOwes(g) {
		t.Error("block_decision_seats says the defender owes a decision it cannot make")
	}
}

// --- CantBeBlockedExceptBy with a keyword predicate: Canopy Cover --

func TestCanopyCoverLetsOnlyFlyingOrReachBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cover := b12Push(g, me.ID, "Canopy Cover", "Enchantment — Aura", canopyCoverOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	flyer := brCreature(g, opp.ID, "Wind Drake", "Creature — Drake", 2, 2, "flying")
	spider := brCreature(g, opp.ID, "Giant Spider", "Creature — Spider", 2, 4, "reach")
	ground := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	brAttach(t, g, cover, bear)
	brAttack(t, g, bear)

	brRefusal(t, g.DeclareBlocker(ground, bear), game.BlockReasonCantBeBlockedExceptBy)
	if err := g.DeclareBlocker(flyer, bear); err != nil {
		t.Fatalf("a flyer blocks: %v", err)
	}
	if err := g.DeclareBlocker(spider, bear); err != nil {
		t.Fatalf("a reach creature blocks: %v", err)
	}
}

// --- CantBeBlockedBy on OnSelf: Legolas Greenleaf -----------------

func TestLegolasCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legolas := b12Push(g, me.ID, "Legolas Greenleaf", "Legendary Creature — Elf Archer", legolasGreenleafOracle, 2, 2)
	small := b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	pumped := b12Creature(g, opp.ID, "Pumped Bears", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	companion := b12Creature(g, me.ID, "Elvish Archer", "Creature — Elf Archer", 2, 2)
	brAttack(t, g, legolas, companion)
	// Power is read live at declaration, counters included.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(pumped, game.CounterPlusOne, 1) })

	offered := brOffered(t, g, opp.ID)
	if brOffers(offered, small, legolas) || !brOffers(offered, big, legolas) || !brOffers(offered, pumped, legolas) {
		t.Errorf("the enumerator disagrees with the rule: %v", offered)
	}
	refused := brRefusal(t, g.DeclareBlocker(small, legolas), game.BlockReasonCantBeBlockedBy)
	if refused.Source != legolas || refused.Label != "creatures with power 2 or less" {
		t.Errorf("refusal source %s label %q", refused.Source, refused.Label)
	}
	if err := g.DeclareBlocker(pumped, legolas); err != nil {
		t.Fatalf("a 2/2 with a +1/+1 counter has power 3 and blocks: %v", err)
	}
	// The rule is about Legolas, not about every attacker at the table.
	if !brOffers(offered, small, companion) {
		t.Error("the enumerator withholds a block on the attacker beside Legolas")
	}
	if err := g.DeclareBlocker(small, companion); err != nil {
		t.Fatalf("the rule bound the attacker beside Legolas: %v", err)
	}
}

// CR 613.1f: the rule is Legolas's own ability, so Legolas losing all
// abilities loses it.
func TestLegolasWithNoAbilitiesCanBeBlockedByAnything(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legolas := b12Push(g, me.ID, "Legolas Greenleaf", "Legendary Creature — Elf Archer", legolasGreenleafOracle, 2, 2)
	small := b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(legolas),
			[]game.Mod{game.LoseAllAbilitiesMod()}, g.UntilEndOfTurnDuration(), "loses all abilities")
	})
	brAttack(t, g, legolas)
	if err := g.DeclareBlocker(small, legolas); err != nil {
		t.Fatalf("a Legolas with no abilities has no rule: %v", err)
	}
}

// --- CantBlockAttackers: Champion of Lambholt ---------------------

func TestChampionOfLambholtStopsSmallerCreaturesBlockingYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champion := b12Push(g, me.ID, "Champion of Lambholt", "Creature — Human Warrior", championOfLambholtOracle, 1, 1)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(champion, game.CounterPlusOne, 2) }) // power 3
	small := b12Creature(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", 2, 2)
	equal := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	brAttack(t, g, champion, bear)

	offered := brOffered(t, g, opp.ID)
	if brOffers(offered, small, bear) || brOffers(offered, small, champion) {
		t.Errorf("the enumerator offers a power-2 blocker against the Champion's creatures: %v", offered)
	}
	refused := brRefusal(t, g.DeclareBlocker(small, bear), game.BlockReasonCantBlockAttacker)
	if refused.Source != champion {
		t.Errorf("the refusal names %s, want the Champion", refused.Source)
	}
	if !strings.Contains(refused.Sentence(uuid.Nil), "power less than Champion of Lambholt's") {
		t.Errorf("the sentence does not say why: %q", refused.Sentence(uuid.Nil))
	}
	// "Less than" is strict: equal power blocks.
	if err := g.DeclareBlocker(equal, bear); err != nil {
		t.Fatalf("a creature with power equal to the Champion's blocks: %v", err)
	}
}

// "Creatures YOU control": a Champion on the DEFENDER's side does not
// stop its controller's small creatures blocking an opponent's
// attacker.
func TestChampionOfLambholtDoesNotBindTheOpponentsAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champion := b12Push(g, opp.ID, "Champion of Lambholt", "Creature — Human Warrior", championOfLambholtOracle, 1, 1)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(champion, game.CounterPlusOne, 2) }) // power 3
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	small := b12Creature(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", 2, 2)
	brAttack(t, g, bear)
	if !brOffers(brOffered(t, g, opp.ID), small, bear) {
		t.Error("the enumerator withholds a block the Champion does not forbid")
	}
	if err := g.DeclareBlocker(small, bear); err != nil {
		t.Fatalf("the Champion bound an attacker its controller does not control: %v", err)
	}
}

// The threshold is the Champion's power NOW: the enter trigger's
// counter raises it.
func TestChampionOfLambholtGrowsWhenAnotherCreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	champion := b12Push(g, me.ID, "Champion of Lambholt", "Creature — Human Warrior", championOfLambholtOracle, 1, 1)
	enterFromHand(t, g, me.ID, "Grizzly Bears", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, champion)
	if c.Counters[game.CounterPlusOne] != 1 {
		t.Fatalf("another creature entered: %d counters, want 1", c.Counters[game.CounterPlusOne])
	}
}

// --- CantBeBlockedWhile: Thieves' Tools ---------------------------

func TestThievesToolsUnblockableOnlyWhilePowerIsThreeOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tools := b12Push(g, me.ID, "Thieves' Tools", "Artifact — Equipment", thievesToolsOracle, 0, 0)
	small := b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	big := b12Creature(g, me.ID, "Craw Wurm", "Creature — Wurm", 6, 4)
	blocker := b12Creature(g, opp.ID, "Wall of Stone", "Creature — Wall", 0, 8)
	brAttach(t, g, tools, small)
	brAttack(t, g, small, big)

	if brOffers(brOffered(t, g, opp.ID), blocker, small) {
		t.Error("the enumerator offers a block on an unblockable creature")
	}
	refused := brRefusal(t, g.DeclareBlocker(blocker, small), game.BlockReasonCantBeBlocked)
	if refused.Source != tools {
		t.Errorf("the refusal names %s, want the Tools", refused.Source)
	}

	// Move the Tools onto the 6-power creature mid-combat: the rule
	// follows the Equipment, and its condition fails there.
	brAttach(t, g, tools, big)
	if err := g.DeclareBlocker(blocker, big); err != nil {
		t.Fatalf("power 6 is not 3 or less: %v", err)
	}
}

func TestThievesToolsMakesATreasureWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	enterFromHand(t, g, me.ID, "Thieves' Tools", "Artifact — Equipment", thievesToolsOracle)
	passPriorityAroundTable(t, g)
	if onBattlefieldNamed(g, "Treasure") != 1 {
		t.Error("entering makes one Treasure")
	}
}

// --- MaxBlockers: Vorrac Battlehorns ------------------------------

func TestVorracBattlehornsRefusesASecondBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	horns := b12Push(g, me.ID, "Vorrac Battlehorns", "Artifact — Equipment", vorracBattlehornsOracle, 0, 0)
	wurm := b12Creature(g, me.ID, "Craw Wurm", "Creature — Wurm", 6, 4)
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	brAttach(t, g, horns, wurm)
	brAttack(t, g, wurm)

	if !game.HasKeyword(mustBattlefieldCard(t, g, wurm), "trample") {
		t.Error("the equipped creature has trample")
	}
	for _, set := range brOffered(t, g, opp.ID) {
		if len(set) > 1 {
			t.Errorf("the enumerator offers a %d-creature block", len(set))
		}
	}
	refused := brRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: wurm}, {Blocker: b, Attacker: wurm},
	}), game.BlockReasonTooManyBlockers)
	if refused.N != 1 {
		t.Errorf("the refusal carries max %d, want 1", refused.N)
	}
	if err := g.DeclareBlocker(a, wurm); err != nil {
		t.Fatalf("one blocker is legal: %v", err)
	}
	brRefusal(t, g.DeclareBlocker(b, wurm), game.BlockReasonTooManyBlockers)
}

// Menace's minimum of two above the Horns' maximum of one: no block is
// legal, none is offered, and the defender owes nothing.
func TestVorracBattlehornsOnAMenaceCreatureMakesItUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	horns := b12Push(g, me.ID, "Vorrac Battlehorns", "Artifact — Equipment", vorracBattlehornsOracle, 0, 0)
	menace := brCreature(g, me.ID, "Menacer", "Creature — Ogre", 3, 3, "menace")
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	brAttach(t, g, horns, menace)
	brAttack(t, g, menace)

	if n := len(brOffered(t, g, opp.ID)); n != 0 {
		t.Errorf("the enumerator offers %d blocks on an unblockable creature", n)
	}
	if g.SeatOwesBlockDecision(opp.ID) || brViewOwes(g) {
		t.Error("the defender owes no decision")
	}
	brRefusal(t, g.DeclareBlocker(a, menace), game.BlockReasonTooFewBlockers)
	brRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: menace}, {Blocker: b, Attacker: menace},
	}), game.BlockReasonTooManyBlockers)
}

// --- MinBlockers: Rampaging Ceratops ------------------------------

func TestRampagingCeratopsNeedsThreeBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ceratops := b12Push(g, me.ID, "Rampaging Ceratops", "Creature — Dinosaur", rampagingCeratopsOracle, 5, 4)
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := b12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	brAttack(t, g, ceratops)

	offered := brOffered(t, g, opp.ID)
	if len(offered) == 0 {
		t.Fatal("the enumerator offers no three-creature block")
	}
	for _, set := range offered {
		if len(set) != 3 {
			t.Errorf("the enumerator offers a %d-creature block", len(set))
		}
	}
	if !g.SeatOwesBlockDecision(opp.ID) || !brViewOwes(g) {
		t.Error("three creatures can block, so the defender owes a decision")
	}
	refused := brRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: ceratops}, {Blocker: b, Attacker: ceratops},
	}), game.BlockReasonTooFewBlockers)
	if refused.N != 3 || !strings.Contains(refused.Sentence(uuid.Nil), "fewer than three") {
		t.Errorf("refusal N %d, sentence %q", refused.N, refused.Sentence(uuid.Nil))
	}
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: ceratops}, {Blocker: b, Attacker: ceratops}, {Blocker: c, Attacker: ceratops},
	}); err != nil {
		t.Fatalf("three blockers are legal: %v", err)
	}
}

// With only two creatures the Ceratops can't be blocked at all.
func TestRampagingCeratopsAgainstTwoCreaturesOwesNoDecision(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ceratops := b12Push(g, me.ID, "Rampaging Ceratops", "Creature — Dinosaur", rampagingCeratopsOracle, 5, 4)
	b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	brAttack(t, g, ceratops)
	if n := len(brOffered(t, g, opp.ID)); n != 0 {
		t.Errorf("the enumerator offers %d blocks the verb refuses", n)
	}
	if g.SeatOwesBlockDecision(opp.ID) || brViewOwes(g) {
		t.Error("the defender owes no decision")
	}
}

// --- BlockRuleUntilEOT: Gingerbrute -------------------------------

// gingerbruteActivate pays {1} and resolves the evasion ability.
func gingerbruteActivate(t *testing.T, g *game.Game, controller, brute uuid.UUID) {
	t.Helper()
	for _, p := range g.Seats {
		if p.ID == controller {
			p.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
	}
	if err := g.ActivateCatalogAbility(controller, brute, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestGingerbruteCantBeBlockedExceptByHasteThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brute := b12Push(g, me.ID, "Gingerbrute", "Artifact Creature — Food Golem", gingerbruteOracle, 1, 1)
	slow := b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	fast := brCreature(g, opp.ID, "Raging Goblin", "Creature — Goblin", 1, 1, "haste")
	advanceTo(t, g, game.StepPrecombatMain)
	gingerbruteActivate(t, g, me.ID, brute)
	if len(g.TurnScopedBlockRules) != 1 {
		t.Fatalf("the ability registers one turn-scoped rule, got %d", len(g.TurnScopedBlockRules))
	}
	brAttack(t, g, brute)

	offered := brOffered(t, g, opp.ID)
	if brOffers(offered, slow, brute) || !brOffers(offered, fast, brute) {
		t.Errorf("the enumerator disagrees with the rule: %v", offered)
	}
	refused := brRefusal(t, g.DeclareBlocker(slow, brute), game.BlockReasonCantBeBlockedExceptBy)
	if refused.Label != "creatures with haste" {
		t.Errorf("label %q", refused.Label)
	}
	if err := g.DeclareBlocker(fast, brute); err != nil {
		t.Fatalf("a creature with haste blocks: %v", err)
	}
}

// Until end of turn: the sweep at the turn's end takes the rule, and
// the next turn anything may block.
func TestGingerbruteRuleEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	brute := b12Push(g, me.ID, "Gingerbrute", "Artifact Creature — Food Golem", gingerbruteOracle, 1, 1)
	advanceTo(t, g, game.StepPrecombatMain)
	gingerbruteActivate(t, g, me.ID, brute)
	advanceToUpkeepOf(t, g, 1)
	if n := len(g.TurnScopedBlockRules); n != 0 {
		t.Errorf("%d turn-scoped rules outlived the turn", n)
	}
}

// CR 611.2c / 400.7: the rule is pinned to the object at resolution;
// the same card back on the battlefield is a new object and blockable.
func TestGingerbruteRuleDoesNotFollowANewObject(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brute := b12Push(g, me.ID, "Gingerbrute", "Artifact Creature — Food Golem", gingerbruteOracle, 1, 1)
	slow := b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	gingerbruteActivate(t, g, me.ID, brute)
	flickerInResponse(t, g, brute)
	g.WithWriteLock(func() {
		c := findBattlefieldCardForTest(g, brute)
		c.SummonedThisTurn = false
	})
	brAttack(t, g, brute)
	if err := g.DeclareBlocker(slow, brute); err != nil {
		t.Fatalf("the rule followed the card to its new object: %v", err)
	}
}

// --- Departed Deckhand: a printed rule and a granted one ---------

func TestDepartedDeckhandCanBeBlockedOnlyBySpirits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	deckhand := b12Push(g, me.ID, "Departed Deckhand", "Creature — Spirit Pirate", departedDeckhandOracle, 2, 2)
	bear := b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	spirit := b12Creature(g, opp.ID, "Spirit", "Creature — Spirit", 1, 1)
	shifter := brCreature(g, opp.ID, "Changeling Outcast", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	brAttack(t, g, deckhand)

	brRefusal(t, g.DeclareBlocker(bear, deckhand), game.BlockReasonCantBeBlockedExceptBy)
	if err := g.DeclareBlocker(spirit, deckhand); err != nil {
		t.Fatalf("a Spirit blocks: %v", err)
	}
	if err := g.DeclareBlocker(shifter, deckhand); err != nil {
		t.Fatalf("a changeling is a Spirit: %v", err)
	}
}

func TestDepartedDeckhandGrantsItsEvasionToAnotherCreatureThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	deckhand := b12Push(g, me.ID, "Departed Deckhand", "Creature — Spirit Pirate", departedDeckhandOracle, 2, 2)
	giant := b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	bear := b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, deckhand, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: giant}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	brAttack(t, g, giant)
	brRefusal(t, g.DeclareBlocker(bear, giant), game.BlockReasonCantBeBlockedExceptBy)
}

// "Another": the Deckhand can't target itself.
func TestDepartedDeckhandCannotTargetItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	deckhand := b12Push(g, me.ID, "Departed Deckhand", "Creature — Spirit Pirate", departedDeckhandOracle, 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, deckhand, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: deckhand}},
	}); err == nil {
		t.Fatal("the Deckhand targeted itself")
	}
}

// --- CantBlockOrBeBlockedBy on a token: Avatar Kuruk's Spirit -----

// brKurukSpirit mints Avatar Kuruk's Spirit token for `owner`.
func brKurukSpirit(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	tok := kurukSpiritToken()
	tok.InstanceID = uuid.New()
	tok.Owner, tok.Controller = owner, owner
	return pushBattlefieldCardWithTimestamp(g, tok)
}

func TestKurukSpiritTokenCarriesItsBlockRule(t *testing.T) {
	tok := kurukSpiritToken()
	if tok.TokenKey == "" || tok.OracleID != "" {
		t.Fatalf("the token resolves through its own key: %+v", tok)
	}
	if n := len(game.CatalogBlockRules(game.CatalogKey(tok))); n != 2 {
		t.Errorf("the token's key carries %d block rules, want 2 (one per side)", n)
	}
	if !strings.Contains(game.TokenTextForCard(tok), "can't block or be blocked by non-Spirit creatures") {
		t.Errorf("the token's printed text is missing: %q", game.TokenTextForCard(tok))
	}
	// Forbidden Orchard's Spirit is a different token and prints nothing.
	if game.CatalogBlockRules(game.CatalogKey(TokenCard("1/1 colorless Spirit"))) != nil {
		t.Error("the plain Spirit row picked up Kuruk's rule")
	}
}

func TestKurukSpiritTokenCantBeBlockedByNonSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spirit := brKurukSpirit(t, g, me.ID)
	bear := b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirSpirit := b12Creature(g, opp.ID, "Spirit", "Creature — Spirit", 1, 1)
	brAttack(t, g, spirit)

	offered := brOffered(t, g, opp.ID)
	if brOffers(offered, bear, spirit) || !brOffers(offered, theirSpirit, spirit) {
		t.Errorf("the enumerator disagrees with the token's rule: %v", offered)
	}
	refused := brRefusal(t, g.DeclareBlocker(bear, spirit), game.BlockReasonCantBeBlockedBy)
	if refused.Source != spirit || refused.Label != "non-Spirit creatures" {
		t.Errorf("refusal source %s label %q", refused.Source, refused.Label)
	}
	if err := g.DeclareBlocker(theirSpirit, spirit); err != nil {
		t.Fatalf("a Spirit blocks the token: %v", err)
	}
}

func TestKurukSpiritTokenCantBlockNonSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	mySpirit := b12Creature(g, me.ID, "Spirit", "Creature — Spirit", 1, 1)
	token := brKurukSpirit(t, g, opp.ID)
	brAttack(t, g, bear, mySpirit)

	offered := brOffered(t, g, opp.ID)
	if brOffers(offered, token, bear) || !brOffers(offered, token, mySpirit) {
		t.Errorf("the enumerator disagrees with the token's rule: %v", offered)
	}
	refused := brRefusal(t, g.DeclareBlocker(token, bear), game.BlockReasonCantBlockAttacker)
	if refused.Source != token {
		t.Errorf("the refusal names %s, want the token", refused.Source)
	}
	if err := g.DeclareBlocker(token, mySpirit); err != nil {
		t.Fatalf("the token blocks a Spirit: %v", err)
	}
}

// The whole path: Avatar Kuruk's cast trigger makes the token, and the
// token on the battlefield is the one with the rule.
func TestAvatarKurukMakesTheSpiritWithTheRule(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Avatar Kuruk", "Legendary Creature — Avatar", theLegendOfKurukOracleID+"#1", 4, 3)
	castCatalogSpell(t, g, "Opt", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	id := findBattlefieldByName(g, "Spirit")
	if id == uuid.Nil {
		t.Fatal("no Spirit token")
	}
	c := mustBattlefieldCard(t, g, id)
	if len(game.CatalogBlockRules(game.CatalogAbilityKey(*c))) != 2 {
		t.Errorf("the minted Spirit carries no block rule (key %q)", game.CatalogAbilityKey(*c))
	}
}

// --- helpers --------------------------------------------------------

func mustBattlefieldCard(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	t.Fatalf("%s is not on the battlefield", id)
	return nil
}

func findBattlefieldCardForTest(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}
