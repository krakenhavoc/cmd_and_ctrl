package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Genji Glove — Artifact — Equipment {5}:
//
//	"Equipped creature has double strike.
//	 Whenever equipped creature attacks, if it's the first combat phase of
//	 the turn, untap it. After this phase, there is an additional combat
//	 phase.
//	 Equip {3}"
//
// "If it's the first combat phase of the turn" is an intervening if
// (CR 603.4): asked as the creature attacks and again as the trigger
// resolves. The untap reads the equipped creature as the ability resolves
// (Sting's shape); the extra combat is the shared
// ExtraCombatAfterThisPhase and needs no creature, so it still happens
// when the host has left.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8242cd78-6675-43b5-8924-0a88c3aa6f76",
		Name:         "Genji Glove",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GrantToAttached("double strike")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureAttacked(ev, source) && effects.IsFirstCombatPhase(g)
			}, "Genji Glove — untap equipped creature; after this phase, there is an additional combat phase",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.IsFirstCombatPhase(g) {
						return nil
					}
					if err := effects.UntapAttachedHost(g, item); err != nil {
						return err
					}
					return effects.ExtraCombatAfterThisPhase().Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
