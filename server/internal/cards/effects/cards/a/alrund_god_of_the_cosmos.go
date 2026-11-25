package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alrund, God of the Cosmos // Hakka, Whispering Raven — modal double-faced
// card, {3}{U}{U} // {1}{U} (Kaldheim).
//
// Front, Legendary Creature — God, 1/1:
//
//	"Alrund gets +1/+1 for each card in your hand and each foretold card you
//	 own in exile.
//	 At the beginning of your end step, choose a card type, then reveal the top
//	 two cards of your library. Put all cards of the chosen type revealed this
//	 way into your hand and the rest on the bottom of your library in any
//	 order."
//
// The bonus is a layer 7c per-count static read from the board on every pass:
// the cards in the controller's hand plus the face-down foretold cards they own
// in exile. The end step trigger asks the type first (the nine card types
// that appear on cards), then reveals two, takes every card of that type and
// orders the rest on the bottom.
//
// Back, Legendary Creature — Bird, 2/3 (registered under "<oracle_id>#1",
// ADR 0034):
//
//	"Flying
//	 Whenever Hakka deals combat damage to a player, return it to its owner's
//	 hand, then scry 2."
//
// The return is Hakka's own bounce; the scry happens whether or not it reached
// the hand, since "then" is a separate instruction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a69c3dd-52a3-4ef6-b53f-b0780e0dc5f8",
		Name:         "Alrund, God of the Cosmos",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.S82NeoB18SelfPT("Alrund gets +1/+1 for each card in your hand and each foretold card you own in exile.",
				func(g *game.Game, source *game.Card) int {
					n := 0
					if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Hand != nil {
						n = p.Hand.Size()
					}
					if g.Exile != nil {
						for _, c := range g.Exile.Cards {
							if c.Owner == source.Controller && game.CardIsForetold(c) {
								n++
							}
						}
					}
					return n
				}),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Alrund, God of the Cosmos — choose a card type, then reveal the top two cards of your library; put all cards of the chosen type into your hand and the rest on the bottom in any order",
				func(g *game.Game, item *game.StackItem) error {
					options := make([]game.ChoiceOption, len(effects.S82OldB064AlrundCardTypes))
					for i, t := range effects.S82OldB064AlrundCardTypes {
						options[i] = game.ChoiceOption{Label: t}
					}
					g.QueueOptionPickForEffect(game.OptionPickPrompt{
						Chooser:  item.Controller,
						Source:   item.SourceCardID,
						Question: "Alrund, God of the Cosmos — choose a card type",
						Options:  options,
						ThenKey:  effects.S82OldB064AlrundTypeChosen,
					})
					return nil
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        "0a69c3dd-52a3-4ef6-b53f-b0780e0dc5f8#1",
		Name:            "Hakka, Whispering Raven",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Scry: 2},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Hakka, Whispering Raven — return it to its owner's hand, then scry 2",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if !effects.SourceIsStillThisPermanent(g, item) {
						return effects.Scry{Player: item.Controller, N: 2}.Apply(ctx)
					}
					return effects.BounceToHand{Target: item.SourceCardID, ThenTail: ctx.ItemTail(effects.S82OldB064HakkaScry)}.Apply(ctx)
				}),
		},
	})
}
