package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mishra, Excavation Prodigy — Legendary Creature — Human Artificer
// {2}{R}, 2/1 (The Brothers' War):
//
//	"Haste
//	 {1}, {T}, Discard a card: Draw a card.
//	 Whenever you discard one or more artifact cards, add {R}{R}. This
//	 ability triggers only once each turn."
//
// The mana is a triggered ability that goes on the stack and resolves
// (it is not a mana ability: a discard is not a mana ability resolving,
// CR 605.1b), so the {R}{R} lands in the pool as that trigger resolves.
// The trigger is the shared Urza/Mishra one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4eba0e29-985b-4763-9a70-0946ba044b43",
		Name:            "Mishra, Excavation Prodigy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Discard a card: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Discards: 1, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost(), DiscardACard()),
			Effect:  Do(DrawCards{N: 1}),
		}},
		Triggered: []game.TriggeredAbility{
			wheneverYouDiscardOneOrMoreArtifactCardsOnce(
				"Mishra, Excavation Prodigy — add {R}{R}",
				Do(AddMana{Produced: "{R}{R}"})),
		},
	})
}
