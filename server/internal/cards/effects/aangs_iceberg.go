package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aang's Iceberg — Enchantment {2}{W}:
//
//	"Flash
//	 When this enchantment enters, exile up to one other target nonland
//	 permanent until this enchantment leaves the battlefield.
//	 Waterbend {3}: Sacrifice this enchantment. If you do, scry 2."
//
// The enters trigger is Earth Kingdom Jailer's exile-until-this-leaves
// (CR 610.3) over Deceit's "up to one other target nonland permanent".
// If the Iceberg has already left when it resolves (sacrificed to its
// own ability in response, say), nothing is exiled (the ruling,
// CR 610.3b).
//
// The waterbend ability sacrifices on RESOLUTION, not as a cost (the
// ruling): SacrificeThisThen sacrifices only the object that activated
// it, still on the battlefield, and scries 2 only if it did. Waterbend
// {3} is WaterbendCost: each generic may be paid by tapping an untapped
// artifact or creature you control (CR 701.67a).
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:        "646f9ad8-9d18-4b42-954a-e21a62e9f0e8",
		Name:            "Aang's Iceberg",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Aang's Iceberg — exile up to one other target nonland permanent until this enchantment leaves the battlefield",
				exileChosenTargetUntilThisLeaves("Aang's Iceberg — the exiled card returns when Aang's Iceberg leaves the battlefield")),
				Another(TargetPermanent("up to one other target nonland permanent", Nonland()).WithCount(0, 1))),
		},
		Activated: []ActivatedAbility{{
			Label:   "Waterbend {3}: Sacrifice this enchantment. If you do, scry 2.",
			Purpose: game.Purpose{Scry: 2, Answers: game.AnswerValue},
			Cost:    WaterbendCost("{3}"),
			Effect: SacrificeThisThen("then/aangs-iceberg-scry", func(ctx *Context) error {
				return Scry{N: 2}.Apply(ctx)
			}),
		}},
	})
}
