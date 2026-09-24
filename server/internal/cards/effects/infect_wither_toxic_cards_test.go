package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// infect_wither_toxic_cards_test.go plays ADR 0056's first card wave
// (#748) through real combat and real casts: the damage tail turning
// infect, wither and toxic damage into counters is exercised through
// the cards that print or grant the keywords.

const (
	plagueMyrOracle           = "2f328e05-5edf-4b21-9c2a-50dcf1e7b3ec"
	blightedAgentOracle       = "e48ea9ea-64bc-4f1c-a424-592d48569244"
	ichorRatsOracle           = "f8148664-49c0-421f-93a5-cd59e4e9ea36"
	punctureBlastOracle       = "e296581d-01ac-43bf-898c-2edb4c81bcbe"
	bloatedContaminatorOracle = "090018e0-4dcb-4b3c-b4e0-7ba62de0484d"
	taintedStrikeOracle       = "95a53dd7-76dc-46f1-8833-fafd02ba49c4"
	triumphOfTheHordesOracle  = "3ded0c0c-40ce-4d14-a9a6-b023bc19ee0e"
	karumonixOracle           = "c017f54c-e4c0-411e-b8c1-eb20b1b86c56"
)

func poisonOn(p *game.Player) int { return p.Counters[game.CounterPoison] }

func minusOneOn(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	n := -1
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.Counters[game.CounterMinusOne]
			}
		}
	})
	return n
}

// Plague Myr and Blighted Agent carry infect from their catalog entry,
// with no help from the importer, and Blighted Agent can't be blocked.
func TestInfectCreaturesCarryTheKeyword(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myr := b12Push(g, me.ID, "Plague Myr", "Artifact Creature — Phyrexian Myr", plagueMyrOracle, 1, 1)
	agent := b12Push(g, me.ID, "Blighted Agent", "Creature — Phyrexian Human Rogue", blightedAgentOracle, 1, 1)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	for _, id := range []uuid.UUID{myr, agent} {
		if !slices.Contains(effectiveAbilities(t, g, id), game.KeywordInfect) {
			t.Errorf("%v does not have infect", id)
		}
	}
	declareAttack(t, g, opp.ID, agent)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, agent); err == nil {
		t.Error("Blighted Agent was blocked")
	}
	advanceTo(t, g, game.StepCombatDamage)
	if poisonOn(opp) != 1 || opp.Life != game.StartingLife {
		t.Errorf("opponent: %d poison, %d life; want 1 and %d", poisonOn(opp), opp.Life, game.StartingLife)
	}
}

// Tainted Strike grants infect until end of turn: the pumped creature's
// combat damage is poison, not life loss (ADR 0056 test plan item 8).
func TestTaintedStrikeTurnsCombatDamageIntoPoison(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Tainted Strike", "Instant", taintedStrikeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !slices.Contains(effectiveAbilities(t, g, bear), game.KeywordInfect) {
		t.Fatal("the Bear did not gain infect")
	}
	attackWith(t, g, opp.ID, bear)
	if poisonOn(opp) != 3 || opp.Life != game.StartingLife {
		t.Errorf("opponent: %d poison, %d life; want 3 and %d", poisonOn(opp), opp.Life, game.StartingLife)
	}
}

// Triumph of the Hordes: every creature you control deals poison.
func TestTriumphOfTheHordesGivesEveryAttackerInfect(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 3, 3)
	castCatalogSpell(t, g, "Triumph of the Hordes", "Sorcery", triumphOfTheHordesOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		abs := effectiveAbilities(t, g, id)
		if !slices.Contains(abs, game.KeywordInfect) || !slices.Contains(abs, "trample") {
			t.Errorf("%v abilities = %v, want trample and infect", id, abs)
		}
	}
	attackWith(t, g, opp.ID, a, b)
	if poisonOn(opp) != 7 || opp.Life != game.StartingLife {
		t.Errorf("opponent: %d poison, %d life; want 7 (3+4) and %d", poisonOn(opp), opp.Life, game.StartingLife)
	}
}

