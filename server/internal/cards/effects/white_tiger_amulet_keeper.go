package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// White Tiger, Amulet Keeper — Legendary Creature — Human Hero {1}{G},
// 3/1 (Marvel Super Heroes Commander, #4455):
//
//	"{3}{G}, Exile this card from your graveyard: Draw a card. You may
//	 put a land card from your hand onto the battlefield."
//
// A graveyard ability whose cost exiles the card itself (Buried
// Treasure's shape, CR 113.6), at instant speed. The draw comes first, so
// the card just drawn can be the land; the put is optional
// (PutFromHandOntoBattlefield, Optional) and a hand with no land is not
// asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0fa01155-638d-44ed-8963-46b202b0e720",
		Name:         "White Tiger, Amulet Keeper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}, Exile this card from your graveyard: Draw a card. You may put a land card from your hand onto the battlefield.",
			Cost:    Plus(ManaCost("{3}{G}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: Do(
				DrawCards{N: 1},
				PutFromHandOntoBattlefield{
					Match:    Land(),
					Optional: true,
					Label:    "White Tiger, Amulet Keeper — you may put a land card from your hand onto the battlefield",
				}),
		}},
	})
}
