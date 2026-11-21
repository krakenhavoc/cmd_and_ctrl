package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Obsessive Stitcher — {1}{U}{B} Creature — Human Wizard 0/3:
//
//	"{T}: Draw a card, then discard a card.
//	 {2}{U}{B}, {T}, Sacrifice this creature: Return target creature
//	 card from your graveyard to the battlefield."
//
// A rummaging tap ability beside Doomed Necromancer's three-part cost
// (mana, tap and sacrifice-this all paid on activation, so killing the
// Stitcher in response does nothing). "From your graveyard" means the
// creature returns to its owner, the activator, and never from an
// opponent's pile. Both abilities have {T} in the cost, so summoning
// sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e8210145-a6ac-4f84-a74c-a5799c5e9434",
		Name:         "Obsessive Stitcher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Draw a card, then discard a card.",
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DrawThenDiscard(g, item.Controller, item.SourceCardID, 1, 1, "")
				},
			},
			{
				Label:   "{2}{U}{B}, {T}, Sacrifice this creature: Return target creature card from your graveyard to the battlefield.",
				Cost:    effects.Plus(effects.ManaCost("{2}{U}{B}"), effects.TapCost(), effects.SacrificeThis()),
				Targets: effects.TargetCreatureInYourGraveyard(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					_, _ = effects.ReanimateSingleTarget(ctx, item.Controller)
					return nil
				},
			},
		},
	})
}
