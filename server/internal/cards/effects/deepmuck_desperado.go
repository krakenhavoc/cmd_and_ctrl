package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deepmuck Desperado — Creature — Homarid Mercenary {2}{U}, 2/4 (Outlaws of
// Thunder Junction, #4468):
//
//	"Whenever you commit a crime, each opponent mills three cards. This
//	 ability triggers only once each turn. (Targeting opponents, anything
//	 they control, and/or cards in their graveyards is a crime.)"
//
// The crime is the engine's once-per-announcement event (YouCommittedCrime,
// CR 700.13), as Magda, the Hoardmaster reads it, and "only once each turn"
// is the same per-turn trigger record: once this ability has been put on the
// stack this turn, later crimes that turn do not trigger it. The mill is Ruin
// Crab's, three cards for each opponent in APNAP order through the mill
// primitive, so "whenever a player mills" payoffs see each one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b75ffa62-a291-44d7-a950-b50c8394e066",
		Name:         "Deepmuck Desperado",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return YouCommittedCrime(ev, source) &&
					!b11TriggeredThisTurn(g, source.InstanceID, deepmuckDesperadoLabel)
			}, deepmuckDesperadoLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return millEachPlayerInTurn(ctx, ctx.Opponents(), constantMill(3), game.KeyedTail{})
			}),
		},
	})
}

const deepmuckDesperadoLabel = "Deepmuck Desperado — each opponent mills three cards (crime)"
