package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tiller of Flesh — Creature — Phyrexian Knight {3}{W}, 2/4 (March of
// the Machine):
//
//	"Whenever you cast a spell that targets one or more permanents,
//	 incubate 2."
//
// Read at the cast (CR 601.2i) from the targets the spell announced: a
// card target still on the battlefield. A player, a spell on the stack
// and a card in a graveyard are not permanents. A spell with several
// permanent targets triggers once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b6d2f392-5637-49ec-b6cb-603edd89e788",
		Name:         "Tiller of Flesh",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				return rfCreatureACastSpellTargetsAny(g, ev.CardID, func(t game.TargetRef) bool {
					if t.Kind != game.TargetCard {
						return false
					}
					z := g.FindCardZoneForEffect(t.ID)
					return z != nil && z.Kind == game.ZoneBattlefield
				})
			}, "Tiller of Flesh — incubate 2", Do(Incubate{N: 2})),
		},
	})
}
