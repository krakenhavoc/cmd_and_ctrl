package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fall of Cair Andros — Enchantment {2}{R}:
//
//	"Whenever a creature an opponent controls is dealt excess
//	 noncombat damage, amass Orcs X, where X is that excess damage.
//	 {7}{R}: This enchantment deals 7 damage to target creature."
//
// Magmatic Galleon's excess-damage reading (excess_damage.go) turned
// into a number: the trigger is harvested as the damage lands, so the
// excess is computed there and carried on the stack item's Params
// (Kaervek's shape), and the effect amasses that many. One trigger per
// creature dealt excess damage, not one per batch, because the text
// names each creature ("a creature ... is dealt"). The activated
// ability is a plain targeted 7 damage.
//
// Deathtouch is read off the damaging source (excess_damage.go): one
// point is lethal from it, so everything past the first is excess.
func init() {
	const label = "Fall of Cair Andros — amass Orcs X, where X is that excess damage"
	effects.Register(effects.Spec{
		OracleID:     "33c0a8c0-d1d9-4b03-869d-65dab8ace3df",
		Name:         "Fall of Cair Andros",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.ExcessNoncombatDamageToOpponentCreature(ev, source, g) > 0
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Amount = effects.ExcessNoncombatDamageToOpponentCreature(ev, source, g)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Amass{Subtype: "Orc", N: item.Params.Amount}.Apply(effects.NewContext(g, item))
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}{R}: This enchantment deals 7 damage to target creature.",
			Cost:    effects.ManaCost("{7}{R}"),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 7)),
			Effect:  effects.DealDamageToTheTarget(7),
		}},
	})
}
