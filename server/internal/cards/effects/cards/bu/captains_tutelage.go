package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captain's Tutelage — {1}{R} Enchantment — Aura (Star Trek, #4973):
//
//	"Flash
//	 Enchant creature
//	 When this Aura enters, enchanted creature gains first strike until
//	 end of turn.
//	 Enchanted creature gets +2/+0."
//
// Super Speed's shape: the enters trigger names "enchanted creature" as it
// resolves, or as the Aura last was if the Aura has gone (CR 608.2h), and
// grants first strike until end of turn; the +2/+0 is the Aura's own
// static and lasts as long as it is attached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1d41aacf-e15d-4457-b12a-c71bcdb7b5de",
		Name:            "Captain's Tutelage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Static:          []game.StaticAbility{effects.PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Captain's Tutelage — enchanted creature gains first strike until end of turn", captainsTutelageEnters),
		},
	})
}

func captainsTutelageEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	host, ok := effects.EnchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return effects.GrantKeywordUntilEOT{
		Target:   host,
		Keywords: []string{"first strike"},
		Label:    "Captain's Tutelage — first strike until end of turn",
	}.Apply(ctx)
}
