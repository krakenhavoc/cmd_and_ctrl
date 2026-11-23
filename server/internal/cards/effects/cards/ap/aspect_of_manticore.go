package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "899f3758-bb06-4f9c-81eb-7a9002f03245",
		Name:            "Aspect of Manticore",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Targets:         effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Aspect of Manticore — enchanted creature gains first strike until end of turn", aspectOfManticoreEnters),
		},
	})
}

func aspectOfManticoreEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	host, ok := effects.EnchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return effects.GrantKeywordUntilEOT{
		Target:   host,
		Keywords: []string{"first strike"},
		Label:    "Aspect of Manticore — first strike until end of turn",
	}.Apply(ctx)
}
