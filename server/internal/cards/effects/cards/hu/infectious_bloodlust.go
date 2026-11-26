package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infectious Bloodlust — Enchantment — Aura {1}{R} (Magic Origins):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+1, has haste, and attacks each combat if
//	 able.
//	 When enchanted creature dies, you may search your library for a card
//	 named Infectious Bloodlust, reveal it, put it into your hand, then
//	 shuffle."
//
// The statics are the Aura's (the attack requirement belongs to the Aura and
// ends with it). The dies trigger is harvested while the Aura is still
// attached (Angelic Destiny's shape); "you" is the Aura's controller, and the
// search is optional, by name, to hand, revealed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9cd8ca01-e8a9-47ea-8899-b016a2dcaea8",
		Name:         "Infectious Bloodlust",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 1),
			effects.GrantToAttached("haste"),
			effects.AttacksEachCombatWhere(effects.AttachedToSource),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Infectious Bloodlust — you may search your library for a card named Infectious Bloodlust",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return game.HasName(c, "Infectious Bloodlust") },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Source:    item.SourceCardID,
						Reason:    "Infectious Bloodlust — search for a card named Infectious Bloodlust",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
