package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chakram Retriever — Creature — Elemental Dog {4}{U}, 2/4 (Battlebond, #5529):
//
//	"Partner with Chakram Slinger (When this creature enters, target player may put Chakram Slinger into their hand from their library, then shuffle.)"
//	"Whenever you cast a spell during your turn, untap target creature."
//
// The cast trigger fires only while it is your turn; the target is chosen as the trigger goes on the stack and re-checked on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ad2ecd5-2d50-4a54-a672-335d7e636a5a",
		Name:         "Chakram Retriever",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Chakram Retriever", "Chakram Slinger"),
			effects.Targeting(effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				active := g.Seats[g.Turn.ActiveSeat]
				return ev.Kind == game.EventCast && ev.Actor == source.Controller && active != nil && active.ID == source.Controller
			}, "Chakram Retriever — untap target creature", effects.UntapEachLegalTarget),
				effects.TargetCreature("target creature")),
		},
	})
}
