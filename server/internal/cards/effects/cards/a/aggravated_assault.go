package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aggravated Assault — Enchantment {2}{R}:
//
//	"{3}{R}{R}: Untap all creatures you control. After this main
//	 phase, there is an additional combat phase followed by an
//	 additional main phase. Activate only as a sorcery."
//
// The Relentless Assault shape on a repeatable ability (ADR 0059
// sub-PR 2b, #753). Sorcery timing means it resolves in a main phase,
// so the anchor always holds; activating it again in the added main
// phase adds another combat and main after THAT one, which is what
// the card does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20129459-a386-41eb-899d-1aede3427300",
		Name:         "Aggravated Assault",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{R}{R}: Untap all creatures you control. After this main phase, there is an additional combat phase followed by an additional main phase.",
			Cost:         effects.ManaCost("{3}{R}{R}"),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.UntapAllCreaturesYouControl{}, effects.ExtraCombatAndMainAfterThisMain()),
		}},
	})
}
