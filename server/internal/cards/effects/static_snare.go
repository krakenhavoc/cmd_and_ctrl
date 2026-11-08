package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Static Snare — Enchantment {4}{W} (Tarkir: Dragonstorm, #4463):
//
//	"Flash
//	 This spell costs {1} less to cast for each attacking creature.
//	 When this enchantment enters, exile target artifact or creature an
//	 opponent controls until this enchantment leaves the battlefield."
//
// Flash is a printed keyword. The reduction is Ancient Stone Idol's
// CostsLessEach over every attacking creature, whoever controls it, and it
// spends against generic mana only, so {W} is never reduced (CR 601.2f).
// The exile is Mystical Tether's (Glass Casket's): the target is chosen as
// the trigger goes on the stack (CR 603.3d), the card returns when the
// Snare leaves the battlefield (CR 610.3c), and if the Snare has already
// left as the trigger resolves, nothing is exiled (CR 610.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b5b742d2-f19b-47c9-84bd-e85d975900e6",
		Name:            "Static Snare",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(PermanentsOnBattlefield(AttackingCreature()),
				"This spell costs {1} less to cast for each attacking creature."),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Static Snare — exile target artifact or creature an opponent controls until this enchantment leaves the battlefield",
				exileChosenTargetUntilThisLeaves("Static Snare — the exiled card returns when Static Snare leaves the battlefield")),
				TargetPermanent("target artifact or creature an opponent controls", Or(Artifact(), Creature()), OpponentControls())),
		},
	})
}
