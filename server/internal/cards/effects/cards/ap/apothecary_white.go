package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Apothecary White — Legendary Creature — Human Cleric {3}{W}, 3/4:
//
//	Vigilance
//	Whenever you attack, create a Food token for each player being attacked.
//	{W}, {T}, Tap X untapped Foods you control: Create X 1/1 white Human creature tokens.
//
// The activated ability is #1421's second proof: the chosen Foods
// supply X and the effect reads that announcement back at resolution.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7bf8bb41-6f46-492c-b6cf-5cfb29f973a4",
		Name:            "Apothecary White",
		Completeness:    effects.CompletenessFull,
		XMatters:        true,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Apothecary White — create a Food for each player being attacked", func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.FoodToken(), N: apothecaryWhitePlayersAttacked(g, item.Controller)}.Apply(effects.NewContext(g, item))
			})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}, Tap X untapped Foods you control: Create X 1/1 white Human creature tokens.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost(), effects.TapXUntapped("X untapped Foods you control", effects.OfSubtype("Food"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 white Human"), N: ctx.X()}.Apply(ctx)
			},
		}},
	})
}

func apothecaryWhitePlayersAttacked(g *game.Game, controller uuid.UUID) int {
	seen := map[uuid.UUID]bool{}
	for _, id := range effects.B13AttackingCreaturesYouControl(g, controller) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if defender := g.DefendingPlayerForAttackForEffect(c.AttackingTarget); defender != uuid.Nil {
			seen[defender] = true
		}
	}
	return len(seen)
}
