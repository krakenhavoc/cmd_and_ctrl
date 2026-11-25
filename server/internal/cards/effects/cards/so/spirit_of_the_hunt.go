package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spirit of the Hunt — Creature — Wolf Spirit {1}{G}{G}, 3/3 (Eldritch Moon,
// S82 OLD, #5541):
//
//	"Flash
//	 When this creature enters, each other creature you control that's a
//	 Wolf or a Werewolf gets +0/+3 until end of turn."
//
// Bess's "each other creature you control" boost, matching the layered
// subtypes (a changeling counts) as the trigger resolves (CR 611.2c: only
// the creatures there then are affected).
//
// No simplification.
func init() {
	const label = "Spirit of the Hunt — each other creature you control that's a Wolf or a Werewolf gets +0/+3 until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "3a81a3a6-4940-4d69-94c1-d431023b275c",
		Name:            "Spirit of the Hunt",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl(), effects.NotSelf(item.SourceCardID),
						effects.Or(effects.HasSubtype("Wolf"), effects.HasSubtype("Werewolf"))),
					Toughness: 3,
					Label:     label,
				}.Apply(ctx)
			}),
		},
	})
}
