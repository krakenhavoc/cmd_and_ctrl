package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pacifism — Enchantment — Aura for {1}{W}:
//
//	"Enchant creature
//	 Enchanted creature can't attack or block."
//
// The card the restriction vocabulary was built for, and the reason
// it is a field on Characteristic rather than a string in the
// keyword list. "Can't attack" is not an ability the creature has —
// it is an effect the AURA has — so CR 613 gives it no layer, and
// "enchanted creature loses all abilities" does not switch it off.
// server/internal/game/restrictions.go has the argument in full.
//
// Two rules consequences the engine gets for free, both of them
// things players notice:
//
//   - The restriction is checked when attackers and blockers are
//     DECLARED (CR 508.1c, 509.1b). Pacifying a creature that is
//     already attacking does not remove it from combat — CR 506.4's
//     list of things that do is leaving the battlefield, changing
//     control and ceasing to be a creature, and this is none of
//     them. The creature stays in combat and deals its damage.
//   - Destroy the Aura and the creature attacks again in the same
//     beat, because the static's AppliesTo re-reads the attachment
//     on every recompute and there is no remembered state to unwind.
//
// White removal that leaves the body on the board: the creature is
// still a creature, still a Merfolk for your lord, still sacrificed
// to your altar. That is the deal, and it is why Pacifism is a
// weaker answer than Swords to Plowshares and a better one than
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f5e0b10-c8cf-450c-bfd3-bcb0528ec330",
		Name:         "Pacifism",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			RestrictAttached(game.CantAttackOrBlock),
		},
	})
}
