package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Klothys, God of Destiny — Legendary Enchantment Creature — God {1}{R}{G}, 4/5:
//
//	"Indestructible
//	 As long as your devotion to red and green is less than seven, Klothys
//	 isn't a creature.
//	 At the beginning of your first main phase, exile target card from a
//	 graveyard. If it was a land card, add {R} or {G}. Otherwise, you gain 2
//	 life and Klothys deals 2 damage to each opponent."
//
// "First main phase" is the precombat main phase. The card's type is read
// from the exile zone, after it has been exiled; a card that did not reach
// exile (a replacement sent it elsewhere) does nothing more.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e8473969-195f-41ee-a84d-463cc8c0ef02",
		Name:            "Klothys, God of Destiny",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static:          []game.StaticAbility{effects.GodUnlessDevotionToColors(7, "R", "G")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtYourPrecombatMain("Klothys, God of Destiny — exile target card from a graveyard; a land makes {R} or {G}, anything else gains 2 life and deals 2 damage to each opponent",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.ExileTarget{Target: id, ThenTail: ctx.ItemTail(klothysExiled)}.Apply(ctx)
					}),
				effects.TargetCardInGraveyard("target card from a graveyard"),
			),
		},
	})
}

// klothysExiled is the clause after the exile, told the card that reached
// exile. An on-disk key: never renamed, never reused.
var klothysExiled = effects.CardTail("klothys/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	card, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	if card.IsLand() {
		return effects.AddMana{Amount: effects.ManaN(1, "R", "G")}.Apply(ctx)
	}
	if err := (effects.GainLife{Amount: 2}).Apply(ctx); err != nil {
		return err
	}
	return effects.DamageEachOpponent{N: 2}.Apply(ctx)
})
