package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Otter-Penguin — Creature — Otter Bird {1}{U}, 2/1 (Avatar: The Last
// Airbender, #4459):
//
//	"Whenever you draw your second card each turn, this creature gets
//	 +1/+2 until end of turn and can't be blocked this turn."
//
// The second-card trigger is Knights of Dol Amroth's
// (b41DrewYourSecondCardThisTurn): it fires on the draw that makes the
// turn's tally two and on no other, so at most once a turn. Cards put
// into your hand without the word "draw" are not draws (the ruling).
// Once the creature is blocked, the restriction does not unblock it
// (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e2171c48-a816-41b2-aa55-e0ba2503fd6a",
		Name:         "Otter-Penguin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Otter-Penguin — gets +1/+2 until end of turn and can't be blocked this turn",
				avatarB6BoostAndUnblockableSelf("Otter-Penguin", 1, 2)),
		},
	})
}
