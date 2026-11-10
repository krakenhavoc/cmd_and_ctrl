package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "d3912c82-37f8-456e-ba49-65c7f5b39d13",
		Name:            "Fae Flight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Targets:         EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fae Flight — enchanted creature gains hexproof until end of turn", faeFlightEnters),
		},
	})
}

// faeFlightEnters gives the enchanted creature hexproof until end of
// turn, or does nothing if the Aura no longer enchants one.
func faeFlightEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := enchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   host,
		Keywords: []string{"hexproof"},
		Label:    "Fae Flight — hexproof until end of turn",
	}.Apply(ctx)
}
