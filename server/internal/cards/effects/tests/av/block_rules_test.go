package testsav

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- CantBeBlockedExceptBy on OnAttached: Prowler's Helm ---------

func TestProwlersHelmLetsOnlyWallsBlockTheEquippedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	helm := effectstest.B12Push(g, me.ID, "Prowler's Helm", "Artifact — Equipment", effectstest.ProwlersHelmOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	wall := effectstest.B12Creature(g, opp.ID, "Wall of Stone", "Creature — Wall", 0, 8)
	elf := effectstest.B12Creature(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	effectstest.BrAttach(t, g, helm, bear)
	effectstest.BrAttack(t, g, bear, other)

	offered := effectstest.BrOffered(t, g, opp.ID)
	if effectstest.BrOffers(offered, elf, bear) {
		t.Error("the enumerator offers the Elf on the Helm's creature, which the verb refuses")
	}
	if !effectstest.BrOffers(offered, wall, bear) {
		t.Error("the enumerator does not offer the Wall on the Helm's creature")
	}
	if !effectstest.BrOffers(offered, elf, other) {
		t.Error("the rule bound the creature that is NOT wearing the Helm")
	}

	refused := effectstest.BrRefusal(t, g.DeclareBlocker(elf, bear), game.BlockReasonCantBeBlockedExceptBy)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	helm := effectstest.B12Push(g, me.ID, "Prowler's Helm", "Artifact — Equipment", effectstest.ProwlersHelmOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	effectstest.BrAttach(t, g, helm, bear)
	effectstest.BrAttack(t, g, bear)

	if n := len(effectstest.BrOffered(t, g, opp.ID)); n != 0 {
		t.Errorf("the enumerator offers %d blocks the verb refuses", n)
	}
	if g.SeatOwesBlockDecision(opp.ID) || effectstest.BrViewOwes(g) {
		t.Error("block_decision_seats says the defender owes a decision it cannot make")
	}
}

// --- CantBeBlockedExceptBy with a keyword predicate: Canopy Cover --

func TestCanopyCoverLetsOnlyFlyingOrReachBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cover := effectstest.B12Push(g, me.ID, "Canopy Cover", "Enchantment — Aura", effectstest.CanopyCoverOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	flyer := effectstest.BrCreature(g, opp.ID, "Wind Drake", "Creature — Drake", 2, 2, "flying")
	spider := effectstest.BrCreature(g, opp.ID, "Giant Spider", "Creature — Spider", 2, 4, "reach")
	ground := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	effectstest.BrAttach(t, g, cover, bear)
	effectstest.BrAttack(t, g, bear)

	effectstest.BrRefusal(t, g.DeclareBlocker(ground, bear), game.BlockReasonCantBeBlockedExceptBy)
	if err := g.DeclareBlocker(flyer, bear); err != nil {
		t.Fatalf("a flyer blocks: %v", err)
	}
	if err := g.DeclareBlocker(spider, bear); err != nil {
		t.Fatalf("a reach creature blocks: %v", err)
	}
}

// --- CantBeBlockedBy on OnSelf: Legolas Greenleaf -----------------

func TestLegolasCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legolas := effectstest.B12Push(g, me.ID, "Legolas Greenleaf", "Legendary Creature — Elf Archer", effectstest.LegolasGreenleafOracle, 2, 2)
	small := effectstest.B12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	pumped := effectstest.B12Creature(g, opp.ID, "Pumped Bears", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	companion := effectstest.B12Creature(g, me.ID, "Elvish Archer", "Creature — Elf Archer", 2, 2)
	effectstest.BrAttack(t, g, legolas, companion)
	// Power is read live at declaration, counters included.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(pumped, game.CounterPlusOne, 1) })

	offered := effectstest.BrOffered(t, g, opp.ID)
	if effectstest.BrOffers(offered, small, legolas) || !effectstest.BrOffers(offered, big, legolas) || !effectstest.BrOffers(offered, pumped, legolas) {
		t.Errorf("the enumerator disagrees with the rule: %v", offered)
	}
	refused := effectstest.BrRefusal(t, g.DeclareBlocker(small, legolas), game.BlockReasonCantBeBlockedBy)
	if refused.Source != legolas || refused.Label != "creatures with power 2 or less" {
		t.Errorf("refusal source %s label %q", refused.Source, refused.Label)
	}
	if err := g.DeclareBlocker(pumped, legolas); err != nil {
		t.Fatalf("a 2/2 with a +1/+1 counter has power 3 and blocks: %v", err)
	}
	// The rule is about Legolas, not about every attacker at the table.
	if !effectstest.BrOffers(offered, small, companion) {
		t.Error("the enumerator withholds a block on the attacker beside Legolas")
	}
	if err := g.DeclareBlocker(small, companion); err != nil {
		t.Fatalf("the rule bound the attacker beside Legolas: %v", err)
	}
}

// CR 613.1f: the rule is Legolas's own ability, so Legolas losing all
// abilities loses it.
func TestLegolasWithNoAbilitiesCanBeBlockedByAnything(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legolas := effectstest.B12Push(g, me.ID, "Legolas Greenleaf", "Legendary Creature — Elf Archer", effectstest.LegolasGreenleafOracle, 2, 2)
	small := effectstest.B12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(legolas),
			[]game.Mod{game.LoseAllAbilitiesMod()}, g.UntilEndOfTurnDuration(), "loses all abilities")
	})
	effectstest.BrAttack(t, g, legolas)
	if err := g.DeclareBlocker(small, legolas); err != nil {
		t.Fatalf("a Legolas with no abilities has no rule: %v", err)
	}
}

