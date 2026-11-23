package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battle Angels of Tyr — Creature — Angel Knight {2}{W}{W}, 4/4 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Flying, myriad
//	 Whenever this creature deals combat damage to a player, draw a card if that
//	 player has more cards in hand than each other player. Then you create a
//	 Treasure token if that player controls more lands than each other player.
//	 Then you gain 3 life if that player has more life than each other player."
//
// Myriad is Warchief Giant's (GrantedMyriad). The three sentences are one
// resolution read in order, "that player" being the player the damage was
// dealt to; "each other player" is every player in the game except that one,
// you included, and a player who has left the game is no longer one (CR
// 800.4a). A player ties are not "more", so a tie with any other player stops
// that sentence only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c2bd0c24-3973-48bc-a591-3779e3640c5d",
		Name:            "Battle Angels of Tyr",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Draws: 1, Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.GrantedMyriad("Battle Angels of Tyr"),
			effects.WheneverThisDealsCombatDamageToAPlayer("Battle Angels of Tyr — draw if that player has the most cards in hand, then a Treasure if they control the most lands, then gain 3 life if they have the most life",
				battleAngelsOfTyr),
		},
	})
}

func battleAngelsOfTyr(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	damaged := g.PlayerByIDForEffect(item.Trigger.Event.Target)
	if damaged == nil {
		return nil
	}
	mostOf := func(read func(p *game.Player) int) bool {
		mine := read(damaged)
		for _, p := range g.Seats {
			if p == nil || p.ID == damaged.ID || p.Eliminated {
				continue
			}
			if read(p) >= mine {
				return false
			}
		}
		return true
	}
	if mostOf(func(p *game.Player) int { return p.Hand.Size() }) {
		if err := (effects.DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if mostOf(func(p *game.Player) int { return effects.B02CountLandsControlledBy(g, p.ID) }) {
		if err := (effects.CreateToken{Controller: item.Controller, Template: effects.TreasureToken(), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if mostOf(func(p *game.Player) int { return p.Life }) {
		return effects.GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
	}
	return nil
}
