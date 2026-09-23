package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exhaust_permission.go — #1184: the vocabulary a card file writes
// Spec.ExhaustPermissions in. The engine half is
// game/exhaust_permission.go.
//
// One card prints this today (Elvish Refueler) and the constructors
// are still named for the printed clause rather than for it, because
// the clause is what the next one will print too:
//
//	ExhaustPermissions: []game.ExhaustPermission{
//	    MayActivateExhaustAbilitiesAgain(
//	        "<the oracle sentence>",
//	        DuringTheControllersTurn(),
//	        ControllerHasActivatedNoExhaustAbilityThisTurn()),
//	},
//
// Predicates are ANDed and read left to right in the order the card
// says them, exactly as a CostPredicate list does.

// ExhaustPermissionPredicate narrows a permission to the moments its
// card actually grants it. `asker` is the player the exhaust gate is
// being answered for; `source` is the permanent granting the
// permission.
//
// READ-ONLY and under g.mu — the same contract every predicate in the
// catalog has.
type ExhaustPermissionPredicate func(g *game.Game, asker uuid.UUID, source game.Card) bool

// MayActivateExhaustAbilitiesAgain is "you may activate exhaust
// abilities as though they haven't been activated", narrowed by the
// printed conditions.
//
// The permission always reaches its own controller and only its own
// controller: every printed clause in this family says "you". A card
// that ever says "each player" adds a predicate here rather than a
// second constructor — but until one is printed, the "you" is the
// default rather than something a card file can forget, because
// forgetting it would hand every opponent the permission.
func MayActivateExhaustAbilitiesAgain(label string, when ...ExhaustPermissionPredicate) game.ExhaustPermission {
	return game.ExhaustPermission{
		Label: label,
		Applies: func(g *game.Game, asker uuid.UUID, source game.Card) bool {
			if asker == uuid.Nil || asker != source.Controller {
				return false
			}
			for _, p := range when {
				if p == nil || !p(g, asker, source) {
					return false
				}
			}
			return true
		},
	}
}

// DuringTheControllersTurn — "During your turn". The permission is a
// static ability, so this is read as the gate is asked rather than
// latched at any point.
func DuringTheControllersTurn() ExhaustPermissionPredicate {
	return func(g *game.Game, asker uuid.UUID, _ game.Card) bool {
		return g != nil && IsYourTurn(g, asker)
	}
}

// ControllerHasActivatedNoExhaustAbilityThisTurn — "as long as you
// haven't activated an exhaust ability this turn".
//
// Counts BOTH activation kinds, which is what makes the permission
// self-limiting rather than infinite: the exhaust ability the player
// activates under it is itself an exhaust ability activated this
// turn, so the permission is off for the rest of the turn the moment
// it is used, and nothing has to turn it off.
func ControllerHasActivatedNoExhaustAbilityThisTurn() ExhaustPermissionPredicate {
	return func(g *game.Game, asker uuid.UUID, _ game.Card) bool {
		return g != nil && g.ExhaustAbilitiesActivatedThisTurn(asker) == 0
	}
}
