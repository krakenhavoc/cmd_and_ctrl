package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// untap_step.go — the shared shape of "untap <these> during each
// other player's untap step" (Seedborn Muse, Unwinding Clock,
// Drumbellower, Bender's Waterskin, the second half of Quest for
// Renewal).
//
// Every card in the family is the same sentence with a different
// noun, so the whole family is one constructor and a predicate. See
// game/untap.go for why this is a permission on a turn-based action
// rather than a triggered ability, and Spec.UntapStep for what the
// difference buys.

// eachOtherPlayersUntapStep is the AppliesTo shared by the whole
// family: the untap step being entered belongs to someone other than
// the source's controller.
//
// "Each OTHER player" is how all of them are printed, and it is not
// a restriction the engine could skip — the controller's own
// permanents already untap during their own untap step by CR 502.1,
// so a permission that also fired there would be dead weight on
// every one of their turns.
func eachOtherPlayersUntapStep(_ *game.Game, source *game.Card, activePlayer uuid.UUID) bool {
	return activePlayer != uuid.Nil && activePlayer != source.Controller
}

// untapDuringEachOtherPlayersUntapStep builds the permission for
// "untap all <match> you control during each other player's untap
// step". A nil `match` is "all permanents you control" — Seedborn
// Muse, which names no type.
//
// `match` reads a game.Card by value and should use the effective
// type predicates (IsArtifact, IsCreature): Unwinding Clock says
// "artifacts you control", and a land animated into an artifact is
// one for as long as it is one.
func untapDuringEachOtherPlayersUntapStep(label string, match func(game.Card) bool) game.UntapStepPermission {
	return game.UntapStepPermission{
		Label:     label,
		AppliesTo: eachOtherPlayersUntapStep,
		Untaps: func(_ *game.Game, source, target *game.Card) bool {
			if target.Controller != source.Controller {
				return false
			}
			return match == nil || match(*target)
		},
	}
}

// untapSelfDuringEachOtherPlayersUntapStep is the same clause aimed
// at the source alone — Bender's Waterskin's "untap this artifact
// during each other player's untap step", which is how a mana rock
// is made to pay for something on every turn of the table rather
// than one.
func untapSelfDuringEachOtherPlayersUntapStep(label string) game.UntapStepPermission {
	return game.UntapStepPermission{
		Label:     label,
		AppliesTo: eachOtherPlayersUntapStep,
		Untaps: func(_ *game.Game, source, target *game.Card) bool {
			return target.InstanceID == source.InstanceID
		},
	}
}
