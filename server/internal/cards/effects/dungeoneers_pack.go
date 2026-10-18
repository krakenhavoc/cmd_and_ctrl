package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dungeoneer's Pack — Artifact {3}:
//
//	"This artifact enters tapped.
//	 {2}, {T}, Sacrifice this artifact: You take the initiative, gain 3
//	 life, draw a card, and create a Treasure token. Activate only as a
//	 sorcery."
//
// Enters tapped is the self-replacement every tapped permanent uses. The
// ability is a sorcery-speed CR 602 ability whose cost includes
// sacrificing the Pack; its effect runs in the printed order, and taking
// the initiative (ADR 0203, #4096) triggers the venture into Undercity,
// which goes on the stack once the ability has resolved.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c66976d3-57f1-4c48-ada0-d74081f1699a",
		Name:         "Dungeoneer's Pack",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Activated: []ActivatedAbility{{
			Label:        "{2}, {T}, Sacrifice this artifact: You take the initiative, gain 3 life, draw a card, and create a Treasure token. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			SorcerySpeed: true,
			Effect: Do(
				TakeTheInitiative{},
				GainLife{Amount: 3},
				DrawCards{N: 1},
				CreateToken{Template: TreasureToken(), N: 1},
			),
		}},
	})
}
