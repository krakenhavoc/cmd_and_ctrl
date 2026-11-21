package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghost of Illyria — {3}{W} Creature — Elemental Spirit 3/3 (Star Trek,
// #4973):
//
//	"Flying
//	 When this creature enters, exile up to one other target nonland
//	 permanent you control. Return that card to the battlefield under its
//	 owner's control at the beginning of the next end step."
//
// Elrond, Moon-Reader's blink with one target: the exile and a delayed
// trigger at the next end step over the card that reached exile
// (exileTargetsThenScheduleReturn). It comes back as a new object under
// its owner's control (CR 400.7); a token exiled this way ceases to exist
// and does not return (CR 111.7). With no target chosen, nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e705b055-14b3-4de0-b4a7-39626839dea8",
		Name:            "Ghost of Illyria",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Ghost of Illyria — exile up to one other target nonland permanent you control",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTargetsThenScheduleReturn(effects.NewContext(g, item), "Ghost of Illyria — return the exiled card")
				}),
				effects.Another(effects.TargetPermanent("up to one other target nonland permanent you control", effects.Nonland(), effects.YouControl()).WithCount(0, 1))),
		},
	})
}
