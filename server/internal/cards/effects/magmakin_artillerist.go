package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magmakin Artillerist — 1/4 Creature — Elemental Pirate for {2}{R}:
//
//	"Whenever you discard one or more cards, this creature deals that
//	 much damage to each opponent.
//	 Cycling {1}{R}. When you cycle this card, it deals 1 damage to
//	 each opponent."
//
// Glint-Horn Buccaneer's effect on a body that doesn't need to
// attack. With the commander looting every turn this is a slow but
// inevitable clock on the whole table.
//
// Batching (CR 603.1): the card reads "one or more cards… that
// much damage", one trigger per batch. The engine emits one
// EventDiscardCard per card, so a two-card discard deals 1 twice
// rather than 2 once. Same total, two log lines.
//
// Cycling isn't modelled, so the cycling trigger can't fire either.
// Cycling is not a cast: it is an activated ability that works only
// from hand, "{1}{R}, Discard this card: Draw a card" (CR 702.29a),
// and the engine has neither a discard cost component nor activation
// from hand (#660). "When you cycle this card" triggers from wherever
// the card ends up, normally the graveyard (CR 702.29c).
func init() {
	Register(Spec{
		OracleID:     "900b9409-9c16-414d-8674-2ea42c2415a1",
		Name:         "Magmakin Artillerist",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"It can't be cycled, so the cycling damage trigger never happens; discarding several cards at once deals the damage as separate 1s."},
		Triggered: []game.TriggeredAbility{
			On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			}, "Magmakin Artillerist — 1 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			}),
		},
	})
}
