package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Casal, Lurkwood Pathfinder // Casal, Pathbreaker Owlbear — Legendary
// Creature — Tiefling Druid {3}{G}, 3/3 // Legendary Creature — Bird Bear, 6/6
// (Commander Legends: Battle for Baldur's Gate Commander, Secret Lair Drop):
//
//	Front: "Vigilance
//	        When Casal enters, search your library for a Forest card, put it
//	        onto the battlefield tapped, then shuffle.
//	        Whenever Casal attacks, you may pay {1}{G}. If you do, transform
//	        her."
//	Back:  "Vigilance, trample
//	        When this creature transforms into Casal, Pathbreaker Owlbear,
//	        other legendary creatures you control get +2/+2 and gain trample
//	        until end of turn.
//	        At the beginning of your upkeep, transform Casal."
//
// The attack payment is asked as the trigger resolves, and the transform is
// ignored if Casal has already turned over since the trigger went on the stack
// (CR 701.27f). The back face's "transforms into" trigger affects the legendary
// creatures other than Casal that its controller controls as it resolves (the
// set is locked then, CR 611.2c), for the turn.
//
// No simplification.
func init() {
	const oracle = "b7bdb688-f8ee-4d22-a679-37a2ecd390a1"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Casal, Lurkwood Pathfinder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Purpose:         game.Purpose{Tutors: 1, Lands: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Casal, Lurkwood Pathfinder — search for a Forest card, put it onto the battlefield tapped, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player: item.Controller,
						Predicate: func(c game.Card) bool {
							return c.IsLand() && c.HasSubtypeWord("Forest")
						},
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        "Casal — a Forest card",
					}.Apply(effects.NewContext(g, item))
				}),
			effects.WheneverThisAttacks("Casal, Lurkwood Pathfinder — you may pay {1}{G}; if you do, transform her",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{G}",
						Question: "Casal — pay {1}{G} to transform her?",
						Then:     casalTransform,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Casal, Pathbreaker Owlbear",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventTransform, effects.ThisTransformedInto(1),
				"Casal, Pathbreaker Owlbear — other legendary creatures you control get +2/+2 and gain trample until end of turn",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					casal := ctx.Source()
					match := effects.And(effects.Creature(), effects.YouControl(), effects.Legendary(),
						func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != casal })
					if err := (effects.BoostUntilEOT{Match: match, Power: 2, Toughness: 2,
						Label: "Casal, Pathbreaker Owlbear — +2/+2"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: match, Keywords: []string{"trample"},
						Label: "Casal, Pathbreaker Owlbear — trample"}.Apply(ctx)
				}))),
			effects.AtYourUpkeep("Casal, Pathbreaker Owlbear — transform Casal", effects.Do(effects.TransformThis{})),
		},
	})
}

// casalTransform is the paid half of the attack trigger: transform her. An
// on-disk key: never renamed, never reused.
var casalTransform = effects.PayUnlessBranches("pay-unless/casal-lurkwood-pathfinder-transform",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error { return effects.TransformThis{}.Apply(ctx) }, nil)
