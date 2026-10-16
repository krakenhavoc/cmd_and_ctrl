package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_placer_source_test.go — #3942's rules tests. Every counter
// placement names the player putting the counters at the source, so the
// "you put" replacements (Vorinclex, Monstrous Raider; Innkeeper's
// Talent at level 3) read who is putting them, not who controls the
// permanent getting them.

// placerVorinclex puts Vorinclex onto the battlefield under `owner`.
func placerVorinclex(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Vorinclex, Monstrous Raider",
		OracleID:   vorinclexPlacerOracle,
		TypeLine:   "Legendary Creature — Phyrexian Praetor",
		Power:      6, Toughness: 6,
		Owner: owner, Controller: owner,
	})
}

// castInscriptionCounters has `caster` cast Inscription of Abundance's
// first mode ("Put two +1/+1 counters on target creature") at `target`
// and resolves it.
func castInscriptionCounters(t *testing.T, g *game.Game, caster *game.Player, target uuid.UUID) {
	t.Helper()
	id := handCardFull(caster, "Inscription of Abundance", "Instant", "{1}{G}", inscriptionOfAbundanceOracle, []string{"G"})
	g.WithWriteLock(func() { _ = g.AddManaForEffect(caster.ID, uuid.Nil, "{G}{G}") })
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{
		Strict:  true,
		Modes:   []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target, Mode: 0}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// The issue's example. Your spell putting counters on an opponent's
// creature is YOUR placement: Vorinclex doubles it. The old reading,
// which fell back to the creature's controller, halved it.
func TestVorinclexDoublesYourSpellsCountersOnAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	placerVorinclex(g, me.ID)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	castInscriptionCounters(t, g, me, theirs)

	if got := countersOn(g, theirs, game.CounterPlusOne); got != 4 {
		t.Errorf("+1/+1 counters on the opponent's creature = %d, want 4 (you put two, doubled)", got)
	}
}

// And the other way round: an opponent's spell putting counters on your
// creature is THEIR placement, halved.
func TestVorinclexHalvesAnOpponentsSpellsCountersOnYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	placerVorinclex(g, me.ID)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)

	castInscriptionCounters(t, g, opp, mine)

	if got := countersOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters on your creature = %d, want 1 (an opponent put two, halved)", got)
	}
}

// CR 714.3c: "that player puts a lore counter on each Saga they
// control". Vorinclex names no effect, so it applies to the turn-based
// action: your Saga gets two, and an opponent's gets half of one,
// rounded down — none.
func TestVorinclexDoublesYourSagasLoreCounterAndStopsAnOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	placerVorinclex(g, me.ID)
	mySaga := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Saga", TypeLine: "Enchantment — Saga",
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: 1},
	})
	theirSaga := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Saga", TypeLine: "Enchantment — Saga",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterLore: 1},
	})

	aangAdvanceToMain(t, g, 0)
	if got := countersOn(g, mySaga, game.CounterLore); got != 3 {
		t.Errorf("your Saga's lore counters = %d, want 3 (1 + the turn's counter doubled)", got)
	}
	aangAdvanceToMain(t, g, 1)
	if got := countersOn(g, theirSaga, game.CounterLore); got != 1 {
		t.Errorf("the opponent's Saga's lore counters = %d, want 1 (their counter halved to none)", got)
	}
}

// CR 614.16: "if an EFFECT would put" does not reach a turn-based
// action, so Doubling Season leaves the Saga's turn counter alone. It
// did before #3942 too, by skipping the window altogether; now the
// window opens and the Season declines.
func TestDoublingSeasonLeavesASagasTurnLoreCounterAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Doubling Season", TypeLine: "Enchantment",
		OracleID: doublingSeasonOracle, Owner: me.ID, Controller: me.ID,
	})
	saga := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Saga", TypeLine: "Enchantment — Saga",
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: 1},
	})

	aangAdvanceToMain(t, g, 0)
	if got := countersOn(g, saga, game.CounterLore); got != 2 {
		t.Errorf("lore counters = %d, want 2 (the turn-based counter is not an effect's)", got)
	}
}

// A sandbox edit is the seat making it (actions.add_counter names the
// caller): an opponent's edit onto your creature is halved.
func TestVorinclexHalvesAnOpponentsSandboxEdit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	placerVorinclex(g, me.ID)
	mine := seedCreature(g, "Mine", me.ID)
	if err := g.AddCounterBy(opp.ID, mine, game.CounterPlusOne, 4); err != nil {
		t.Fatalf("AddCounterBy: %v", err)
	}
	if got := countersOn(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("counters = %d, want 2 (an opponent's edit, halved)", got)
	}
}

// Innkeeper's Talent's level 3 is Vorinclex's doubling half, and shipped
// with Vorinclex's caveat. With the placer named at the source it
// doubles the counters your spell puts on an opponent's creature and
// leaves alone the ones an opponent's spell puts on yours.
func TestInnkeepersTalentLevelThreeReadsWhoPutTheCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	talent := pushClass(g, me.ID, "Innkeeper's Talent", "Enchantment — Class", sweepFInnkeeperOracle)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	toMainForCost(t, g)
	if err := levelUpTo(t, g, me.ID, talent, 0, "{G}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if err := levelUpTo(t, g, me.ID, talent, 1, "{3}{G}"); err != nil {
		t.Fatalf("level 3: %v", err)
	}

	castInscriptionCounters(t, g, me, theirs)
	if got := countersOn(g, theirs, game.CounterPlusOne); got != 4 {
		t.Errorf("your spell's counters on the opponent's creature = %d, want 4 (doubled)", got)
	}
	castInscriptionCounters(t, g, opp, mine)
	if got := countersOn(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("an opponent's spell's counters on your creature = %d, want 2 (not yours to double)", got)
	}
}
