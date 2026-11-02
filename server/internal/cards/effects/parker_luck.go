package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Parker Luck — Enchantment {2}{B} (Marvel's Spider-Man promos, #4460):
//
//	"At the beginning of your end step, two target players each reveal the
//	 top card of their library. They each lose life equal to the mana value
//	 of the card revealed by the other player. Then they each put the card
//	 they revealed into their hand."
//
// Two different players are targeted (you may be one of them). As the
// trigger resolves each still-legal one reveals the top card of their
// library; each then loses life equal to the mana value of the card the
// OTHER revealed. A player who is no longer a legal target reveals nothing
// (CR 608.2b), so the other loses nothing, and a player with an empty
// library reveals nothing either. Last, each revealed card still on top of
// its library goes to its owner's hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "03d1a4e9-3ebc-491f-8741-5d933b500abf",
		Name:         "Parker Luck",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventBeginEndStep, ByYou, parkerLuckLabel, parkerLuckResolve),
				TargetPlayer("two target players").WithCount(2, 2)),
		},
	})
}

const parkerLuckLabel = "Parker Luck — two target players each reveal their top card and lose life equal to the other's mana value"

func parkerLuckResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var players []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			players = append(players, t.ID)
		}
	}
	revealed := make([]uuid.UUID, len(players))
	value := make([]int, len(players))
	for i, p := range players {
		ids := g.RevealTopOfLibraryForEffect(p, ctx.Source(), 1, parkerLuckLabel)
		if len(ids) == 0 {
			continue
		}
		revealed[i] = ids[0]
		if c, ok := g.LookupCardForEffect(ids[0]); ok {
			value[i] = c.ManaValue()
		}
	}
	if len(players) == 2 {
		for i, p := range players {
			if loss := value[1-i]; loss > 0 {
				if err := g.ChangePlayerLifeForEffect(ctx.Source(), p, -loss); err != nil {
					return err
				}
			}
		}
	}
	for i, p := range players {
		if revealed[i] != uuid.Nil {
			g.TakeFromLibraryToHandForEffect(p, []uuid.UUID{revealed[i]})
		}
	}
	return nil
}
