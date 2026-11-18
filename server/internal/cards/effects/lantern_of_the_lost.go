package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lantern of the Lost — Artifact {1} (Innistrad: Crimson Vow, #5495):
//
//	"When this artifact enters, exile target card from a graveyard.
//	 {1}, {T}, Exile this artifact: Exile all cards from all graveyards,
//	 then draw a card."
//
// The enters trigger is Ashnod's Harvester's "exile target card from a
// graveyard" (any graveyard, chosen as it goes on the stack; a target that
// has gone does nothing, CR 608.2b). The ability's cost exiles the Lantern
// itself (ExileThis, CR 118.3), so the effect resolves with the Lantern
// already gone. It exiles every card in every graveyard as one batch
// (Kaya, Geist Hunter's -6) and the draw follows the exile, as "then"
// says, and it happens even when no graveyard had a card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f0c4414f-f153-40f6-a3b3-94bcbf1d3e29",
		Name:         "Lantern of the Lost",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Lantern of the Lost — exile target card from a graveyard", b27ExileChosenTarget),
				TargetCardInGraveyard("target card from a graveyard")),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Exile this artifact: Exile all cards from all graveyards, then draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(ManaCost("{1}"), TapCost(), ExileThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ids := allGraveyardsCardIDs(ctx, func(game.Card) bool { return true })
				if len(ids) == 0 {
					return DrawCards{N: 1}.Apply(ctx)
				}
				return g.ExileCardsKeyedForEffect(ids, ctx.Tail(lanternOfTheLostDraw))
			},
		}},
	})
}

// lanternOfTheLostDraw is the "then draw a card", after the exile. An
// on-disk key: never renamed, never reused.
var lanternOfTheLostDraw = CardTail("lantern-of-the-lost/draw", func(ctx *Context, _ game.TailOutcome) error {
	return DrawCards{N: 1}.Apply(ctx)
})
