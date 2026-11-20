package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Horizon Spellbomb — Artifact {1} (Scars of Mirrodin, #5577):
//
//	"{2}, {T}, Sacrifice this artifact: Search your library for a basic land
//	 card, reveal it, put it into your hand, then shuffle.
//	 When this artifact is put into a graveyard from the battlefield, you
//	 may pay {G}. If you do, draw a card."
//
// A Spellbomb: the first ability is Ominous Parcel's basic land search,
// the second Nihil Spellbomb's draw (somB2SpellbombDraw).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "12080ea9-0800-4601-b553-6711dd9f9806",
		Name:         "Horizon Spellbomb",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Horizon Spellbomb — a basic land card, to your hand",
				}.Apply(NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{somB2SpellbombDraw("Horizon Spellbomb", "{G}")},
	})
}
