package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clever Conjurer — Creature — Gnome Wizard {2}{U}, 2/3 (Adventures in
// the Forgotten Realms):
//
//	"Mage Hand — {T}: Untap target permanent not named Clever Conjurer.
//	 Activate only as a sorcery."
//
// Aphetto Alchemist's untap on a tap ability, with the name bar read off
// the target as it is chosen and again as it resolves (CR 608.2b); a
// second Conjurer is therefore not a legal target, the activating one
// included. "Mage Hand" has no rules meaning (CR 207.2c).
//
// No simplification.
func init() {
	notConjurer := func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Name != "Clever Conjurer" }
	Register(Spec{
		OracleID:     "628e59e2-3665-40dd-ba4f-11ab5824de22",
		Name:         "Clever Conjurer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "Mage Hand — {T}: Untap target permanent not named Clever Conjurer. Activate only as a sorcery.",
			Cost:         TapCost(),
			Targets:      TargetPermanent("target permanent not named Clever Conjurer", notConjurer),
			SorcerySpeed: true,
			Effect:       untapTheTarget,
		}},
	})
}
