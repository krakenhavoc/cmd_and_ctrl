package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crossover Collaboration — {2}{R} Instant:
//
//	"Teamwork 2 (As an additional cost to cast this spell, you may tap
//	 any number of creatures you control with total power 2 or more.)
//	 Exile the top two cards of your library. Until the end of your
//	 next turn, you may play those cards. If this spell was cast using
//	 teamwork, create a Treasure token. (It's an artifact with '{T},
//	 Sacrifice this token: Add one mana of any color.')"
//
// #1703: Teamwork(2). The impulse exile is
// ExileTopNUntilYourNextTurn — "play", not "cast", so an exiled land
// is playable, unlike Ragavan's CastOnly grant. No simplification.
func init() {
	Register(Spec{
		OracleID:      "150cefd2-061d-43b2-8470-c17be0144021",
		Name:          "Crossover Collaboration",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Teamwork(2)},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := ExileTopNUntilYourNextTurn(ctx, 2); err != nil {
				return err
			}
			if !ctx.UsedTeamwork() {
				return nil
			}
			return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(ctx)
		},
	})
}
