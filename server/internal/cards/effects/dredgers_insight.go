package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3f21c1c6-7992-4036-ad3e-b599153e54fe",
		Name:         "Dredger's Insight",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny(leftGraveyardEvents,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return artifactOrCreatureCardLeftYourGraveyard(ev, source, g)
				},
				"Dredger's Insight — you gain 1 life",
				Do(GainLife{Amount: 1}))),
			WhenThisEnters("Dredger's Insight — mill four cards, then you may put an artifact, creature, or land card from among them into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{N: 4, ThenTail: ctx.ItemTail(dredgersInsightMilled)}.Apply(ctx)
				}),
		},
	})
}

var dredgersInsightMilled = CardTail("dredgers-insight/after-mill", func(ctx *Context, o game.TailOutcome) error {
	return mayTakeOneFromAmongThem(ctx, o.IDs, Or(Artifact(), Creature(), Land()),
		"Dredger's Insight — you may put an artifact, creature, or land card from among the milled cards into your hand")
})
