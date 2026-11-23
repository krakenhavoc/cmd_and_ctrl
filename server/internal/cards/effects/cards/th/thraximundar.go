package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thraximundar — Legendary Creature — Zombie Assassin {4}{U}{B}{R}, 6/6
// (Alara Reborn):
//
//	"Haste
//	 Whenever Thraximundar attacks, defending player sacrifices a creature
//	 of their choice.
//	 Whenever a player sacrifices a creature, you may put a +1/+1 counter
//	 on Thraximundar."
//
// The defending player is the player attacked, or the controller of the
// planeswalker or battle attacked (CR 506.3d); they pick the creature as
// the trigger resolves (Cabal Executioner's edict) and a player with none
// sacrifices nothing. Mortician Beetle's second trigger: the sacrifice
// event fires while the permanent is still on the battlefield, so whether
// it was a creature is read as it last stood (CR 603.10a), and any
// player's sacrifice counts, the edict's included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9e0e4217-fefe-48dd-9153-032460192b19",
		Name:            "Thraximundar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.WheneverThisAttacks("Thraximundar — defending player sacrifices a creature of their choice",
				func(g *game.Game, item *game.StackItem) error {
					victim := effects.DefendingPlayerOfAttack(g, item.Trigger.Event)
					if victim == uuid.Nil || g.PlayerByIDForEffect(victim) == nil {
						return nil
					}
					g.PlayerSacrificesForEffect(item.SourceCardID, victim, effects.SacrificeSpec("a creature", effects.Creature()),
						"Thraximundar — sacrifice a creature")
					return nil
				})),
			effects.Optional(effects.On(game.EventSacrifice, effects.RoeB1APlayerSacrificedACreature,
				"Thraximundar — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
				"Thraximundar — put a +1/+1 counter on it?"),
		},
	})
}
