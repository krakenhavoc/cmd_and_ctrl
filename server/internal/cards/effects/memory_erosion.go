package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Memory Erosion — Enchantment {1}{U}{U} (EDHREC rank 3689):
//
//	"Whenever an opponent casts a spell, that player mills two cards."
//
// The mill deck's tax on the rest of the table. Sunscorch Regent's
// condition (b15OpponentCastSpell — any spell, any opponent, read
// off EventCast with the caster in Actor) with Memory Erosion's
// body: the caster, captured in Build by value, mills two, or the
// rest of their library when it holds fewer — a mill never loses a
// player the game (CR 701.17b). Copies are not cast, so a copied
// spell is silent, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f4a96881-586d-44ad-b427-5cdf8988f9a1",
		Name:         "Memory Erosion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b15OpponentCastSpell(ev, source)
			},
			Key: "Memory Erosion — that player mills two cards",
			Effect: func(g *game.Game, item *game.StackItem) error {
				caster := item.Trigger.Event.Actor
				if g.PlayerByIDForEffect(caster) == nil {
					return nil
				}
				return MillCards{Player: caster, N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
