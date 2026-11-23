package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hazoret, Godseeker — Legendary Creature — God {1}{R}, 5/3:
//
//	"Indestructible, haste
//	 Start your engines!
//	 {1}, {T}: Target creature with power 2 or less can't be blocked this
//	 turn.
//	 Hazoret can't attack or block unless you have max speed."
//
// ADR 0138 (#2122, #2711). The restriction is a layer 6 static on Hazoret
// that is on only while its controller lacks max speed (CR 702.178a, the
// speed is the player's), setting CantAttack and CantBlock, so the engine
// refuses the declaration and the enumerator does not offer it. The
// evasion ability is Access Tunnel's shape with a lower power cap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6c5d8599-75ca-4479-a02b-1b257b7aea57",
		Name:            "Hazoret, Godseeker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible", "haste", effects.StartYourEngines},
		Static: []game.StaticAbility{{
			Label: "Can't attack or block unless you have max speed.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.SelfOnly(target, g, source) && !effects.YouHaveMaxSpeed(g, effects.SpeedPlayerOf(source))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Target creature with power 2 or less can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature with power 2 or less", effects.PowerLE(2)),
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
						Label:        "Hazoret, Godseeker — can't be blocked",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
