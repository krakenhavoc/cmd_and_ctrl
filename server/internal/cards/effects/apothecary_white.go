package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "7bf8bb41-6f46-492c-b6cf-5cfb29f973a4",
		Name:            "Apothecary White",
		Completeness:    CompletenessFull,
		XMatters:        true,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Apothecary White — create a Food for each player being attacked", func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: FoodToken(), N: apothecaryWhitePlayersAttacked(g, item.Controller)}.Apply(NewContext(g, item))
			})),
		},
		Activated: []ActivatedAbility{{
			Label: "{W}, {T}, Tap X untapped Foods you control: Create X 1/1 white Human creature tokens.",
			Cost:  Plus(ManaCost("{W}"), TapCost(), TapXUntapped("X untapped Foods you control", OfSubtype("Food"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white Human"), N: ctx.X()}.Apply(ctx)
			},
		}},
	})
}

func apothecaryWhitePlayersAttacked(g *game.Game, controller uuid.UUID) int {
	seen := map[uuid.UUID]bool{}
	for _, id := range b13AttackingCreaturesYouControl(g, controller) {
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
