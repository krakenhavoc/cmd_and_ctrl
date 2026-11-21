package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infernal Phantom — Creature — Spirit {3}{R}, 2/3:
//
//	"Eerie — Whenever an enchantment you control enters and whenever
//	 you fully unlock a Room, this creature gets +2/+0 until end of
//	 turn.
//	 When this creature dies, it deals damage equal to its power to any
//	 target."
//
// Eerie is one ability with two conditions (Eerie, rooms.go), and the
// pump is thisCreatureUntilEOT. The dies trigger targets any target as
// it goes on the stack, and reads the power the Phantom had as it died
// (CR 603.10, b13LastKnownPower: layers and counters both count, so a
// turn's worth of eerie pumps is in the damage). The damage is dealt by
// that last-known object, not by the card now in the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a8225b2e-62bf-45b7-b573-e3ce301a1eab",
		Name:         "Infernal Phantom",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Eerie("Infernal Phantom — +2/+0 until end of turn (eerie)",
				effects.ThisCreatureUntilEOT("Infernal Phantom — +2/+0 until end of turn", 2, 0)),
			effects.EerieDiesWithPower("Infernal Phantom — it deals damage equal to its power to any target", effects.TargetAny(),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ref, ok := ctx.SourceRef()
					if !ok {
						return nil
					}
					for _, t := range ctx.LegalTargets() {
						return effects.DealDamage{SourceObject: &ref, Target: t.ID, Amount: item.Params.Amount}.Apply(ctx)
					}
					return nil
				}),
		},
	})
}
