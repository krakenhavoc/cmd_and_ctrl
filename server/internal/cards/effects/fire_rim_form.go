package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fire-Rim Form — Enchantment — Aura {1}{R} (Tarkir: Dragonstorm, #4463):
//
//	"Flash
//	 Enchant creature
//	 When this Aura enters, enchanted creature gains first strike until
//	 end of turn.
//	 Enchanted creature gets +2/+0."
//
// Super Speed's shape: the enters trigger names "enchanted creature" as
// it resolves and grants first strike until end of turn; the +2/+0 is the
// Aura's own static and lasts as long as it is attached.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "89fb21dc-4cf2-4c9a-b0ae-cc6e10277fb6",
		Name:            "Fire-Rim Form",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Targets:         EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fire-Rim Form — enchanted creature gains first strike until end of turn", fireRimFormEnters),
		},
	})
}

// fireRimFormEnters gives the enchanted creature first strike until end
// of turn.
func fireRimFormEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := enchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   host,
		Keywords: []string{"first strike"},
		Label:    "Fire-Rim Form — first strike until end of turn",
	}.Apply(ctx)
}
