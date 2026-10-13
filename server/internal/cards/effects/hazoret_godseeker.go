package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "6c5d8599-75ca-4479-a02b-1b257b7aea57",
		Name:            "Hazoret, Godseeker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible", "haste", StartYourEngines},
		Static: []game.StaticAbility{{
			Label: "Can't attack or block unless you have max speed.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfOnly(target, g, source) && !YouHaveMaxSpeed(g, speedPlayerOf(source))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Target creature with power 2 or less can't be blocked this turn.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("target creature with power 2 or less", PowerLE(2)),
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
						Label:        "Hazoret, Godseeker — can't be blocked",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
