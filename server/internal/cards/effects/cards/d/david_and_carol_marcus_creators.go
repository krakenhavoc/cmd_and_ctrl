package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// David and Carol Marcus, Creators — Legendary Creature — Human
// Scientist {2}{G}, 1/3 (Star Trek, #4973):
//
//	"When David and Carol Marcus enter, create a tapped Planet land token
//	 named New Planet with "{T}: Add one mana of any color.""
//
// The token is the printed New Planet (ttrk 27), Genesis Effect's, put
// onto the battlefield tapped. A land put onto the battlefield is not
// played, so it does not use a land drop.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15885e58-76a9-49fa-ae67-a0482aacd81c",
		Name:         "David and Carol Marcus, Creators",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("David and Carol Marcus — create a tapped New Planet",
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.TrekB16NewPlanetToken()).EntersTapped(), N: 1})),
		},
	})
}
