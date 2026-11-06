package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aquitect's Defenses — Enchantment — Aura {1}{U} (Lorwyn Eclipsed,
// #4458):
//
//	"Flash
//	 Enchant creature you control
//	 When this Aura enters, enchanted creature gains hexproof until end of
//	 turn. (It can't be the target of spells or abilities your opponents
//	 control.)
//	 Enchanted creature gets +1/+2."
//
// Accelerated Evolution's shape at +1/+2: a flash Aura on a creature you
// control, a static pump on the host, and an enters trigger that gives the
// creature the Aura enchants as it resolves hexproof until end of turn
// (CR 608.2h). An Aura that has left by then gives nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b9dee727-8ad8-42e0-93c6-5ef91d3f7309",
		Name:            "Aquitect's Defenses",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Targets:         EnchantCreature(YouControl()),
		Static: []game.StaticAbility{
			PumpAttached(1, 2),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Aquitect's Defenses — enchanted creature gains hexproof until end of turn", aquitectsDefensesEnters),
		},
	})
}

// aquitectsDefensesEnters gives the enchanted creature hexproof until end
// of turn, or does nothing if the Aura no longer enchants one.
func aquitectsDefensesEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := enchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   host,
		Keywords: []string{"hexproof"},
		Label:    "Aquitect's Defenses — hexproof until end of turn",
	}.Apply(ctx)
}
