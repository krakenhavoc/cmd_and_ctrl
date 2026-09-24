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
// Cycling {1}{R} arrived with #660 and works: an activated ability
// from hand (CR 702.29a) whose discard feeds the damage trigger above
// on every OTHER Artillerist on the board.
//
// The card's OWN cycle trigger fires from the GRAVEYARD (CR 702.29c —
// the card is already there when the ability triggers), watching via
// `InGraveyard`, #925's zone dimension on TriggeredAbility. That
// closed the gap ADR 0062 Decision 7 deferred (its 2026-09-18
// amendment). Cycling this card now deals 2 total: 1 from the
// discard trigger above (the discard IS the cycling activation's
// cost) and 1 from the cycle trigger below.
func init() {
	Register(Spec{
		OracleID:     "900b9409-9c16-414d-8674-2ea42c2415a1",
		Name:         "Magmakin Artillerist",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Discarding several cards at once deals the damage as separate 1s.",
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			}, "Magmakin Artillerist — 1 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			}),
			InGraveyard(On(game.EventCycle, Self,
				"Magmakin Artillerist — cycled: 1 damage to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					return damageToEachOpponent(g, item, 1)
				})),
		},
		Activated: []ActivatedAbility{Cycling("{1}{R}")},
	})
}
