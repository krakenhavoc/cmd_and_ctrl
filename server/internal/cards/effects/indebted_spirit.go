package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const indebtedSpiritAfterlife = "indebted-spirit/afterlife-1"

// Indebted Spirit — Enchantment Creature — Spirit {W}, 1/1 (Modern
// Horizons 3):
//
//	"Bestow {2}{W} (If you cast this card for its bestow cost, it's an Aura
//	 spell with enchant creature. It becomes a creature again if it's not
//	 attached.)
//	 Afterlife 1 (When this permanent is put into a graveyard from the
//	 battlefield, create a 1/1 white and black Spirit creature token with
//	 flying.)
//	 Enchanted creature gets +1/+1 and has afterlife 1."
//
// Bestow is the shared alternative cost (ADR 0141); the pump is
// PumpAttached. Afterlife is Afterlife Insurance's dies-trigger shape (the
// engine has no keyword token for it), carried by the card itself, so it
// fires for the Aura going to the graveyard as well as for the creature,
// and as an ADR 0093 bundle over the enchanted creature, whose own death
// makes the Spirit under its controller.
//
// No simplification.
func init() {
	const label = "Indebted Spirit — afterlife 1: create a 1/1 white and black Spirit creature token with flying"
	const text = "Afterlife 1 (When this permanent is put into a graveyard from the battlefield, create a 1/1 white and black Spirit creature token with flying.)"
	body := Do(CreateToken{Template: kayasGuileSpirit(), N: 1})
	Register(Spec{
		OracleID:         "c4c0a622-7325-4024-8b2e-6335f9a729e9",
		Name:             "Indebted Spirit",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Bestow("{2}{W}")},
		Grants: []AbilityGrant{{
			Key:       indebtedSpiritAfterlife,
			Triggered: []game.TriggeredAbility{WhenThisDies(label, body)},
			Text:      text,
		}},
		Triggered: []game.TriggeredAbility{WhenThisDies(label, body)},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantAbilitiesToAttached(indebtedSpiritAfterlife),
		},
	})
}
