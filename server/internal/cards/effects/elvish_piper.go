package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elvish Piper — Creature — Elf Shaman {3}{G}, 1/1 (Urza's Destiny,
// #5621):
//
//	"{G}, {T}: You may put a creature card from your hand onto the
//	 battlefield."
//
// Quicksilver Amulet's put at {G} on a creature: on resolution its
// controller may pick a creature card from their hand, or decline, and
// the card enters without being cast. {T} needs the Piper to have been
// under its controller's control since their most recent turn began
// (CR 302.6).
//
// The put was the seam #391 skipped this card for; Quicksilver Amulet
// shipped it (PutFromHandOntoBattlefield).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "872ef617-7cdb-4a7a-85f5-efc9e12bece5",
		Name:         "Elvish Piper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{G}, {T}: You may put a creature card from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{G}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PutFromHandOntoBattlefield{
					Match:    Creature(),
					Optional: true,
					Label:    "Elvish Piper — you may put a creature card from your hand onto the battlefield",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
