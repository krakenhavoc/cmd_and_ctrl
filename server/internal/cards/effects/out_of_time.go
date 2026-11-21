package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Out of Time — Enchantment {1}{W}{W} (Doctor Who):
//
//	"When this enchantment enters, untap all creatures, then those
//	 creatures phase out until this enchantment leaves the battlefield. Put
//	 a time counter on this enchantment for each creature that phased out
//	 this way.
//	 Vanishing (At the beginning of your upkeep, remove a time counter from
//	 this enchantment. When the last is removed, sacrifice it.)"
//
// Vanishing with no number enters with no counters: they come from the
// enter trigger, which untaps every creature at the table, phases them out
// until this leaves (PhaseOutUntilLeaves; the Auras and Equipment on them go
// with them, CR 702.26g) and counts the creatures that actually phased out.
// Vanishing's upkeep removal and last-counter sacrifice are
// cmdr2B17Vanishing*; when the last counter is gone the enchantment leaves
// and every creature phases back in at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8491de1e-dc37-4a89-b4f7-56a2c668f0fc",
		Name:         "Out of Time",
		Completeness: CompletenessFull,
		Triggered: append(cmdr2B17VanishingTriggers(),
			WhenThisEnters("Out of Time — untap all creatures, then they phase out until this enchantment leaves the battlefield; put a time counter on this enchantment for each",
				outOfTimeEnters)),
	})
}

func outOfTimeEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, Creature()) {
		ids = append(ids, c.InstanceID)
	}
	if err := untapEach(ctx, ids); err != nil {
		return err
	}
	if err := (PhaseOutUntilLeaves{Targets: ids, Until: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	out := map[uuid.UUID]bool{}
	for _, c := range g.PhasedOutCardsForEffect() {
		out[c.InstanceID] = true
	}
	n := 0
	for _, id := range ids {
		if out[id] {
			n++
		}
	}
	if n == 0 || !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterTime, N: n}.Apply(ctx)
}
