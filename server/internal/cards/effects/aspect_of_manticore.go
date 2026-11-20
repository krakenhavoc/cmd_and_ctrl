package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aspect of Manticore — Enchantment — Aura {2}{R} (Theros Beyond Death, #5514):
//
//	"Flash
//	 Enchant creature
//	 When this Aura enters, enchanted creature gains first strike
//	 until end of turn.
//	 Enchanted creature gets +2/+0."
//
// Fire-Rim Form's text: flash, +2/+0 for as long as it is attached, and
// an enters trigger that gives the creature it enchants as it resolves
// first strike until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "899f3758-bb06-4f9c-81eb-7a9002f03245",
		Name:            "Aspect of Manticore",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Targets:         EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Aspect of Manticore — enchanted creature gains first strike until end of turn", aspectOfManticoreEnters),
		},
	})
}

func aspectOfManticoreEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := enchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   host,
		Keywords: []string{"first strike"},
		Label:    "Aspect of Manticore — first strike until end of turn",
	}.Apply(ctx)
}
