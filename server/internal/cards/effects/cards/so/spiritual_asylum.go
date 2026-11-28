package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spiritual Asylum — Enchantment {2}{W}{W} (Nemesis, S82 OLD, #5618):
//
//	"Creatures and lands you control have shroud.
//	 When a creature you control attacks, sacrifice this enchantment."
//
// One layer-6 shroud grant over the controller's creatures and lands (read live,
// so an animated land is both), and a trigger on every attacker its controller
// declares; a second attacker's trigger finds the Asylum gone and does nothing.
//
// No simplification.
func init() {
	shroud := effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && (target.IsCreature() || target.IsLand())
	}, "shroud")
	shroud.Label = "Creatures and lands you control have shroud."
	effects.Register(effects.Spec{
		OracleID:     "91924536-7a2b-44ec-9835-2efa402c83f9",
		Name:         "Spiritual Asylum",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{shroud},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Spiritual Asylum — sacrifice this enchantment", effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
