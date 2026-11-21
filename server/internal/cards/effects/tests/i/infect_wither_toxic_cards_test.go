package testsi

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Plague Myr and Blighted Agent carry infect from their catalog entry,
// with no help from the importer, and Blighted Agent can't be blocked.
func TestInfectCreaturesCarryTheKeyword(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myr := effectstest.B12Push(g, me.ID, "Plague Myr", "Artifact Creature — Phyrexian Myr", effectstest.PlagueMyrOracle, 1, 1)
	agent := effectstest.B12Push(g, me.ID, "Blighted Agent", "Creature — Phyrexian Human Rogue", effectstest.BlightedAgentOracle, 1, 1)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	for _, id := range []uuid.UUID{myr, agent} {
		if !slices.Contains(effectstest.EffectiveAbilities(t, g, id), game.KeywordInfect) {
			t.Errorf("%v does not have infect", id)
		}
	}
	effectstest.DeclareAttack(t, g, opp.ID, agent)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, agent); err == nil {
		t.Error("Blighted Agent was blocked")
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if effectstest.PoisonOn(opp) != 1 || opp.Life != game.StartingLife {
		t.Errorf("opponent: %d poison, %d life; want 1 and %d", effectstest.PoisonOn(opp), opp.Life, game.StartingLife)
	}
}

// Tainted Strike grants infect until end of turn: the pumped creature's
// combat damage is poison, not life loss (ADR 0056 test plan item 8).
func TestTaintedStrikeTurnsCombatDamageIntoPoison(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Tainted Strike", "Instant", effectstest.TaintedStrikeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, bear), game.KeywordInfect) {
		t.Fatal("the Bear did not gain infect")
	}
	effectstest.AttackWith(t, g, opp.ID, bear)
	if effectstest.PoisonOn(opp) != 3 || opp.Life != game.StartingLife {
		t.Errorf("opponent: %d poison, %d life; want 3 and %d", effectstest.PoisonOn(opp), opp.Life, game.StartingLife)
	}
}

// Triumph of the Hordes: every creature you control deals poison.
func TestTriumphOfTheHordesGivesEveryAttackerInfect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Triumph of the Hordes", "Sorcery", effectstest.TriumphOfTheHordesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		abs := effectstest.EffectiveAbilities(t, g, id)
		if !slices.Contains(abs, game.KeywordInfect) || !slices.Contains(abs, "trample") {
			t.Errorf("%v abilities = %v, want trample and infect", id, abs)
		}
	}
	effectstest.AttackWith(t, g, opp.ID, a, b)
	if effectstest.PoisonOn(opp) != 7 || opp.Life != game.StartingLife {
		t.Errorf("opponent: %d poison, %d life; want 7 (3+4) and %d", effectstest.PoisonOn(opp), opp.Life, game.StartingLife)
	}
}

// Puncture Blast is wither from the stack: three -1/-1 counters on a
// creature, three life off a player. Doubling Season doubles it,
// because a spell is "an effect" (ADR 0056 Decision 5).
func TestPunctureBlastPutsCountersFromTheStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Test", 12, 12)
	effectstest.CastCatalogSpell(t, g, "Puncture Blast", "Instant", effectstest.PunctureBlastOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: big}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MinusOneOn(t, g, big); got != 3 {
		t.Errorf("-1/-1 counters = %d, want 3", got)
	}

	effectstest.PushPermanentForTest(g, opp.ID, "Doubling Season", effectstest.DoublingSeasonOracle, "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Puncture Blast", "Instant", effectstest.PunctureBlastOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: big}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MinusOneOn(t, g, big); got != 9 {
		t.Errorf("-1/-1 counters = %d, want 3 + 6 (a wither SPELL is doubled)", got)
	}

	effectstest.CastCatalogSpell(t, g, "Puncture Blast", "Instant", effectstest.PunctureBlastOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != game.StartingLife-3 || effectstest.PoisonOn(opp) != 0 {
		t.Errorf("opponent: %d life, %d poison; want %d and 0 — wither to a player is life loss",
			opp.Life, effectstest.PoisonOn(opp), game.StartingLife-3)
	}
	_ = me
}

