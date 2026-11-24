package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Biowaste Blob — Creature — Ooze {2}{G}{G}, 0/0 (Commander Legends):
//
//	"Oozes you control get +1/+1.
//	 At the beginning of your upkeep, if you control a commander, create a
//	 token that's a copy of this creature."
//
// A layer 7c anthem over Oozes you control, itself included (an Ooze by
// effective subtypes, so a changeling counts), so a lone Blob is a 1/1. The
// upkeep ability is an intervening "if" (CR 603.4) whose body is Mirror-Sigil
// Sergeant's token copy of the source; a token copy has the same abilities, so
// each copy triggers on later upkeeps too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7fa8647a-7668-4db9-a77d-87e004b160e1",
		Name:         "Biowaste Blob",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					target.HasSubtype(source.TextSubtype("Ooze"))
			}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.S82OldB083YouControlACommander,
				effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return effects.ByYou(ev, source, lki, g) && effects.ControlsACommander(g, source.Controller)
				}, "Biowaste Blob — create a token that's a copy of this creature",
					func(g *game.Game, item *game.StackItem) error {
						return effects.CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 1}.Apply(effects.NewContext(g, item))
					})),
		},
	})
}
