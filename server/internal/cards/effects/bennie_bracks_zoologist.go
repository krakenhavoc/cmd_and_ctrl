package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bennie Bracks, Zoologist — Legendary Creature — Elf Druid {3}{W},
// 3/2 (EDHREC rank 1830):
//
//	"Convoke (Your creatures can help cast this spell. Each creature
//	 you tap while casting this spell pays for {1} or one mana of
//	 that creature's color.)
//	 At the beginning of each end step, if you created a token this
//	 turn, draw a card."
//
// The token deck's card-a-turn. Convoke is the real S22 tap cost;
// the draw is "at the beginning of EACH end step" — any player's,
// b15EndStepBegan — with the intervening if read off the event log
// (b16CreatedATokenThisTurn: an EventTokenCreated by the controller
// since the turn's upkeep began). A token created in response to the
// trigger is too late, as printed: the condition is checked when the
// trigger would go on the stack, and a trigger whose condition
// fails is not put on the stack at all.
//
// Intervening-if caveat, shared with every such card in the catalog:
// the condition is not re-checked on resolution — for this one it
// cannot become false (a token created stays created).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17bd7ef7-8b4b-4a2d-a667-c751e10e2a47",
		Name:         "Bennie Bracks, Zoologist",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b15EndStepBegan(ev) && b16CreatedATokenThisTurn(g, source.Controller)
			}, "Bennie Bracks, Zoologist — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
