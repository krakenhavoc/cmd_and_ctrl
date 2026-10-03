package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hellhole Flailer — Creature — Ogre Warrior {1}{B}{R}, 3/2:
//
//	"Unleash (You may have this creature enter with a +1/+1 counter on
//	 it. It can't block as long as it has a +1/+1 counter on it.)
//	 {2}{B}{R}, Sacrifice this creature: It deals damage equal to its
//	 power to target player or planeswalker."
//
// Unleash is the engine's keyword (ADR 0109 §10). The sacrifice is the
// cost, so by resolution the Flailer is gone: "its power" is its power
// as it last existed on the battlefield (CR 113.7a, ctx.SourcePermanent),
// an unleash counter included, and it is still the damage's source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e13c1703-1184-4a8c-9310-f11e4e673597",
		Name:            "Hellhole Flailer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUnleash},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}{R}, Sacrifice this creature: It deals damage equal to its power to target player or planeswalker.",
			Cost:    Plus(ManaCost("{2}{B}{R}"), SacrificeThis()),
			Targets: targetPlayerOrPlaneswalker(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				info, ok := ctx.SourcePermanent()
				legal := ctx.LegalTargets()
				if !ok || info.Power <= 0 || len(legal) == 0 {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: legal[0].ID, Amount: info.Power}.Apply(ctx)
			},
		}},
	})
}
