package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Detective's Phoenix — Enchantment Creature — Phoenix {2}{R}, 2/2 (Murders
// at Karlov Manor, #5948):
//
//	"Bestow—{R}, Collect evidence 6. (To pay this bestow cost, pay {R} and
//	 exile cards with total mana value 6 or greater from your graveyard.)
//	 Flying, haste
//	 Enchanted creature gets +2/+2 and has flying and haste.
//	 You may cast this card from your graveyard using its bestow ability."
//
// Bestow is the shared alternative cost (ADR 0141), here with a collect
// evidence half (BestowCollectEvidence): the caster names graveyard cards
// with total mana value 6 or more on alt_cost_ids, exiled with the spell on
// the stack, and paying it is collecting evidence (CR 701.59a). The
// graveyard permission is the same offer bound to the graveyard, keyed
// BestowGraveyardKey: a card that opens a zone and prices it must be cast
// from it for that price (CR 601.2, the cast path's rule 3), so the card
// can leave the graveyard only bestowed. It is on the stack when the cost
// is paid (CR 601.2a), so it can't exile itself as evidence.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "52c3fe55-74ee-44cb-b62b-dc6c5dfdc33e",
		Name:            "Detective's Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		CastableZones:   []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{
			effects.BestowCollectEvidence("{R}", 6, ""),
			effects.BestowCollectEvidence("{R}", 6, game.ZoneGraveyard),
		},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("flying", "haste"),
		},
	})
}
