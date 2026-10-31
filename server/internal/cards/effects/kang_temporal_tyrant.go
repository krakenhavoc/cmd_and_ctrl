package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kang, Temporal Tyrant — Legendary Creature — Human Villain {2}{U}{B}, 3/4
// (Marvel Super Heroes, #4455):
//
//	"Whenever Kang attacks, he connives.
//	 Whenever you draw your second card each turn, each opponent loses 1
//	 life and you gain 1 life."
//
// The attack trigger is Atlantean Skirmisher's sourceConnives. The draw
// trigger reads the second draw of the turn the way Daughter of the Deep
// does (b41DrewYourSecondCardThisTurn: once a turn, only your own draws), and
// the drain is Doom Reigns Supreme's, in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b43350a8-fc4d-41ed-becd-f0277512bb5f",
		Name:         "Kang, Temporal Tyrant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Kang, Temporal Tyrant — he connives", Do(Instruction(sourceConnives))),
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Kang, Temporal Tyrant — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
		},
	})
}
