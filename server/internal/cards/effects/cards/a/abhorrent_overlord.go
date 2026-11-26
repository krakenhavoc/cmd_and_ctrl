package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abhorrent Overlord — Creature — Demon {5}{B}{B}, 6/6 (Theros, S82 old
// sets):
//
//	"Flying
//	 When this creature enters, create a number of 1/1 black Harpy creature
//	 tokens with flying equal to your devotion to black. (Each {B} in the
//	 mana costs of permanents you control counts toward your devotion to
//	 black.)
//	 At the beginning of your upkeep, sacrifice a creature."
//
// The devotion is read as the trigger resolves, with the Overlord itself on
// the battlefield (CR 700.5), and a count of zero creates nothing. The
// upkeep sacrifice is the controller's own choice of any creature they
// control, the Overlord and the Harpies included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "13f63240-a6bb-4333-a3dd-818725c6fb73",
		Name:            "Abhorrent Overlord",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Abhorrent Overlord — create a number of 1/1 black Harpy creature tokens with flying equal to your devotion to black",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := effects.S82OldB378DevotionOf(ctx, "B")
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("1/1 black Harpy with flying"), N: n}.Apply(ctx)
				}),
			effects.AtYourUpkeep("Abhorrent Overlord — sacrifice a creature",
				func(g *game.Game, item *game.StackItem) error {
					g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller,
						effects.SacrificeSpec("a creature", effects.Creature()), "Abhorrent Overlord — sacrifice a creature")
					return nil
				}),
		},
	})
}
