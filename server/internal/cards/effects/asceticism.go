package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Asceticism — Enchantment {3}{G}{G}:
//
//	"Creatures you control have hexproof.
//	 {1}{G}: Regenerate target creature."
//
// The green "you do not get to have my board" enchantment, and the
// card #667 exists for. The hexproof half has worked since #353 —
// it is a plain Layer 6 keyword grant over "creatures you control",
// and game.CanBeTargetedBy has read hexproof at both the CR 601.2c
// announce gate and the CR 608.2b re-check ever since S23. The
// regenerate half had nowhere to go at all: the engine had no
// regeneration shield, so #95 filed the card under protection, #662
// corrected that, and the card waited for CR 701.19.
//
// It is a repeatable shield, not a one-shot: {1}{G} as many times as
// the mana lasts, and each activation stacks another shield on its
// target (CR 701.19a). Two activations on one creature survive two
// destructions, and any shield still unspent at the cleanup step is
// gone. A wipe that prints "they can't be regenerated" — Damnation,
// Wrath of God, Damn — beats it without spending the shields
// (CR 701.19c/d).
//
// Note the two halves answer DIFFERENT threats, which is why the card
// is played over either half alone: hexproof stops targeted removal,
// and it does nothing at all about a board wipe or combat damage,
// which is exactly what the shield is for.
func init() {
	Register(Spec{
		OracleID:     "9396546c-d067-4ce3-9c3d-b62cca970b4f",
		Name:         "Asceticism",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(b16CreaturesYouControl, "hexproof"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: Regenerate target creature.",
			Cost:    ManaCost("{1}{G}"),
			Targets: TargetCreature("target creature"),
			Effect:  regenerateTheTargetPermanent,
		}},
	})
}