// Doubling Season does NOT double combat-damage counters: combat
// damage is a turn-based action, not "an effect" (ADR 0056 Decision 5,
// the judge rulings it cites).
func TestInfectCombatCountersAreNotDoubledByDoublingSeason(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myr := effectstest.B12Push(g, me.ID, "Plague Myr", "Artifact Creature — Phyrexian Myr", effectstest.PlagueMyrOracle, 1, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Test", 3, 3)
	effectstest.PushPermanentForTest(g, opp.ID, "Doubling Season", effectstest.DoublingSeasonOracle, "Enchantment")
	effectstest.DeclareAttack(t, g, opp.ID, myr)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, myr); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if got := effectstest.MinusOneOn(t, g, blocker); got != 1 {
		t.Errorf("-1/-1 counters on the blocker = %d, want 1: combat damage is not an effect", got)
	}
}

// Ichor Rats gives every player a poison counter, its controller
// included.
func TestIchorRatsGivesEachPlayerAPoisonCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Ichor Rats", "Creature — Phyrexian Rat", effectstest.IchorRatsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if effectstest.PoisonOn(p) != 1 {
			t.Errorf("%s has %d poison, want 1", p.Name, effectstest.PoisonOn(p))
		}
	}
}

// Bloated Contaminator: toxic 1 lands with the combat damage, so the
// proliferate trigger finds a poison counter to add to.
func TestBloatedContaminatorProliferatesItsOwnPoison(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := effectstest.B12Push(g, me.ID, "Bloated Contaminator", "Creature — Phyrexian Beast", effectstest.BloatedContaminatorOracle, 4, 4)
	effectstest.AttackWith(t, g, opp.ID, beast)
	if opp.Life != game.StartingLife-4 || effectstest.PoisonOn(opp) != 1 {
		t.Fatalf("after damage: %d life, %d poison; want %d and 1", opp.Life, effectstest.PoisonOn(opp), game.StartingLife-4)
	}
	effectstest.SettleAnsweringProliferate(t, g)
	if effectstest.PoisonOn(opp) != 2 {
		t.Errorf("after the proliferate: %d poison, want 2", effectstest.PoisonOn(opp))
	}
}

// Karumonix's grant is cumulative: a Rat that prints toxic 1 has toxic
// 2 under it, and a plain Rat gets toxic 1 (CR 702.164b).
func TestKarumonixToxicGrantIsCumulative(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Karumonix, the Rat King", "Legendary Creature — Phyrexian Rat", effectstest.KarumonixOracle, 3, 3)
	toxicRat := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Toxic Rat", TypeLine: "Creature — Phyrexian Rat",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, Keywords: []string{"toxic 1"},
	})
	plainRat := effectstest.B12Creature(g, me.ID, "Plain Rat", "Creature — Rat", 1, 1)
	notARat := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	total := func(id uuid.UUID) int {
		n := 0
		for _, a := range effectstest.EffectiveAbilities(t, g, id) {
			if v, ok := game.ToxicValue(a); ok {
				n += v
			}
		}
		return n
	}
	if total(toxicRat) != 2 || total(plainRat) != 1 || total(notARat) != 0 {
		t.Fatalf("toxic totals: printed Rat %d, plain Rat %d, Bear %d; want 2, 1, 0",
			total(toxicRat), total(plainRat), total(notARat))
	}
	effectstest.AttackWith(t, g, opp.ID, toxicRat)
	if effectstest.PoisonOn(opp) != 2 || opp.Life != game.StartingLife-1 {
		t.Errorf("opponent: %d poison, %d life; want 2 and %d", effectstest.PoisonOn(opp), opp.Life, game.StartingLife-1)
	}
}
