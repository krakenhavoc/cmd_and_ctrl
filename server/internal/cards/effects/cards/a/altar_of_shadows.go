package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Altar of Shadows — Artifact {7} (Mirrodin, #5606):
//
//	"At the beginning of your first main phase, add {B} for each charge
//	 counter on this artifact.
//	 {7}, {T}: Destroy target creature. Then put a charge counter on this
//	 artifact."
//
// Gardenize's first-main-phase mana (AtYourPrecombatMain, counters read
// as the trigger resolves, the mana empties with the phase, CR 106.4)
// beside a destroy ability. The charge counter goes on after the
// destruction, whether or not the creature was destroyed (an
// indestructible one, a regeneration shield); a creature that is no
// longer a legal target leaves the ability with nothing to do, counter
// included (CR 608.2b). An Altar that left the battlefield in response
// gets no counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b34a324-d939-453a-97e3-a21c584a0c74",
		Name:         "Altar of Shadows",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourPrecombatMain("Altar of Shadows — add {B} for each charge counter on this artifact",
				func(g *game.Game, item *game.StackItem) error {
					n := effects.ChargeCountersOn(g, item.Controller, item.SourceCardID)
					if n <= 0 {
						return nil
					}
					return effects.AddMana{Player: item.Controller, Amount: effects.ManaN(n, "B")}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}, {T}: Destroy target creature. Then put a charge counter on this artifact.",
			Cost:    effects.Plus(effects.ManaCost("{7}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.DestroyFirstLegalCardTarget(g, item); err != nil {
					return err
				}
				return effects.B33PutChargeCounterOnSelf(g, item)
			},
		}},
	})
}
