package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "15885e58-76a9-49fa-ae67-a0482aacd81c",
		Name:         "David and Carol Marcus, Creators",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("David and Carol Marcus — create a tapped New Planet",
				Do(CreateTokenAdvanced{Spec: Token(trekB16NewPlanetToken()).EntersTapped(), N: 1})),
		},
	})
}
