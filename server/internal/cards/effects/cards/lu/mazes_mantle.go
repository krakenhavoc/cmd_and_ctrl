package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maze's Mantle — Enchantment — Aura {2}{G}
// (Phyrexia: All Will Be One, #5479):
//
//	"Flash
//	 Enchant creature
//	 When this Aura enters, if enchanted creature has toxic, that
//	 creature gains hexproof until end of turn.
//	 Enchanted creature gets +2/+2."
//
// The enters trigger is an intervening "if" (CR 603.4): it triggers only
// if the enchanted creature has toxic as the Aura enters, and does
// nothing if it has lost it by resolution. Toxic is read through the
// layers (game.ToxicTotal).
//
// No simplification.
func init() {
	trig := effects.InterveningIf(effects.S82OneB05CondEnchantedHasToxic{},
		effects.On(game.EventETB, effects.AllOf(effects.Self, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return effects.S82OneB05AttachedHasToxic(g, source)
		}), "Maze's Mantle — enchanted creature gains hexproof until end of turn", mazesMantleEnters))
	effects.Register(effects.Spec{
		OracleID:        "b4e85b7c-3047-46d1-a6e6-2ec5eac0268a",
		Name:            "Maze's Mantle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Static:          []game.StaticAbility{effects.PumpAttached(2, 2)},
		Triggered:       []game.TriggeredAbility{trig},
	})
}

// mazesMantleEnters gives the enchanted creature hexproof until end of turn.
func mazesMantleEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	host, ok := effects.EnchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: host, Keywords: []string{"hexproof"},
		Label: "Maze's Mantle — hexproof until end of turn"}.Apply(ctx)
}
