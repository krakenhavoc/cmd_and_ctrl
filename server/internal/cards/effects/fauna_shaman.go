package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fauna Shaman — 2/2 Creature — Elf Shaman for {1}{G}:
//
//	"{G}, {T}, Discard a creature card: Search your library for a
//	 creature card, reveal it, put it into your hand, then shuffle."
//
// Survival of the Fittest on a body, and the first card in the
// catalog to pay a DISCARD as part of an activated ability's cost
// (#660, engine-seams row "Discard-a-card cost on an activated
// ability"). The component is general — N cards matching a clause —
// which is what Survival of the Fittest, Tortured Existence and
// Cryptbreaker are waiting on next.
//
// Three things the shared cost path gets right without this file
// saying anything:
//
//   - The discard happens at ANNOUNCE, so a Marauding Mako sees it
//     and its trigger goes on the stack ABOVE the search. Folding the
//     discard into the effect would invert that and would hand the
//     card back if the ability were countered.
//   - It goes through the one discard helper with cause cost, so the
//     CR 614 window runs over it and it cannot pause (CR 602.2b) —
//     which is what makes pitching a commander here land in the
//     graveyard rather than open a CR 903.9 prompt mid-payment.
//   - The activator picks the creature card from hand in the client's
//     cost picker and sends it as `discard_ids`; the engine re-checks
//     that it is in hand and that it is a creature card.
//
// The search is the plain "reveal it, put it into your hand, then
// shuffle" — an unsuccessful search still shuffles (CR 701.19).
func init() {
	Register(Spec{
		OracleID:     "35b8fa77-4e85-418b-b335-cd1af127075c",
		Name:         "Fauna Shaman",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{G}, {T}, Discard a creature card: Search your library for a creature card, reveal it, put it into your hand, then shuffle.",
			Cost:  Plus(ManaCost("{G}"), TapCost(), DiscardCardsMatching(1, "a creature card", MatchCreature)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: MatchCreature,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Fauna Shaman — a creature card, to your hand",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
