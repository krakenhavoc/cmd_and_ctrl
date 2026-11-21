package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Free Borg Revolutionaries — {3}{W}{W} Artifact Creature — Borg Rebel
// 4/5 (Star Trek, #4973):
//
//	"When this creature enters, creatures you control get +1/+1 until
//	 end of turn.
//	 Basic landcycling {2} ({2}, Discard this card: Search your library
//	 for a basic land card, reveal it, put it into your hand, then
//	 shuffle.)"
//
// Dwarven Lightsmith's enters trigger: the creatures you control as it
// resolves get +1/+1, the Revolutionaries included, and a creature that
// enters later does not (CR 611.2c). Basic landcycling is the cycling
// family's BasicLandcycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "28914782-92f4-407f-beb6-5e290a49000c",
		Name:         "Free Borg Revolutionaries",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Free Borg Revolutionaries — creatures you control get +1/+1 until end of turn",
				effects.Do(effects.BoostUntilEOT{
					Match:     effects.And(effects.Creature(), effects.YouControl()),
					Power:     1,
					Toughness: 1,
					Label:     "Free Borg Revolutionaries — +1/+1",
				})),
		},
		Activated: []effects.ActivatedAbility{effects.BasicLandcycling("{2}")},
	})
}
