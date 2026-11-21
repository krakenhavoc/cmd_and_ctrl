package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corsairs of Umbar — Creature — Human Pirate {3}{U}, 3/3:
//
//	"{2}{U}: Target Goblin, Orc, or Pirate can't be blocked this turn.
//	 Whenever this creature deals combat damage to a player, amass
//	 Orcs 3."
//
// The activation is Access Tunnel's can't-be-blocked restriction behind
// a tribe-OR target clause; the Corsairs are a Pirate and can aim it at
// themselves. The combat-damage trigger is the Bowmasters' amass,
// firing once per damage event to a player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "edc74fb9-a368-4f73-8147-f318a32a3d06",
		Name:         "Corsairs of Umbar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Target Goblin, Orc, or Pirate can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{2}{U}"),
			Targets: effects.TargetCreature("target Goblin, Orc, or Pirate",
				effects.Or(effects.OfCreatureType("Goblin"), effects.OfCreatureType("Orc"), effects.OfCreatureType("Pirate"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				// CR 608.2b: a target that left in response is skipped.
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return effects.RestrictUntilEOT{
						Target:       t.ID,
						Restrictions: game.CantBeBlocked,
						Label:        "Corsairs of Umbar — can't be blocked",
					}.Apply(ctx)
				}
				return nil
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Corsairs of Umbar — amass Orcs 3", effects.Do(effects.Amass{Subtype: "Orc", N: 3})),
		},
	})
}
