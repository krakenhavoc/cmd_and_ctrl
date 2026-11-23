package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desmond Miles — Legendary Creature — Human Assassin {1}{B}, 1/3
// (Assassin's Creed, #5948):
//
//	"Menace
//	 Desmond Miles gets +1/+0 for each other Assassin you control and each
//	 Assassin card in your graveyard.
//	 Whenever Desmond Miles deals combat damage to a player, surveil X, where
//	 X is the amount of damage it dealt to that player."
//
// The static is a layer 7c count read as the layers are computed (Boneclub
// Berserker's shape): the other Assassins its controller controls (never
// Desmond himself) plus the Assassin cards in their graveyard, so a creature
// dying changes the power on the next pass. The trigger carries the damage
// dealt to the player on the item and surveils that many cards; damage
// prevented to zero triggers nothing (CR 120.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "654d84bb-f801-4d03-898c-78952f1c5a73",
		Name:            "Desmond Miles",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{{
			Label:     "Desmond Miles gets +1/+0 for each other Assassin you control and each Assassin card in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				assassin := source.TextSubtype("Assassin")
				n := effects.CountPermanentsYouControl(g, source, func(other *game.Card) bool {
					return other.InstanceID != source.InstanceID && other.IsCreature() && other.HasSubtype(assassin)
				})
				if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Graveyard != nil {
					for _, gc := range p.Graveyard.Cards {
						if gc.HasSubtype(assassin) {
							n++
						}
					}
				}
				c.Power += n
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: effects.ThisDealtCombatDamageToAPlayer,
			Key:       "Desmond Miles — surveil X, where X is the damage dealt",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Desmond Miles — surveil X, where X is the damage dealt")
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Params.Amount <= 0 {
					return nil
				}
				return effects.Surveil{Player: item.Controller, N: item.Params.Amount}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
