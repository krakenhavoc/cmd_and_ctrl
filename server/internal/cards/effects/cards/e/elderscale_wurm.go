package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elderscale Wurm — Creature — Wurm {4}{G}{G}{G}, 7/7 (Magic 2013, S82 OLD, #5568):
//
//	"Trample
//	 When this creature enters, if your life total is less than 7, your life
//	 total becomes 7.
//	 As long as you have 7 or more life, damage that would reduce your life
//	 total to less than 7 reduces it to 7 instead."
//
// The enters trigger is Resolute Archangel's life change behind an intervening
// "if" (CR 603.4, CR 119.5). The static is a replacement on the life window
// (Serra the Benevolent's emblem shape, CR 614, CR 120.3a): it reads the total
// as the damage is dealt, applies to the damage's life only (lifelink and
// commander damage still see all of it) and never to a loss that is not damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1e81454a-ce30-4e41-9d09-3256ae30efc2",
		Name:            "Elderscale Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("your life total is less than 7", effects.S82OldB429LifeBelowSeven,
				effects.WhenThisEnters("Elderscale Wurm — your life total becomes 7",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						return effects.B31LifeBecomes(ctx, ctx.Controller(), 7)
					})),
		},
		Replacements: []game.ReplacementEffect{
			effects.S82OldB429SevenLifeFloor("Elderscale Wurm: damage can't reduce your life total below 7 while you have 7 or more"),
		},
	})
}
