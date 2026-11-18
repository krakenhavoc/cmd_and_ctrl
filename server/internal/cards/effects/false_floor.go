package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// False Floor — Artifact {4}:
//
//	"This artifact enters tapped.
//	 Creatures enter tapped.
//	 {2}, {T}, Exile this artifact: Exile all untapped creatures. Activate
//	 only as a sorcery."
//
// Two replacements: its own tapped entry, and the symmetric "creatures
// enter tapped" (every player's, including a creature that enters
// alongside it). The exile is part of the cost (CR 602.2b), so the sweep
// resolves with the floor already gone; "untapped" is judged as the
// ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f6828a87-eebe-48dd-873e-251624107e38",
		Name:         "False Floor",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			SelfEntersTapped(),
			s82SNC22CreaturesEnterTapped("False Floor: creatures enter tapped"),
		},
		Activated: []ActivatedAbility{{
			Label:        "{2}, {T}, Exile this artifact: Exile all untapped creatures. Activate only as a sorcery.",
			Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile, Partial: true}},
			Cost:         Plus(ManaCost("{2}"), TapCost(), ExileThis()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ExileAllMatching{Match: And(Creature(), Untapped())}.Apply(NewContext(g, item))
			},
		}},
	})
}
