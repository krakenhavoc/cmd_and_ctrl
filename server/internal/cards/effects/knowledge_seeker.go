package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knowledge Seeker — Creature — Fox Spirit {1}{U}, 2/1:
//
//	"Vigilance
//	 Whenever you draw your second card each turn, put a +1/+1 counter
//	 on this creature.
//	 When this creature dies, create a Clue token."
//
// The second-card trigger is Atlantean Cavalry's: it fires on the draw
// that makes your turn's tally two, and on no other, so it triggers at
// most once a turn and never if the Seeker wasn't there for that draw
// (the rulings). The death trigger creates the set's Clue; it is not
// investigating (the ruling), so it is a plain token creation.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:        "bdc2d980-ca14-4040-a9a1-786d724dc65c",
		Name:            "Knowledge Seeker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Knowledge Seeker — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			WhenThisDies("Knowledge Seeker — create a Clue token",
				Do(CreateToken{Template: ClueToken(), N: 1})),
		},
	})
}
