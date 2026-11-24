package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rootweaver Druid — Creature — Elf Druid {2}{G}, 2/1 (Commander Legends,
// #5507):
//
//	"When this creature enters, each opponent may search their library for
//	 up to three basic land cards. They each put one of those cards onto
//	 the battlefield tapped under your control and the rest onto the
//	 battlefield tapped under their control. Then each player who searched
//	 their library this way shuffles."
//
// Jungle Wayfinder's "each opponent may search", in turn order from the
// Druid's controller, each search the opponent's own (their library, their
// prompt, their shuffle when they search). The take is basic lands, up to
// three, onto the battlefield tapped (CR 614.12 through the entry
// pipeline). The first card the opponent picks enters under the Druid
// controller's control and the rest under the opponent's:
// SearchLibrarySpec.FirstEntersUnderControlOf (#5507). An opponent who
// declines does not shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c2e6447-c32b-4b92-a3e6-4d770e3dd250",
		Name:         "Rootweaver Druid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rootweaver Druid — each opponent may search their library for up to three basic land cards; one enters tapped under your control and the rest tapped under theirs",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, opp := range effects.OpponentsInTurnOrder(g, item.Controller) {
						if err := (effects.SearchLibrary{
							Player:        opp,
							Predicate:     effects.IsBasicLand,
							Dest:          game.ZoneBattlefield,
							Limit:         3,
							Reveal:        true,
							Optional:      true,
							Shuffle:       true,
							TappedOnEntry: true,
							Reason:        "Rootweaver Druid — up to three basic land cards; the first you pick enters under the Druid's controller's control, the rest under yours",
							FirstEnters:   item.Controller,
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
