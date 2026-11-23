package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Poison-Blade Mentor — Creature — Human Assassin {1}{B}, 2/1 (Assassin's
// Creed, #5948):
//
//	"Deathtouch
//	 Whenever this creature attacks, another target Assassin you control
//	 gains deathtouch until end of turn."
//
// Captain America's "another target creature you control" attack trigger
// with an Assassin filter and a keyword grant for its body; with no other
// Assassin there is no trigger, and a target that left in response gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "60a3345c-3234-4f4b-baab-df12dad6b44b",
		Name:            "Poison-Blade Mentor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventAttack, effects.ThisAttacked,
				"Poison-Blade Mentor — another target Assassin you control gains deathtouch until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}, Label: "Poison-Blade Mentor"}.Apply(ctx)
				}),
				effects.Another(effects.TargetCreature("another target Assassin you control", effects.HasSubtype("Assassin"), effects.YouControl()))),
		},
	})
}
