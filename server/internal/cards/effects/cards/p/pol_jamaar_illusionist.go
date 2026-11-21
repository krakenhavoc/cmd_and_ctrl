package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pol Jamaar, Illusionist — Legendary Creature — Human Illusion Wizard
// {4}{U}{U}, 4/5 (Foundations Jumpstart, #5929):
//
//	"Flying
//	 When Pol Jamaar enters, choose a creature type. Draw a card for each
//	 creature you control of that type."
//
// The type is named as the trigger resolves (Selfless Safewright's choice);
// the creatures are then counted through their effective types, so a
// changeling counts for any name and Pol Jamaar himself counts when he
// shares it, and the cards are drawn for that count in one draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2f200a14-7e5b-4b7c-a283-5cf9abb1f95d",
		Name:            "Pol Jamaar, Illusionist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pol Jamaar — choose a creature type; draw a card for each creature you control of that type",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					effects.ChooseCreatureTypeTail(g, item.Controller, item.SourceCardID,
						"Pol Jamaar — choose a creature type", ctx.ItemTail(polJamaarTypeChosen))
					return nil
				}),
		},
	})
}

var polJamaarTypeChosen = effects.CreatureTypeChosen("then/bns-b52-pol-jamaar-type-chosen", func(ctx *effects.Context, t string, _ game.TailOutcome) error {
	n := effects.CountYouControl(ctx, effects.And(effects.Creature(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasSubtype(t) }))
	return effects.DrawCards{N: n}.Apply(ctx)
})
