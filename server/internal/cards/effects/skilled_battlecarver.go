package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skilled Battlecarver — Creature — Human Warrior {1}{R}, 2/1:
//
//	"During your turn, this creature has first strike.
//	 {1}{R}: This creature gets +1/+0 until end of turn."
//
// The keyword is a layer-6 grant on itself that holds only while its
// controller is the active player (the Zurgo Helmsmasher shape), so it
// is gone on the opponents' turns, when the creature blocks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4fc7c7a0-bfff-4833-a0d0-a38dc29e9dfb",
		Name:         "Skilled Battlecarver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && isActivePlayer(g, source.Controller)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = append(c.Abilities, "first strike")
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Skilled Battlecarver — +1/+0"}.Apply(ctx)
			},
		}},
	})
}
