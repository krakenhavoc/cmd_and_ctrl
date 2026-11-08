package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Uthros, Titanic Godcore — Land — Planet (Edge of Eternities, #4461):
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Planet. Station only as a sorcery.)
//	 12+ | {U}, {T}: Add {U} for each artifact you control."
//
// Evendo, Waking Haven's shape with artifacts in place of creatures: a
// station Planet (CR 721.1) whose 12+ line is a mana ability gated on the
// charge counters (ManaAtChargeCounters), listed and greyed below twelve.
// The amount is Mox Opal's artifact count (ProducedPerPermanent), read as
// the ability resolves; Uthros is a land, not an artifact, so it does not
// count itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df08ac72-010f-42f8-beb3-6d645c638e1e",
		Name:         "Uthros, Titanic Godcore",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{U}",
				Label:    "Add {U}",
			},
			{
				Cost:           ManaAbilityCost{Tap: true, Mana: "{U}"},
				Condition:      ManaAtChargeCounters(12),
				ProducedAmount: ProducedPerPermanent("U", MatchArtifact),
				Label:          "12+ — {U}, {T}: Add {U} for each artifact you control",
			},
		},
		Activated: []ActivatedAbility{Station()},
	})
}
