package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lorehold Excavation — Enchantment {R}{W} (Strixhaven: School of Mages,
// #5502):
//
//	"At the beginning of your end step, mill a card. If a land card was milled
//	 this way, you gain 1 life. Otherwise, this enchantment deals 1 damage to
//	 each opponent. (To mill a card, put the top card of your library into your
//	 graveyard.)
//	 {5}, Exile a creature card from your graveyard: Create a tapped 3/2 red
//	 and white Spirit creature token."
//
// The mill's keyed tail runs once the card has landed (or at once on an empty
// library), and reads what was milled: a land card gains 1 life, anything else,
// or nothing at all, deals 1 damage from the enchantment to each opponent still
// in the game. The exiled creature card is paid at announce (CR 602.2b) and
// is another card than the source; the Spirit enters tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ed2a07c-a7a5-4910-a63c-e72803947afb",
		Name:         "Lorehold Excavation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Lorehold Excavation — mill a card; gain 1 life if it is a land card, otherwise 1 damage to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{Player: ctx.Controller(), N: 1, ThenTail: ctx.ItemTail(loreholdExcavationMilled)}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{5}, Exile a creature card from your graveyard: Create a tapped 3/2 red and white Spirit creature token.",
			Cost:  effects.Plus(effects.ManaCost("{5}"), effects.ExileFromGraveyard(1, "a creature card", game.Card.IsCreature)),
			Effect: effects.Do(effects.CreateTokenAdvanced{
				Spec: effects.Token(effects.TokenCard("3/2 red and white Spirit")).EntersTapped(), N: 1,
			}),
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
		}},
	})
}

// loreholdExcavationMilled is the clause after the mill. An on-disk key.
var loreholdExcavationMilled = effects.CardTail("old-b044/lorehold-excavation-milled", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsLand() {
			return effects.GainLife{Amount: 1}.Apply(ctx)
		}
	}
	for _, opp := range ctx.Opponents() {
		if err := (effects.DealDamage{Source: ctx.Source(), Target: opp, Amount: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
