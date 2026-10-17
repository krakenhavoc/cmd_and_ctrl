package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mayor of Avabruck // Howlpack Alpha — {1}{G} Creature — Human Advisor
// Werewolf 1/1 // Creature — Werewolf 3/3 (#4071, ADR 0132 amendment
// 2026-10-16):
//
//	Front: "Other Human creatures you control get +1/+1.
//	        At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "Each other creature you control that's a Werewolf or a Wolf
//	        gets +1/+1.
//	        At the beginning of your end step, create a 2/2 green Wolf
//	        creature token.
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// Both lords are the shared tribal anthem over your OTHER creatures of
// the type, so turning over swaps which tribe gets the bonus in the same
// layer pass. The end-step token is the Alpha's controller's only.
//
// No simplification.
func init() {
	const oracle = "6e638587-30b3-4c4b-b463-fb415ea048f7"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Mayor of Avabruck",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Human"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			TransformAtUpkeepIfNoSpellsLastTurn("Mayor of Avabruck"),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Howlpack Alpha",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Werewolf", "Wolf"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Howlpack Alpha — create a 2/2 green Wolf creature token",
				Do(CreateToken{Template: TokenCard("2/2 green Wolf"), N: 1})),
			TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Howlpack Alpha"),
		},
	})
}