// --- CantBlockAttackers: Champion of Lambholt ---------------------

func TestChampionOfLambholtStopsSmallerCreaturesBlockingYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champion := effectstest.B12Push(g, me.ID, "Champion of Lambholt", "Creature — Human Warrior", effectstest.ChampionOfLambholtOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(champion, game.CounterPlusOne, 2) }) // power 3
	small := effectstest.B12Creature(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", 2, 2)
	equal := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	effectstest.BrAttack(t, g, champion, bear)

	offered := effectstest.BrOffered(t, g, opp.ID)
	if effectstest.BrOffers(offered, small, bear) || effectstest.BrOffers(offered, small, champion) {
		t.Errorf("the enumerator offers a power-2 blocker against the Champion's creatures: %v", offered)
	}
	refused := effectstest.BrRefusal(t, g.DeclareBlocker(small, bear), game.BlockReasonCantBlockAttacker)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champion := effectstest.B12Push(g, opp.ID, "Champion of Lambholt", "Creature — Human Warrior", effectstest.ChampionOfLambholtOracle, 1, 1)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(champion, game.CounterPlusOne, 2) }) // power 3
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	small := effectstest.B12Creature(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", 2, 2)
	effectstest.BrAttack(t, g, bear)
	if !effectstest.BrOffers(effectstest.BrOffered(t, g, opp.ID), small, bear) {
		t.Error("the enumerator withholds a block the Champion does not forbid")
	}
	if err := g.DeclareBlocker(small, bear); err != nil {
		t.Fatalf("the Champion bound an attacker its controller does not control: %v", err)
	}
}

// The threshold is the Champion's power NOW: the enter trigger's
// counter raises it.
func TestChampionOfLambholtGrowsWhenAnotherCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	champion := effectstest.B12Push(g, me.ID, "Champion of Lambholt", "Creature — Human Warrior", effectstest.ChampionOfLambholtOracle, 1, 1)
	effectstest.EnterFromHand(t, g, me.ID, "Grizzly Bears", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, champion)
	if c.Counters[game.CounterPlusOne] != 1 {
		t.Fatalf("another creature entered: %d counters, want 1", c.Counters[game.CounterPlusOne])
	}
}

// --- CantBeBlockedWhile: Thieves' Tools ---------------------------

