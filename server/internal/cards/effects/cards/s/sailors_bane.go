package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sailors' Bane — Creature — Dragon Turtle {7}{U}{U}, 7/7 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"This spell costs {1} less to cast for each card you own in exile and in
//	 your graveyard that's an instant card, a sorcery card, or a card that has
//	 an Adventure.
//	 Ward {4}"
//
// The reduction is the Dawning Archaic's self cost modifier over two zones:
// the controller's own graveyard and the exile cards they OWN (a card owned by
// someone else does not count), each an instant, a sorcery, or an Adventure
// card (layout adventure, which a card keeps in every zone). The spell being
// cast is never counted. Ward {4} is the engine's ward.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5d22ab00-b454-4407-93d7-65b47f56cbaa",
		Name:         "Sailors' Bane",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(sailorsBaneCards,
				"This spell costs {1} less to cast for each card you own in exile and in your graveyard that's an instant card, a sorcery card, or a card that has an Adventure."),
		},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{4}"), "Sailors' Bane — ward {4}"),
		},
	})
}

func sailorsBaneCards(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	count := func(cards []game.Card) int {
		n := 0
		for _, c := range cards {
			if c.InstanceID == q.Card.InstanceID || c.Owner != q.Controller {
				continue
			}
			if c.IsInstant() || c.IsSorcery() || c.Layout == game.LayoutAdventure {
				n++
			}
		}
		return n
	}
	n := 0
	if p := q.Game.PlayerByIDForEffect(q.Controller); p != nil && p.Graveyard != nil {
		n += count(p.Graveyard.Cards)
	}
	if q.Game.Exile != nil {
		n += count(q.Game.Exile.Cards)
	}
	return n
}
