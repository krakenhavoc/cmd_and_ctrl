package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jidoor, Aristocratic Capital // Overture — Land — Town, with a
// Sorcery — Adventure half {4}{U}{U} (oracle bd513d9d…, #2176):
//
//	Jidoor, Aristocratic Capital
//	  "This land enters tapped.
//	   {T}: Add {U}."
//	Overture
//	  "Target opponent mills half their library, rounded down. (Then
//	   exile this card. You may play the land later from exile.)"
//
// The first Adventure whose main half is a land (the Final Fantasy
// Town cycle). Playing the land is a land play and not a cast, so
// from hand it is the ordinary special action; Overture is the
// Adventure half, cast like any other. CR 715.3d's grant on the
// exiled card is a PLAY permission when face 0 is a land
// (game/adventure.go: adventureMainHalfIsLand), so the land can be
// played from exile for as long as it stays there, spending the
// turn's land drop (CR 305.2) and only at sorcery timing on its
// controller's own turn. A countered Overture goes to the graveyard
// like any other countered Adventure.
//
// Two keys, one card: the land keeps the bare oracle ID and Overture
// takes "<oracle_id>#1" (ADR 0034 §5).
//
// No simplification.
const jidoorOracleID = "bd513d9d-5aa2-4860-bd86-8b5d9430f133"

func init() {
	Register(Spec{
		OracleID:     jidoorOracleID,
		Name:         "Jidoor, Aristocratic Capital",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
	})
	Register(Spec{
		OracleID:     jidoorOracleID + "#1",
		Name:         "Overture",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve:    millHalfOfTargetPlayer,
	})
}
