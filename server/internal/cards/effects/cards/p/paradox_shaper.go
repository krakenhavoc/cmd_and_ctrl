package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Paradox Shaper",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.FraBecomesPreparedAtUpkeep("Paradox Shaper")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Put target card from your graveyard on the bottom of your library.",
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetCardInGraveyard("target card in your graveyard", effects.YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.PutInLibraryInAnyOrder{
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
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Omit Variables",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.OmitVariablesResolve,
	})
}
