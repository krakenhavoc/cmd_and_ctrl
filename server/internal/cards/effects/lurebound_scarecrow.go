package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lurebound Scarecrow — Artifact Creature — Scarecrow {3}, 4/4:
//
//	"As this creature enters, choose a color.
//	 When you control no permanents of the chosen color, sacrifice this
//	 creature."
//
// ADR 0107 §1 (#1858). The colour is chosen as it enters (the catalog's
// as-enters colour choice, #742), for the bot as the colour it wants to
// keep. The sacrifice is a CR 603.8 state trigger over its controller's
// permanents of that colour, the Scarecrow itself included — it is
// colourless, so it never counts.
//
// The engine asks the colour as a prompt just after the permanent lands,
// and in the rules the choice is made before it is on the battlefield, so
// the state is not asked while that prompt is open. A Scarecrow that
// never chose (a permanent that became a copy of one) has no chosen colour,
// controls no permanent of it, and is sacrificed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8353f834-678a-4fa1-857a-8bfdfb8d6378",
		Name:         "Lurebound Scarecrow",
		Completeness: CompletenessFull,
		AsEnters:     ChooseColorAsEnters(game.ColorForBenefit, "Lurebound Scarecrow"),
		Triggered: []game.TriggeredAbility{
			WhenState("Lurebound Scarecrow — sacrifice it",
				func(g *game.Game, source *game.Card, controller uuid.UUID) bool {
					if source.ChosenColor == "" && colorChoicePendingFor(g, source.InstanceID) {
						return false
					}
					color := source.ChosenColor
					return !controlsAnyCard(g, controller, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return color != "" && c.HasColor(color)
					})
				}, SacrificeThisIfStillOnBattlefield),
		},
	})
}

// colorChoicePendingFor reports whether `source`'s colour choice is still
// open — the beat between a permanent landing and its "as this enters,
// choose a color" being answered.
func colorChoicePendingFor(g *game.Game, source uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceColor && c.Source == source {
			return true
		}
	}
	return false
}
