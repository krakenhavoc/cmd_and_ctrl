package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Belbe's Portal — Artifact {5} (Nemesis, S82 OLD batch 679, #5618):
//
//	"As this artifact enters, choose a creature type.
//	 {3}, {T}: You may put a creature card of the chosen type from your hand
//	 onto the battlefield."
//
// Quicksilver Amulet's put (not cast: no counterspell answers it, no cast
// trigger fires, ETB triggers do) narrowed to the type the Portal named as it
// entered (effects.ChooseCreatureTypeAsEnters, kept on the permanent). The type
// is read through the layered view, so a changeling card counts. "You may" is a
// real decline, and a hand with no creature card of the type is not prompted.
// Until a type is chosen nothing matches.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46e50ee9-91f8-416e-82fa-f1a76e61fb73",
		Name:         "Belbe's Portal",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCreatureTypeAsEnters("Belbe's Portal"),
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: You may put a creature card of the chosen type from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				tribe := g.NamedTribeOf(item.SourceCardID)
				if tribe == "" {
					return nil
				}
				return effects.PutFromHandOntoBattlefield{
					Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.IsCreature() && c.HasSubtypeWord(tribe)
					},
					Optional: true,
					Label:    "Belbe's Portal — you may put a creature card of the chosen type from your hand onto the battlefield",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
