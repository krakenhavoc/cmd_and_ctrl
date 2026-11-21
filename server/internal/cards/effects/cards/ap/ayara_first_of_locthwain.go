package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ayara, First of Locthwain — Legendary Creature — Elf Noble
// {B}{B}{B}, 2/3 (EDHREC rank 850):
//
//	"Whenever Ayara or another black creature you control enters,
//	 each opponent loses 1 life and you gain 1 life.
//	 {T}, Sacrifice another black creature: Draw a card."
//
// Mono-black's aristocrats commander: every black body is a drain on
// the way in and a card on the way out.
//
// Colour is read from the card's computed colours (Card.HasColor:
// Scryfall's `colors`, falling back to the mana cost). A token counts
// by the colours its template records: the token table records them
// for every coloured key (#1127, pinned by a test), so a Bitterblossom
// Faerie Rogue or a Zombie both drain and feed the sacrifice.
//
// "Another" on the sacrifice cost is object identity (effects.Another,
// CR 109.1): Ayara herself is refused, which is the printed restriction.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "388168b3-ec68-4af2-b88c-6a5ec88c15f6",
		Name:         "Ayara, First of Locthwain",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return true
				}
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && c.HasColor(source.TextColor("B"))
			}, "Ayara — each opponent loses 1, you gain 1", effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice another black creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: effects.Plus(effects.TapCost(), game.AbilityCost{
				SacrificeOther: effects.Another(effects.SacrificeSpec("another black creature",
					effects.Creature(), effects.OfColor("B"))),
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
