package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// emitCombatHits fakes one combat damage step in which each creature in
// `hits` deals 2 combat damage to the matching victim, all in one batch.
func emitCombatHits(g *game.Game, creatures, victims []uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range creatures {
			g.EmitEvent(game.Event{
				Kind: game.EventDealDamage, Source: creatures[i], Target: victims[i],
				Amount: 2, Combat: true,
			})
		}
	})
}

// Hordewing Skaab: three Zombies connecting with two opponents in one
// damage step are ONE trigger that counts the opponents, not the Zombies.
func TestHordewingSkaabCountsDistinctOpponentsDealtDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	skaab := b41CatalogCreature(g, me.ID, "Hordewing Skaab", "Creature — Zombie Horror", b41HordewingSkaabOracle, 3, 3)
	z1 := b12Creature(g, me.ID, "Zombie One", "Creature — Zombie", 2, 2)
	z2 := b12Creature(g, me.ID, "Zombie Two", "Creature — Zombie", 2, 2)
	z3 := b12Creature(g, me.ID, "Zombie Three", "Creature — Zombie", 2, 2)

	emitCombatHits(g, []uuid.UUID{z1, z2, z3}, []uuid.UUID{opp1.ID, opp1.ID, opp2.ID})
	g.RunStateChecksForTest()
	answerLatestTriggerPrompt(t, g, me.ID, true)
	if got := triggerAmountsFrom(g, skaab); len(got) != 1 || got[0] != 2 {
		t.Fatalf("one trigger counting the two opponents hit, want [2], got %v", got)
	}
	// Resolving it draws two and then owes a discard of two.
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2 (one per opponent dealt damage)", got)
	}
	if got := discardOwed(g, me.ID); got != 2 {
		t.Errorf("owes a discard of %d, want 2", got)
	}
}

// One opponent hit by a whole alpha strike is still a count of one.
func TestHordewingSkaabCountsOneForOneOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skaab := b41CatalogCreature(g, me.ID, "Hordewing Skaab", "Creature — Zombie Horror", b41HordewingSkaabOracle, 3, 3)
	z1 := b12Creature(g, me.ID, "Zombie One", "Creature — Zombie", 2, 2)
	z2 := b12Creature(g, me.ID, "Zombie Two", "Creature — Zombie", 2, 2)

	emitCombatHits(g, []uuid.UUID{skaab, z1, z2}, []uuid.UUID{opp.ID, opp.ID, opp.ID})
	g.RunStateChecksForTest()
	answerLatestTriggerPrompt(t, g, me.ID, true)
	if got := triggerAmountsFrom(g, skaab); len(got) != 1 || got[0] != 1 {
		t.Fatalf("one trigger counting one opponent, want [1], got %v", got)
	}
}

// A non-Zombie's hit is not counted.
func TestHordewingSkaabIgnoresNonZombieDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	skaab := b41CatalogCreature(g, me.ID, "Hordewing Skaab", "Creature — Zombie Horror", b41HordewingSkaabOracle, 3, 3)
	zombie := b12Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	emitCombatHits(g, []uuid.UUID{zombie, bear}, []uuid.UUID{opp1.ID, opp2.ID})
	g.RunStateChecksForTest()
	answerLatestTriggerPrompt(t, g, me.ID, true)
	if got := triggerAmountsFrom(g, skaab); len(got) != 1 || got[0] != 1 {
		t.Fatalf("only the Zombie's opponent counts, want [1], got %v", got)
	}
}
