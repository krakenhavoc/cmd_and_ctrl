package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hexplate Wallbreaker — Artifact — Equipment {3}{R}{R} (Phyrexia: All Will
// Be One):
//
//	"For Mirrodin! (When this Equipment enters, create a 2/2 red Rebel
//	 creature token, then attach this to it.)
//	 Equipped creature gets +2/+2.
//	 Whenever equipped creature attacks, if it's the first combat phase of
//	 the turn, untap each attacking creature. After this phase, there is an
//	 additional combat phase.
//	 Equip {3}{R}"
//
// #5479, S82. Genji Glove's trigger with Hellkite Charger's untap: the
// first-combat-phase intervening if (CR 603.4) is asked as the creature
// attacks and again as the trigger resolves, so the added combat's own
// attack does not add a third. The untap reads whoever is attacking as the
// trigger resolves; the extra combat needs no creature, so it still happens
// when the host has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7f78825-10d6-441a-ae26-51b77cc271cd",
		Name:         "Hexplate Wallbreaker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ForMirrodin("Hexplate Wallbreaker"),
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureAttacked(ev, source) && effects.IsFirstCombatPhase(g)
			}, "Hexplate Wallbreaker — untap each attacking creature; after this phase, there is an additional combat phase",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.IsFirstCombatPhase(g) {
						return nil
					}
					return effects.UntapAttackersThenExtraCombat(effects.NewContext(g, item))
				}),
		},
		Static:    []game.StaticAbility{effects.PumpAttached(2, 2)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}{R}")},
	})
}
