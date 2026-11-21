package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire Nation Palace — Land (Avatar: The Last Airbender, #4459):
//
//	"This land enters tapped unless you control a basic land.
//	 {T}: Add {R}.
//	 {1}{R}, {T}: Target creature you control gains firebending 4 until
//	 end of turn. (Whenever it attacks, add {R}{R}{R}{R}. This mana lasts
//	 until end of combat.)"
//
// Ba Sing Se's cycle: the enters-tapped replacement counts basic lands as
// the land would enter, so a basic entering at the same time does not
// count (the ruling). The activation grants a catalog bundle holding
// firebending 4 (Avatar Roku's: an attack trigger adding four red mana
// that lasts until end of combat, not a mana ability, so it uses the
// stack) until end of turn, pinned to the creature as the object it is
// now (CR 611.2c). Identical grants are not merged (CR 113.2c), so two
// activations on one creature trigger twice (the ruling).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2000fb8-39c6-4ad6-a020-5245faaa1eba",
		Name:         "Fire Nation Palace",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.EntersTappedUnless(effects.B11ControlsBasicLand)},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Grants: []effects.AbilityGrant{{
			Key: fireNationPalaceFirebending,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisAttacks("Firebending 4", effects.Do(effects.NewFirebending(4))),
			},
			Text: "Firebending 4 (Whenever this creature attacks, add {R}{R}{R}{R}. This mana lasts until end of combat.)",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, {T}: Target creature you control gains firebending 4 until end of turn.",
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "firebending")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				target, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantAbilitiesFor{
					Target: target,
					Keys:   []string{fireNationPalaceFirebending},
					Label:  "Fire Nation Palace — firebending 4 until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}

// fireNationPalaceFirebending is the granted bundle's key: an on-disk
// identity, never renamed.
const fireNationPalaceFirebending = "fire-nation-palace/firebending-4"
