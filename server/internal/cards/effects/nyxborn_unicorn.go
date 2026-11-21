package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const nyxbornUnicornMentor = "nyxborn-unicorn/mentor"

// Nyxborn Unicorn — Enchantment Creature — Unicorn {1}{W}, 2/2 (Modern
// Horizons 3):
//
//	"Bestow {3}{W} (If you cast this card for its bestow cost, it's an Aura
//	 spell with enchant creature. It becomes a creature again if it's not
//	 attached.)
//	 Mentor (Whenever this creature attacks, put a +1/+1 counter on target
//	 attacking creature with lesser power.)
//	 Enchanted creature gets +2/+2 and has mentor."
//
// Bestow is the shared alternative cost (ADR 0141); the pump is
// PumpAttached and the granted mentor an ADR 0093 bundle over the enchanted
// creature, the same shared Mentor() trigger the card carries itself
// (Aegis of the Legion's shape). A bestowed Unicorn is not a creature, so
// its own mentor never triggers; when it falls off it is a creature again
// and attacks as one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "b9304d88-a1de-4300-9425-50350969bd70",
		Name:             "Nyxborn Unicorn",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Bestow("{3}{W}")},
		Grants: []AbilityGrant{{
			Key:       nyxbornUnicornMentor,
			Triggered: []game.TriggeredAbility{Mentor("Nyxborn Unicorn — mentor")},
			Text:      "Mentor (Whenever this creature attacks, put a +1/+1 counter on target attacking creature with lesser power.)",
		}},
		Triggered: []game.TriggeredAbility{Mentor("Nyxborn Unicorn — mentor")},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantAbilitiesToAttached(nyxbornUnicornMentor),
		},
	})
}
