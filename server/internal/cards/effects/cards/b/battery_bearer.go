package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battery Bearer — Creature — Human Artificer {2}{G}{U}, 3/4 (The
// Brothers' War):
//
//	"Creatures you control have "{T}: Add {C}. This mana can't be spent
//	 to cast a nonartifact spell."
//	 Whenever you cast an artifact spell with mana value 6 or greater,
//	 draw a card."
//
// Cryptolith Rite's layer-6 grant (ADR 0093) of Powerstone's own mana
// ability, restriction included (ManaRestrictNotNonartifactSpell). The
// ability is each creature's own, so {T} taps the creature and CR 302.6
// keeps a creature that entered this turn from using it unless it has
// haste. The mana value of a spell on the stack counts X (CR 202.3e),
// read through ManaValueForEffect.
//
// No simplification.
const batteryBearerGrant = "battery-bearer/powerstone-mana"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "de6fd93e-16d8-4855-a655-1cc0755c5d15",
		Name:         "Battery Bearer",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: batteryBearerGrant,
			Mana: []effects.ManaAbility{{
				Cost:         effects.ManaAbilityCost{Tap: true},
				Produced:     "{C}",
				Label:        "{T}: Add {C} (can't be spent to cast a nonartifact spell)",
				Restrictions: []string{effects.ManaRestrictNotNonartifactSpell},
			}},
			Text: "{T}: Add {C}. This mana can't be spent to cast a nonartifact spell.",
		}},
		Static: []game.StaticAbility{effects.Labelled(effects.GrantAbilities(effects.B16CreaturesYouControl, batteryBearerGrant),
			"Creatures you control have \"{T}: Add {C}. This mana can't be spent to cast a nonartifact spell.\"")},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.And(effects.Artifact(), effects.ManaValueGE(6)),
				"Battery Bearer — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
