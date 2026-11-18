package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knights' Charge — Enchantment {1}{W}{B} (Magic Commander Collection, #5908):
//
//	"Whenever a Knight you control attacks, each opponent loses 1 life and
//	 you gain 1 life.
//	 {6}{W}{B}, Sacrifice this enchantment: Return all Knight creature cards
//	 from your graveyard to the battlefield."
//
// The trigger is Sanctum Seeker's, once per attacking Knight. The sacrifice
// is a cost; the return moves every Knight creature card in your graveyard
// together (CR 400.7: they enter at once).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b4014b7c-aa1d-4feb-ab7a-4921ad150f9a",
		Name:         "Knights' Charge",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b21CreatureOfSubtypeYouControlAttacked(ev, source, g, "Knight")
			}, "Knights' Charge — each opponent loses 1 life, you gain 1 life", Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{6}{W}{B}, Sacrifice this enchantment: Return all Knight creature cards from your graveyard to the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{6}{W}{B}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return ReturnFromGraveyardTogether{
					Targets: graveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
						return c.IsCreature() && c.HasSubtype(ctx.TextSubtype("Knight"))
					}),
					Controller: ctx.Controller(),
				}.Apply(ctx)
			},
		}},
	})
}
