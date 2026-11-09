package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Wilds — Plane — Eldraine (#4870):
//
//	"When you planeswalk to The Wilds and at the beginning of your upkeep,
//	 create a Food token.
//	 Whenever chaos ensues, target player sacrifices a creature of their
//	 choice. If they do, you create a Food token. You create two Food tokens
//	 instead if the sacrificed creature's toughness was 4 or greater."
//
// The sacrifice is the target player's choice and is not a target. "If they
// do" reads what really left the battlefield, and the toughness is the
// creature's last-known toughness (CR 608.2h), as Doomgape reads it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb7ae16a-9982-4782-9e20-07bb771d0601",
		Name:         "The Wilds",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("The Wilds — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
			Targeting(WheneverChaosEnsues("The Wilds — target player sacrifices a creature of their choice; if they do, you create a Food token (two if its toughness was 4 or greater)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return g.PlayerSacrificesThenKeyedForEffect(item.SourceCardID, t.ID,
						sacrificeSpec("a creature", Creature()),
						"The Wilds — sacrifice a creature", 1,
						theWildsFood, []uuid.UUID{t.ID, item.Controller})
				}),
				TargetPlayer("target player")),
		},
	})
}

// theWildsFood is the clause after the sacrifice: one Food for the planar
// controller, two when the sacrificed creature's toughness was 4 or more.
var theWildsFood = game.RegisterPromptRunThen("prompt-run/the-wilds-food", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) < 2 {
		return nil
	}
	victim, me := r.Carry[0], r.Carry[1]
	ids := r.Sacrifices().By(victim)
	if len(ids) == 0 {
		return nil
	}
	n := 1
	if departedCreatureToughness(g, ids[0]) >= 4 {
		n = 2
	}
	return CreateToken{Template: FoodToken(), N: n}.Apply(keyedContext(g, me, r.Source))
})
