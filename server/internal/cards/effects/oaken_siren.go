package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oaken Siren — Artifact Creature — Siren Pirate {1}{U}, 1/2 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Flying, vigilance
//	 {T}: Add {U}. Spend this mana only to cast an artifact spell or
//	 activate an ability of an artifact source."
//
// Cargo Ship's restricted mana ability with {U} in place of {C}: the mana
// may be spent only to cast an artifact spell or to activate an ability
// whose source is an artifact (CR 106.6), and any other use is refused.
// It is a creature, so the {T} ability needs it to have been under its
// controller's control since the turn began (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a06a35a9-a23e-4a0a-90bc-84e7a52d2cef",
		Name:            "Oaken Siren",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "{T}: Add {U}. Spend this mana only to cast an artifact spell or activate an ability of an artifact source.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictType("Artifact")},
			)},
		}},
	})
}
