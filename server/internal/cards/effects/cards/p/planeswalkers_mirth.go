package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Planeswalker's Mirth — Enchantment {2}{W} (Planeshift, S82 OLD b642, #5615):
//
//	"{3}{W}: Target opponent reveals a card at random from their hand. You gain
//	 life equal to that card's mana value."
//
// The card is picked from the game's keyed random stream (so an undo rewinds it)
// and revealed to the table (CR 701.20a). Its mana value is read as it sits in the
// hand, X as 0 (CR 202.3e). An empty hand reveals nothing and gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cebc1e6d-ab14-4e54-9906-e8a7cf4c2043",
		Name:         "Planeswalker's Mirth",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}: Target opponent reveals a card at random from their hand. You gain life equal to that card's mana value.",
			Cost:    effects.ManaCost("{3}{W}"),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				pid := effects.TargetedPlayer(ctx)
				if pid == uuid.Nil {
					return nil
				}
				p := g.PlayerByIDForEffect(pid)
				if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
					return nil
				}
				ids := make([]uuid.UUID, 0, p.Hand.Size())
				for _, c := range p.Hand.Cards {
					ids = append(ids, c.InstanceID)
				}
				pick := g.ChooseAtRandomForEffect(effects.RandomDraw(ctx), ids, 1)
				if len(pick) == 0 {
					return nil
				}
				g.RevealForEffect(game.RevealSpec{Player: pid, Source: item.SourceCardID, Cards: pick, Reason: "Planeswalker's Mirth"})
				card, ok := g.LookupCardForEffect(pick[0])
				if !ok {
					return nil
				}
				mv, ok := g.ManaValueForEffect(card)
				if !ok {
					return nil
				}
				return effects.GainLife{Player: item.Controller, Amount: mv}.Apply(ctx)
			},
		}},
	})
}
