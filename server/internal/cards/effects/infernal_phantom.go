package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a8225b2e-62bf-45b7-b573-e3ce301a1eab",
		Name:         "Infernal Phantom",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Eerie("Infernal Phantom — +2/+0 until end of turn (eerie)",
				thisCreatureUntilEOT("Infernal Phantom — +2/+0 until end of turn", 2, 0)),
			eerieDiesWithPower("Infernal Phantom — it deals damage equal to its power to any target", TargetAny(),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ref, ok := ctx.SourceRef()
					if !ok {
						return nil
					}
					for _, t := range ctx.LegalTargets() {
						return DealDamage{SourceObject: &ref, Target: t.ID, Amount: item.Params.Amount}.Apply(ctx)
					}
					return nil
				}),
		},
	})
}