// Puncture Blast is wither from the stack: three -1/-1 counters on a
// creature, three life off a player. Doubling Season doubles it,
// because a spell is "an effect" (ADR 0056 Decision 5).
func TestPunctureBlastPutsCountersFromTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b12Creature(g, opp.ID, "Big", "Creature — Test", 12, 12)
	castCatalogSpell(t, g, "Puncture Blast", "Instant", punctureBlastOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: big}})
	passPriorityAroundTable(t, g)
	if got := minusOneOn(t, g, big); got != 3 {
		t.Errorf("-1/-1 counters = %d, want 3", got)
	}

	pushPermanentForTest(g, opp.ID, "Doubling Season", doublingSeasonOracle, "Enchantment")
	castCatalogSpell(t, g, "Puncture Blast", "Instant", punctureBlastOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: big}})
	passPriorityAroundTable(t, g)
	if got := minusOneOn(t, g, big); got != 9 {
		t.Errorf("-1/-1 counters = %d, want 3 + 6 (a wither SPELL is doubled)", got)
	}

	castCatalogSpell(t, g, "Puncture Blast", "Instant", punctureBlastOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != game.StartingLife-3 || poisonOn(opp) != 0 {
		t.Errorf("opponent: %d life, %d poison; want %d and 0 — wither to a player is life loss",
			opp.Life, poisonOn(opp), game.StartingLife-3)
	}
	_ = me
}

// Doubling Season does NOT double combat-damage counters: combat
// damage is a turn-based action, not "an effect" (ADR 0056 Decision 5,
// the judge rulings it cites).
func TestInfectCombatCountersAreNotDoubledByDoublingSeason(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myr := b12Push(g, me.ID, "Plague Myr", "Artifact Creature — Phyrexian Myr", plagueMyrOracle, 1, 1)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Test", 3, 3)
	pushPermanentForTest(g, opp.ID, "Doubling Season", doublingSeasonOracle, "Enchantment")
	declareAttack(t, g, opp.ID, myr)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, myr); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	if got := minusOneOn(t, g, blocker); got != 1 {
		t.Errorf("-1/-1 counters on the blocker = %d, want 1: combat damage is not an effect", got)
	}
}

// Ichor Rats gives every player a poison counter, its controller
// included.
func TestIchorRatsGivesEachPlayerAPoisonCounter(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Ichor Rats", "Creature — Phyrexian Rat", ichorRatsOracle, nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if poisonOn(p) != 1 {
			t.Errorf("%s has %d poison, want 1", p.Name, poisonOn(p))
		}
	}
}

// Bloated Contaminator: toxic 1 lands with the combat damage, so the
// proliferate trigger finds a poison counter to add to.
func TestBloatedContaminatorProliferatesItsOwnPoison(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := b12Push(g, me.ID, "Bloated Contaminator", "Creature — Phyrexian Beast", bloatedContaminatorOracle, 4, 4)
	attackWith(t, g, opp.ID, beast)
	if opp.Life != game.StartingLife-4 || poisonOn(opp) != 1 {
		t.Fatalf("after damage: %d life, %d poison; want %d and 1", opp.Life, poisonOn(opp), game.StartingLife-4)
	}
	passPriorityAroundTable(t, g)
	if poisonOn(opp) != 2 {
		t.Errorf("after the proliferate: %d poison, want 2", poisonOn(opp))
	}
}

// Karumonix's grant is cumulative: a Rat that prints toxic 1 has toxic
// 2 under it, and a plain Rat gets toxic 1 (CR 702.164b).
func TestKarumonixToxicGrantIsCumulative(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Karumonix, the Rat King", "Legendary Creature — Phyrexian Rat", karumonixOracle, 3, 3)
	toxicRat := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Toxic Rat", TypeLine: "Creature — Phyrexian Rat",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, Keywords: []string{"toxic 1"},
	})
	plainRat := b12Creature(g, me.ID, "Plain Rat", "Creature — Rat", 1, 1)
	notARat := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	total := func(id uuid.UUID) int {
		n := 0
		for _, a := range effectiveAbilities(t, g, id) {
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
	attackWith(t, g, opp.ID, toxicRat)
	if poisonOn(opp) != 2 || opp.Life != game.StartingLife-1 {
		t.Errorf("opponent: %d poison, %d life; want 2 and %d", poisonOn(opp), opp.Life, game.StartingLife-1)
	}
}
