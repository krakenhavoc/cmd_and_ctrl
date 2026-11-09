package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Liminal Hold — Enchantment {3}{W} (Lorwyn Eclipsed, #4458):
//
//	"When this enchantment enters, exile up to one target nonland
//	 permanent an opponent controls until this enchantment leaves the
//	 battlefield. You gain 2 life."
//
// Makeshift Binding's exile-until-leaves and life gain with an "up to one"
// target. The exile records its own return (CR 610.3), so a token exiled
// this way ceases to exist and an Aura returns to a choice of what it
// enchants. If this leaves before the trigger resolves nothing is exiled
// (CR 610.3b), but the life is still gained while the target is legal; with
// no target chosen the life is gained too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "101407b4-e7e8-4aff-84a2-a97947358b88",
		Name:         "Liminal Hold",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Liminal Hold — exile up to one target nonland permanent an opponent controls until this enchantment leaves the battlefield; you gain 2 life",
				exileChosenTargetUntilThisLeavesThenGainLife(
					"Liminal Hold — the exiled permanent returns when Liminal Hold leaves the battlefield", 2)),
				TargetPermanent("up to one target nonland permanent an opponent controls", And(Nonland(), OpponentControls())).WithCount(0, 1)),
		},
	})
}
