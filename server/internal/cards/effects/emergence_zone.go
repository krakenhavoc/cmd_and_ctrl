package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emergence Zone — Land:
//
//	"{T}: Add {C}.
//	 {1}, {T}, Sacrifice this land: You may cast spells this turn as
//	 though they had flash."
//
// The STORED half of #1195, and the card that makes the two homes
// necessary rather than merely tidy: the land sacrifices itself to
// pay for the ability, so there is no permanent left on the
// battlefield for a derivation to read. The permission is written
// onto the player with ADR 0063's `UntilEndOfTurn` duration and swept
// at that turn's cleanup step (CR 514.2) by the same
// `durationExpiredLocked` every continuous effect in the game uses.
//
// `GrantCastTiming` with no window is exactly "this turn": the one
// write path stamps an unstamped duration against the current turn,
// the posture `GrantCastPermissionForEffect` already takes, so a card
// file that forgets to say gets the narrowest real window rather than
// a permanent grant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7536eb66-959d-4dca-9b75-895572ef733c",
		Name:         "Emergence Zone",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}, Sacrifice this land: You may cast spells this turn as though they had flash.",
			Cost:  Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantCastTiming{
					Label: "You may cast spells this turn as though they had flash.",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
