package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spectral Steel — Enchantment — Aura {1}{W} (Wilds of Eldraine
// Commander, #5368):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2.
//	 {1}{W}, Exile this card from your graveyard: Return another target
//	 Aura or Equipment card from your graveyard to your hand."
//
// The graveyard ability is Bonebind Orator's: it works only from the
// graveyard (Zones), exiling this card is a cost, and "another" keeps it
// from targeting itself. The target is an Aura card or an Equipment card
// (an Aura Equipment is both).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "95f74c0d-3f89-4944-997e-1bdc408bae71",
		Name:         "Spectral Steel",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(2, 2)},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, Exile this card from your graveyard: Return another target Aura or Equipment card from your graveyard to your hand.",
			Cost:    Plus(ManaCost("{1}{W}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Targets: Another(TargetCardInGraveyard("another target Aura or Equipment card from your graveyard", YouOwn(), AnySubtype("Aura", "Equipment"))),
			Effect:  returnTargetGraveyardCardToHand,
		}},
	})
}
