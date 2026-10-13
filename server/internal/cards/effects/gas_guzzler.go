package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gas Guzzler — Creature — Vampire Rogue {B}, 2/1:
//
//	"Start your engines!
//	 This creature enters tapped.
//	 Max speed — {B}, Sacrifice another creature or Vehicle: Draw a
//	 card."
//
// ADR 0138 (#2122, #2711). Enters tapped is the self replacement, not a
// tap after entering. The sacrifice is paid at activation (CR 602.2b) and
// can never be Gas Guzzler itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7c98626b-ba97-4d1e-9be7-291480ae8d22",
		Name:            "Gas Guzzler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		Replacements:    []game.ReplacementEffect{SelfEntersTapped()},
		Activated: []ActivatedAbility{
			MaxSpeedActivated(ActivatedAbility{
				Label:   "Max speed — {B}, Sacrifice another creature or Vehicle: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost: Plus(ManaCost("{B}"),
					SacrificeAnotherN(1, "another creature or Vehicle", Or(Creature(), OfSubtype("Vehicle")))),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{N: 1}.Apply(NewContext(g, item))
				},
			}),
		},
	})
}
