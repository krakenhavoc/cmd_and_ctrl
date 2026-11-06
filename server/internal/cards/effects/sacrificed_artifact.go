package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrificed_artifact.go — "if you've sacrificed an artifact this turn"
// (#5224).
//
// One fact about the player's TURN, read off the per-player card-type
// sacrifice tally (game.TurnTally.SacrificedCardTypes) through
// Game.SacrificedWithCardTypeThisTurn: each permanent the player
// sacrificed counts once per card type it had as it was sacrificed, so
// an artifact creature is one artifact, and a token counts though it no
// longer exists anywhere (CR 704.5d). The Food readers keep their own
// subtype tally (b34YouSacrificedAFoodThisTurn).
//
// Readers, one per place the printed text puts the condition:
//
//   - SacrificedAnArtifactThisTurn(g, you) for a resolution-time check.
//   - YouSacrificedAnArtifactThisTurn for an activated ability's
//     "Activate only if …" (ActivatedAbility.Condition).
//   - CondSacrificedAnArtifact for an intervening "if" (InterveningIf,
//     If), with a monotone footprint condition.

// SacrificedAnArtifactThisTurn reports whether `you` sacrificed an
// artifact this turn. Caller must hold g.mu (every effect body, static
// AppliesTo, block rule and activation condition does).
func SacrificedAnArtifactThisTurn(g *game.Game, you uuid.UUID) bool {
	return g.SacrificedWithCardTypeThisTurn(you, "artifact") > 0
}

// YouSacrificedAnArtifactThisTurn is "Activate only if you've sacrificed
// an artifact this turn" (Detective's Satchel).
func YouSacrificedAnArtifactThisTurn(g *game.Game, controller, _ uuid.UUID) bool {
	return SacrificedAnArtifactThisTurn(g, controller)
}

// CondSacrificedAnArtifact is the condition as an intervening-if
// Condition, for InterveningIf and If: read for the item's controller as
// it resolves. The tally only rises within a turn, so it declares a
// footprint condition the trigger-independence check can argue
// (game.FootprintCondArtifactSacrificed).
type CondSacrificedAnArtifact struct{}

func (CondSacrificedAnArtifact) Holds(ctx *Context) bool {
	return SacrificedAnArtifactThisTurn(ctx.Game, ctx.Controller())
}

func (CondSacrificedAnArtifact) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondArtifactSacrificed, 0), true
}
