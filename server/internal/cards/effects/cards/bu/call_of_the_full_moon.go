package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Call of the Full Moon — Enchantment — Aura {1}{R} (Magic Origins):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+2 and has trample.
//	 At the beginning of each upkeep, if a player cast two or more spells
//	 last turn, sacrifice this Aura."
//
// The pump and trample are the Aura's statics. The trigger is the Innistrad
// werewolves' back-face condition (one player's two spells, not two players'
// one each), checked as each upkeep begins and again as it resolves
// (CR 603.4); it sacrifices the Aura itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e34a3de-ae30-4d7b-88f1-03e2b79c4451",
		Name:         "Call of the Full Moon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 2),
			effects.GrantToAttached("trample"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return g.APlayerCastTwoOrMoreSpellsLastTurnForEffect()
			}, "Call of the Full Moon — sacrifice this Aura", func(g *game.Game, item *game.StackItem) error {
				if !g.APlayerCastTwoOrMoreSpellsLastTurnForEffect() {
					return nil // CR 603.4: the condition is checked again as it resolves
				}
				return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
