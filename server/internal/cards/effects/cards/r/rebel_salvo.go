package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rebel Salvo — Instant {2}{R} (Phyrexia: All Will Be One):
//
//	"Affinity for Equipment (This spell costs {1} less to cast for each
//	 Equipment you control.)
//	 Rebel Salvo deals 5 damage to target creature or planeswalker. That
//	 permanent loses indestructible until end of turn."
//
// #5479, S82. Broadside Barrage's 5 damage, Claws Out's affinity over
// Equipment, and Burn from Within's loss of indestructible taken as the
// damage's continuation: it runs whatever the damage did (the card does not
// say "dealt damage this way"), after the damage is marked, so the
// state-based check that follows the resolution destroys a creature that
// already has lethal damage (CR 702.12b, 704.5g).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a35b7fd3-fdef-4414-b720-978f7ce70bff",
		Name:         "Rebel Salvo",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 5))},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for Equipment", effects.HasSubtype("Equipment")),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return effects.DealDamageThen(ctx, t.ID, 5, effects.ItemRecipient(item, effects.S82OneB14LosesIndestructible))
			}
			return nil
		},
	})
}
