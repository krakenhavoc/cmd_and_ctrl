package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "edc74fb9-a368-4f73-8147-f318a32a3d06",
		Name:         "Corsairs of Umbar",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{2}{U}: Target Goblin, Orc, or Pirate can't be blocked this turn.",
			Cost:  ManaCost("{2}{U}"),
			Targets: TargetCreature("target Goblin, Orc, or Pirate",
				Or(OfCreatureType("Goblin"), OfCreatureType("Orc"), OfCreatureType("Pirate"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				// CR 608.2b: a target that left in response is skipped.
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return RestrictUntilEOT{
						Target:       t.ID,
						Restrictions: game.CantBeBlocked,
						Label:        "Corsairs of Umbar — can't be blocked",
					}.Apply(ctx)
				}
				return nil
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Corsairs of Umbar — amass Orcs 3", Do(Amass{Subtype: "Orc", N: 3})),
		},
	})
}
