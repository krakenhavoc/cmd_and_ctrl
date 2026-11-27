package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Moonring Island — Land — Island (Shadowmoor, #5591):
//
//	"({T}: Add {U}.)
//	 This land enters tapped.
//	 {U}, {T}: Look at the top card of target player's library. Activate
//	 only if you control two or more blue permanents."
//
// Leechridden Swamp's shape with Mishra's Bauble's look: the looker is the
// controller, only they learn the card (CR 701.20), and it goes back on top
// (nothing says to move it). The {U} is the Island type's intrinsic
// ability; the land is colourless so it never counts toward its own two.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf620c66-7db1-4db8-ae56-ee4bc2f77d74",
		Name:         "Moonring Island",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label:     "{U}, {T}: Look at the top card of target player's library. Activate only if you control two or more blue permanents.",
			Cost:      effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Condition: effects.ControlsAtLeast(2, effects.MatchColor("U")),
			Targets:   effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				var victim uuid.UUID
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						victim = t.ID
					}
				}
				if victim == uuid.Nil {
					return nil
				}
				return effects.LookAtLibraryThenPlace{
					Owner:     victim,
					N:         1,
					Placement: game.LibraryPlaceTop,
					Label:     "Moonring Island — put it back on top",
				}.Apply(ctx)
			},
		}},
	})
}
