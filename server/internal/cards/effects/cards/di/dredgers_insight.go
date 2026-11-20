package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dredger's Insight — {1}{G} Enchantment (Aetherdrift, #4464):
//
//	"Whenever one or more artifact and/or creature cards leave your
//	 graveyard, you gain 1 life.
//	 When this enchantment enters, mill four cards. You may put an
//	 artifact, creature, or land card from among the milled cards into
//	 your hand."
//
// The first ability is Desecrated Tomb's batch trigger (OncePerBatch,
// CR 603.2c) over "artifact and/or creature" cards, in any direction and
// by any route (left_graveyard.go): the card put into the hand by the
// second ability leaves the graveyard too, and gains the life. The mill is
// Eerie Gravestone's: the pick is made from among the cards the mill
// really put into the graveyard, as the ability resolves (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f21c1c6-7992-4036-ad3e-b599153e54fe",
		Name:         "Dredger's Insight",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.OnAny(effects.LeftGraveyardEvents,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.ArtifactOrCreatureCardLeftYourGraveyard(ev, source, g)
				},
				"Dredger's Insight — you gain 1 life",
				effects.Do(effects.GainLife{Amount: 1}))),
			effects.WhenThisEnters("Dredger's Insight — mill four cards, then you may put an artifact, creature, or land card from among them into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{N: 4, ThenTail: ctx.ItemTail(dredgersInsightMilled)}.Apply(ctx)
				}),
		},
	})
}

var dredgersInsightMilled = effects.CardTail("dredgers-insight/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.MayTakeOneFromAmongThem(ctx, o.IDs, effects.Or(effects.Artifact(), effects.Creature(), effects.Land()),
		"Dredger's Insight — you may put an artifact, creature, or land card from among the milled cards into your hand")
})
