package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Illusionary Presence — Creature — Illusion {1}{U}{U}, 2/2 (Ice Age, S82 OLD
// batch 839, #5636):
//
//	"Cumulative upkeep {U}
//	 At the beginning of your upkeep, choose a land type. This creature gains
//	 landwalk of the chosen type until end of turn."
//
// The land type is chosen as the trigger resolves, among every CR 205.3i land type
// (Vision Charm's pick over game.LandTypes), and the landwalk is the parameterised
// "<type>walk" token Traveler's Cloak grants, kept until cleanup (CR 514.2). A
// creature that left before the answer is not given it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f319b31-efb0-4f48-a2a3-1f308e7b1dda",
		Name:         "Illusionary Presence",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Illusionary Presence — cumulative upkeep {U}", "{U}"),
			effects.AtYourUpkeep("Illusionary Presence — choose a land type; this creature gains landwalk of the chosen type until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ChooseBasicLandTypeTail(ctx, "Illusionary Presence — choose a land type", game.LandTypes, ctx.ItemTail(illusionaryPresenceChosen))
				}),
		},
	})
}

// illusionaryPresenceChosen gives the creature landwalk of the type chosen, until
// end of turn. An on-disk key.
var illusionaryPresenceChosen = effects.CardTail("then/illusionary-presence-landwalk", func(ctx *effects.Context, o game.TailOutcome) error {
	source := ctx.Item.SourceCardID
	token := game.LandwalkToken(o.Tail.Word)
	if token == "" || !effects.OnBattlefield(ctx.Game, source) || effects.SourceIsNewObject(ctx.Game, ctx.Item) {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: source, Keywords: []string{token}, Label: "Illusionary Presence — landwalk of the chosen type"}.Apply(ctx)
})
