package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Paradox Shaper // Omit Variables — Creature — Octopus Wizard {1}{U/B},
// 1/3 // Sorcery {U/B} (preparation card, CR 722):
//
//	"At the beginning of your upkeep, if this creature isn't prepared,
//	 it becomes prepared.
//	 {2}: Put target card from your graveyard on the bottom of your
//	 library."
//
//	Omit Variables — "Mill three cards."
//
// It does not enter prepared: the upkeep trigger is its only source of
// the designation. The activated ability targets, so a card that leaves
// the graveyard in response fizzles it.
//
// No simplification.
func init() {
	const id = "511951ed-fbff-4e44-9429-27f237496672"
	Register(Spec{
		OracleID:     id,
		Name:         "Paradox Shaper",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{fraBecomesPreparedAtUpkeep("Paradox Shaper")},
		Activated: []ActivatedAbility{{
			Label:   "{2}: Put target card from your graveyard on the bottom of your library.",
			Cost:    ManaCost("{2}"),
			Targets: TargetCardInGraveyard("target card in your graveyard", YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return PutInLibraryInAnyOrder{
							Cards:     []uuid.UUID{t.ID},
							From:      game.ZoneGraveyard,
							Placement: game.LibraryPlaceBottom,
							Label:     "Paradox Shaper — put the card on the bottom of your library",
						}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Omit Variables",
		Completeness: CompletenessFull,
		OnResolve:    omitVariablesResolve,
	})
}
