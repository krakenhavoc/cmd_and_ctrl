package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fae Flight — Enchantment — Aura {1}{U} (Murders at Karlov Manor,
// #4469):
//
//	"Flash
//	 Enchant creature
//	 When this Aura enters, enchanted creature gains hexproof until end
//	 of turn.
//	 Enchanted creature gets +1/+0 and has flying."
//
// Aquitect's Defenses' shape on any creature: the enters trigger gives
// the creature the Aura enchants as it resolves hexproof until end of
// turn (CR 608.2h), and the +1/+0 and flying are the Aura's statics,
// lasting as long as it stays attached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d3912c82-37f8-456e-ba49-65c7f5b39d13",
		Name:            "Fae Flight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Targets:         effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Fae Flight — enchanted creature gains hexproof until end of turn", faeFlightEnters),
		},
	})
}

// faeFlightEnters gives the enchanted creature hexproof until end of
// turn, or does nothing if the Aura no longer enchants one.
func faeFlightEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	host, ok := effects.EnchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return effects.GrantKeywordUntilEOT{
		Target:   host,
		Keywords: []string{"hexproof"},
		Label:    "Fae Flight — hexproof until end of turn",
	}.Apply(ctx)
}
