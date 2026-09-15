package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Managorger Hydra — Creature — Hydra, {2}{G}, 1/1 (EDHREC rank 1004):
//
//	"Trample
//	 Whenever a player casts a spell, put a +1/+1 counter on this
//	 creature."
//
// A three-drop that grows off the whole table's spells — at a
// four-player table it is a 5/5 within a turn cycle. One cast
// trigger with no controller gate ("a player", yours included), a
// counter on itself, and printed trample so the size matters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b3f2265b-dd65-4b74-8b74-35ee0b147617",
		Name:            "Managorger Hydra",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil && ev.CardID != uuid.Nil
			}, "Managorger Hydra — a +1/+1 counter", func(g *game.Game, item *game.StackItem) error {
				if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneBattlefield {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
