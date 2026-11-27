package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b568_helpers.go — shared pieces of Coldsnap batch 568 (#5597, S82).

// S82OldB568PreventDamageToThisFromFirstStrikers is a standing CR 615 prevention
// of damage dealt to the permanent that has it by creatures with first strike
// (Tresserhorn Skyknight): the source as it was when the damage would be dealt
// (its last-known information, CR 608.2h), combat damage or not. A creature
// with double strike but not first strike (it has double strike, not first
// strike) is not one of them.
func S82OldB568PreventDamageToThisFromFirstStrikers(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true, // CR 615.1a
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageTarget != src.InstanceID {
				return false
			}
			lki := ev.SourceLKI
			return lki != nil && slices.Contains(lki.Types, "Creature") && slices.Contains(lki.Abilities, "first strike")
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.Cancel()
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      label,
	}
}