func TestThievesToolsUnblockableOnlyWhilePowerIsThreeOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tools := effectstest.B12Push(g, me.ID, "Thieves' Tools", "Artifact — Equipment", effectstest.ThievesToolsOracle, 0, 0)
	small := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	big := effectstest.B12Creature(g, me.ID, "Craw Wurm", "Creature — Wurm", 6, 4)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall of Stone", "Creature — Wall", 0, 8)
	effectstest.BrAttach(t, g, tools, small)
	effectstest.BrAttack(t, g, small, big)

	if effectstest.BrOffers(effectstest.BrOffered(t, g, opp.ID), blocker, small) {
		t.Error("the enumerator offers a block on an unblockable creature")
	}
	refused := effectstest.BrRefusal(t, g.DeclareBlocker(blocker, small), game.BlockReasonCantBeBlocked)
	if refused.Source != tools {
		t.Errorf("the refusal names %s, want the Tools", refused.Source)
	}

	// Move the Tools onto the 6-power creature mid-combat: the rule
	// follows the Equipment, and its condition fails there.
	effectstest.BrAttach(t, g, tools, big)
	if err := g.DeclareBlocker(blocker, big); err != nil {
		t.Fatalf("power 6 is not 3 or less: %v", err)
	}
}

func TestThievesToolsMakesATreasureWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.EnterFromHand(t, g, me.ID, "Thieves' Tools", "Artifact — Equipment", effectstest.ThievesToolsOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldNamed(g, "Treasure") != 1 {
		t.Error("entering makes one Treasure")
	}
}

// --- MaxBlockers: Vorrac Battlehorns ------------------------------

func TestVorracBattlehornsRefusesASecondBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	horns := effectstest.B12Push(g, me.ID, "Vorrac Battlehorns", "Artifact — Equipment", effectstest.VorracBattlehornsOracle, 0, 0)
	wurm := effectstest.B12Creature(g, me.ID, "Craw Wurm", "Creature — Wurm", 6, 4)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.BrAttach(t, g, horns, wurm)
	effectstest.BrAttack(t, g, wurm)

	if !game.HasKeyword(effectstest.MustBattlefieldCard(t, g, wurm), "trample") {
		t.Error("the equipped creature has trample")
	}
	for _, set := range effectstest.BrOffered(t, g, opp.ID) {
		if len(set) > 1 {
			t.Errorf("the enumerator offers a %d-creature block", len(set))
		}
	}
	refused := effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: wurm}, {Blocker: b, Attacker: wurm},
	}), game.BlockReasonTooManyBlockers)
	if refused.N != 1 {
		t.Errorf("the refusal carries max %d, want 1", refused.N)
	}
	if err := g.DeclareBlocker(a, wurm); err != nil {
		t.Fatalf("one blocker is legal: %v", err)
	}
	effectstest.BrRefusal(t, g.DeclareBlocker(b, wurm), game.BlockReasonTooManyBlockers)
}

// Menace's minimum of two above the Horns' maximum of one: no block is
// legal, none is offered, and the defender owes nothing.
func TestVorracBattlehornsOnAMenaceCreatureMakesItUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	horns := effectstest.B12Push(g, me.ID, "Vorrac Battlehorns", "Artifact — Equipment", effectstest.VorracBattlehornsOracle, 0, 0)
	menace := effectstest.BrCreature(g, me.ID, "Menacer", "Creature — Ogre", 3, 3, "menace")
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.BrAttach(t, g, horns, menace)
	effectstest.BrAttack(t, g, menace)

	if n := len(effectstest.BrOffered(t, g, opp.ID)); n != 0 {
		t.Errorf("the enumerator offers %d blocks on an unblockable creature", n)
	}
	if g.SeatOwesBlockDecision(opp.ID) || effectstest.BrViewOwes(g) {
		t.Error("the defender owes no decision")
	}
	effectstest.BrRefusal(t, g.DeclareBlocker(a, menace), game.BlockReasonTooFewBlockers)
	effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: menace}, {Blocker: b, Attacker: menace},
	}), game.BlockReasonTooManyBlockers)
}

// --- MinBlockers: Rampaging Ceratops ------------------------------

func TestRampagingCeratopsNeedsThreeBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ceratops := effectstest.B12Push(g, me.ID, "Rampaging Ceratops", "Creature — Dinosaur", effectstest.RampagingCeratopsOracle, 5, 4)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, ceratops)

	offered := effectstest.BrOffered(t, g, opp.ID)
	if len(offered) == 0 {
		t.Fatal("the enumerator offers no three-creature block")
	}
	for _, set := range offered {
		if len(set) != 3 {
			t.Errorf("the enumerator offers a %d-creature block", len(set))
		}
	}
	if !g.SeatOwesBlockDecision(opp.ID) || !effectstest.BrViewOwes(g) {
		t.Error("three creatures can block, so the defender owes a decision")
	}
	refused := effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ceratops := effectstest.B12Push(g, me.ID, "Rampaging Ceratops", "Creature — Dinosaur", effectstest.RampagingCeratopsOracle, 5, 4)
	effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, ceratops)
	if n := len(effectstest.BrOffered(t, g, opp.ID)); n != 0 {
		t.Errorf("the enumerator offers %d blocks the verb refuses", n)
	}
	if g.SeatOwesBlockDecision(opp.ID) || effectstest.BrViewOwes(g) {
		t.Error("the defender owes no decision")
	}
}

func TestGingerbruteCantBeBlockedExceptByHasteThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brute := effectstest.B12Push(g, me.ID, "Gingerbrute", "Artifact Creature — Food Golem", effectstest.GingerbruteOracle, 1, 1)
	slow := effectstest.B12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	fast := effectstest.BrCreature(g, opp.ID, "Raging Goblin", "Creature — Goblin", 1, 1, "haste")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.GingerbruteActivate(t, g, me.ID, brute)
	if n := effectstest.ScopedBlockRuleCount(g); n != 1 {
		t.Fatalf("the ability registers one scoped block rule, got %d", n)
	}
	effectstest.BrAttack(t, g, brute)

	offered := effectstest.BrOffered(t, g, opp.ID)
	if effectstest.BrOffers(offered, slow, brute) || !effectstest.BrOffers(offered, fast, brute) {
		t.Errorf("the enumerator disagrees with the rule: %v", offered)
	}
	refused := effectstest.BrRefusal(t, g.DeclareBlocker(slow, brute), game.BlockReasonCantBeBlockedExceptBy)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	brute := effectstest.B12Push(g, me.ID, "Gingerbrute", "Artifact Creature — Food Golem", effectstest.GingerbruteOracle, 1, 1)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.GingerbruteActivate(t, g, me.ID, brute)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if n := effectstest.ScopedBlockRuleCount(g); n != 0 {
		t.Errorf("%d scoped block rules outlived the turn", n)
	}
}

// CR 611.2c / 400.7: the rule is pinned to the object at resolution;
// the same card back on the battlefield is a new object and blockable.
func TestGingerbruteRuleDoesNotFollowANewObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brute := effectstest.B12Push(g, me.ID, "Gingerbrute", "Artifact Creature — Food Golem", effectstest.GingerbruteOracle, 1, 1)
	slow := effectstest.B12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.GingerbruteActivate(t, g, me.ID, brute)
	effectstest.FlickerInResponse(t, g, brute)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, brute)
		c.SummonedThisTurn = false
	})
	effectstest.BrAttack(t, g, brute)
	if err := g.DeclareBlocker(slow, brute); err != nil {
		t.Fatalf("the rule followed the card to its new object: %v", err)
	}
}

// --- Departed Deckhand: a printed rule and a granted one ---------

