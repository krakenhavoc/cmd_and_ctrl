package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Caetus, Sea Tyrant of Segovia — Legendary Creature — Serpent, 3/3, the
// back face of Invasion of Segovia (March of the Machine, #5647):
//
//	"Noncreature spells you cast have convoke.
//	 At the beginning of your end step, untap up to four target
//	 creatures."
//
// The static is a GrantedTapCost over noncreature spells
// (SpellsYouCastHaveConvoke): each one may be paid for by tapping
// creatures, and is a spell that has convoke for anything that asks
// (Joyful Stormsculptor). A spell with convoke of its own gains nothing
// more (CR 702.51d). The end-step trigger targets zero to four creatures,
// anyone's, and untaps each still legal as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     invasionOfSegoviaOracleID + "#1",
		Name:         "Caetus, Sea Tyrant of Segovia",
		Completeness: CompletenessFull,
		GrantedTapCosts: []game.GrantedTapCost{SpellsYouCastHaveConvoke(
			"Noncreature spells you cast have convoke.", SpellIs(Noncreature()))},
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourEndStep("Caetus, Sea Tyrant of Segovia — untap up to four target creatures",
				UntapEachLegalTarget),
				TargetCreature("up to four target creatures").WithCount(0, 4)),
		},
	})
}
