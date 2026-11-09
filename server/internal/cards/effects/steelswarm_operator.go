package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steelswarm Operator — Artifact Creature — Robot Soldier {1}{U}, 1/1
// (Edge of Eternities, #4461):
//
//	"Flying
//	 {T}: Add {U}. Spend this mana only to cast an artifact spell.
//	 {T}: Add {U}{U}. Spend this mana only to activate abilities of
//	 artifact sources."
//
// Two mana abilities with spend restrictions (CR 106.6). The first is
// Ancient Ziggurat's shape: cast, and the spell has the artifact type. The
// second is Pit Automaton's "only to activate abilities" narrowed by the
// type of the ability's SOURCE (ManaSpendForAbility reads the source's
// characteristics): any object with the artifact card type counts, an
// artifact card in a hand or a graveyard included (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "770b2097-5178-42d2-be6c-393b4bfe3727",
		Name:            "Steelswarm Operator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ManaAbilities: []ManaAbility{
			{
				Cost:         ManaAbilityCost{Tap: true},
				Produced:     "{U}",
				Label:        "{T}: Add {U}. Spend this mana only to cast an artifact spell.",
				Restrictions: []string{game.ManaRestrictCast, ManaRestrictType("Artifact")},
			},
			{
				Cost:         ManaAbilityCost{Tap: true},
				Produced:     "{U}{U}",
				Label:        "{T}: Add {U}{U}. Spend this mana only to activate abilities of artifact sources.",
				Restrictions: []string{game.ManaRestrictActivate, ManaRestrictType("Artifact")},
			},
		},
	})
}