func TestDepartedDeckhandCanBeBlockedOnlyBySpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	deckhand := effectstest.B12Push(g, me.ID, "Departed Deckhand", "Creature — Spirit Pirate", effectstest.DepartedDeckhandOracle, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	spirit := effectstest.B12Creature(g, opp.ID, "Spirit", "Creature — Spirit", 1, 1)
	shifter := effectstest.BrCreature(g, opp.ID, "Changeling Outcast", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	effectstest.BrAttack(t, g, deckhand)

	effectstest.BrRefusal(t, g.DeclareBlocker(bear, deckhand), game.BlockReasonCantBeBlockedExceptBy)
	if err := g.DeclareBlocker(spirit, deckhand); err != nil {
		t.Fatalf("a Spirit blocks: %v", err)
	}
	if err := g.DeclareBlocker(shifter, deckhand); err != nil {
		t.Fatalf("a changeling is a Spirit: %v", err)
	}
}

func TestDepartedDeckhandGrantsItsEvasionToAnotherCreatureThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	deckhand := effectstest.B12Push(g, me.ID, "Departed Deckhand", "Creature — Spirit Pirate", effectstest.DepartedDeckhandOracle, 2, 2)
	giant := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, deckhand, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: giant}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.BrAttack(t, g, giant)
	effectstest.BrRefusal(t, g.DeclareBlocker(bear, giant), game.BlockReasonCantBeBlockedExceptBy)
}

// "Another": the Deckhand can't target itself.
func TestDepartedDeckhandCannotTargetItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	deckhand := effectstest.B12Push(g, me.ID, "Departed Deckhand", "Creature — Spirit Pirate", effectstest.DepartedDeckhandOracle, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
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
	tok := effects.KurukSpiritToken()
	tok.InstanceID = uuid.New()
	tok.Owner, tok.Controller = owner, owner
	return effectstest.PushBattlefieldCardWithTimestamp(g, tok)
}

func TestKurukSpiritTokenCarriesItsBlockRule(t *testing.T) {
	tok := effects.KurukSpiritToken()
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
	if game.CatalogBlockRules(game.CatalogKey(effects.TokenCard("1/1 colorless Spirit"))) != nil {
		t.Error("the plain Spirit row picked up Kuruk's rule")
	}
}

func TestKurukSpiritTokenCantBeBlockedByNonSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spirit := brKurukSpirit(t, g, me.ID)
	bear := effectstest.B12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirSpirit := effectstest.B12Creature(g, opp.ID, "Spirit", "Creature — Spirit", 1, 1)
	effectstest.BrAttack(t, g, spirit)

	offered := effectstest.BrOffered(t, g, opp.ID)
	if effectstest.BrOffers(offered, bear, spirit) || !effectstest.BrOffers(offered, theirSpirit, spirit) {
		t.Errorf("the enumerator disagrees with the token's rule: %v", offered)
	}
	refused := effectstest.BrRefusal(t, g.DeclareBlocker(bear, spirit), game.BlockReasonCantBeBlockedBy)
	if refused.Source != spirit || refused.Label != "non-Spirit creatures" {
		t.Errorf("refusal source %s label %q", refused.Source, refused.Label)
	}
	if err := g.DeclareBlocker(theirSpirit, spirit); err != nil {
		t.Fatalf("a Spirit blocks the token: %v", err)
	}
}

func TestKurukSpiritTokenCantBlockNonSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	mySpirit := effectstest.B12Creature(g, me.ID, "Spirit", "Creature — Spirit", 1, 1)
	token := brKurukSpirit(t, g, opp.ID)
	effectstest.BrAttack(t, g, bear, mySpirit)

	offered := effectstest.BrOffered(t, g, opp.ID)
	if effectstest.BrOffers(offered, token, bear) || !effectstest.BrOffers(offered, token, mySpirit) {
		t.Errorf("the enumerator disagrees with the token's rule: %v", offered)
	}
	refused := effectstest.BrRefusal(t, g.DeclareBlocker(token, bear), game.BlockReasonCantBlockAttacker)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Avatar Kuruk", "Legendary Creature — Avatar", effects.TheLegendOfKurukOracleID+"#1", 4, 3)
	effectstest.CastCatalogSpell(t, g, "Opt", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	id := effectstest.FindBattlefieldByName(g, "Spirit")
	if id == uuid.Nil {
		t.Fatal("no Spirit token")
	}
	c := effectstest.MustBattlefieldCard(t, g, id)
	if len(game.CatalogBlockRules(game.CatalogAbilityKey(*c))) != 2 {
		t.Errorf("the minted Spirit carries no block rule (key %q)", game.CatalogAbilityKey(*c))
	}
}
