package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "8491de1e-dc37-4a89-b4f7-56a2c668f0fc",
		Name:         "Out of Time",
		Completeness: effects.CompletenessFull,
		Triggered: append(effects.Cmdr2B17VanishingTriggers(),
			effects.WhenThisEnters("Out of Time — untap all creatures, then they phase out until this enchantment leaves the battlefield; put a time counter on this enchantment for each",
				outOfTimeEnters)),
	})
}

func outOfTimeEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range effects.MatchingBattlefield(ctx, effects.Creature()) {
		ids = append(ids, c.InstanceID)
	}
	if err := effects.UntapEach(ctx, ids); err != nil {
		return err
	}
	if err := (effects.PhaseOutUntilLeaves{Targets: ids, Until: item.SourceCardID}).Apply(ctx); err != nil {
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
	if n == 0 || !effects.OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterTime, N: n}.Apply(ctx)
}
