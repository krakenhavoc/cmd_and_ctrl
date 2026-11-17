package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	trig := InterveningIf(s82OneB05CondEnchantedHasToxic{},
		On(game.EventETB, AllOf(Self, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return s82OneB05AttachedHasToxic(g, source)
		}), "Maze's Mantle — enchanted creature gains hexproof until end of turn", acceleratedEvolutionEnters))
	Register(Spec{
		OracleID:        "b4e85b7c-3047-46d1-a6e6-2ec5eac0268a",
		Name:            "Maze's Mantle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Static:          []game.StaticAbility{PumpAttached(2, 2)},
		Triggered:       []game.TriggeredAbility{trig},
	})
}
