package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Data, Professor of Humanities — {3}{U} Legendary Artifact Creature —
// Android Scientist 3/4 (Star Trek Commander, #4974):
//
//	"Seek out — Whenever you draw your second card each turn, create two
//	 1/1 blue Faerie creature tokens with flying."
//
// "Seek out" is an ability word with no rules meaning (CR 207.2c). The
// trigger is Atlantean Cavalry's b41DrewYourSecondCardThisTurn: your
// draw that makes two this turn, on anyone's turn, once a turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "25ecd09e-d8bf-4929-96e2-e5132ea4f718",
		Name:         "Data, Professor of Humanities",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Data, Professor of Humanities — create two 1/1 blue Faerie tokens with flying",
				Do(CreateToken{Template: TokenCard(trekB46FaerieToken), N: 2})),
		},
	})
}
