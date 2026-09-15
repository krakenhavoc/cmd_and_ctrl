package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ruin Crab — Creature — Crab {U}, 0/3 (EDHREC rank 1281):
//
//	"Landfall — Whenever a land you control enters, each opponent
//	 mills three cards. (To mill a card, a player puts the top card of
//	 their library into their graveyard.)"
//
// Hedron Crab pointed at the table. The landfall trigger is the
// Tireless Provisioner shape; the mill is three per opponent, each
// through MillNForEffect, so "whenever you mill" and graveyard
// payoffs on the other side of the table see it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8afc00d4-a1c6-4329-af2c-a7f58a0c33e7",
		Name:         "Ruin Crab",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Ruin Crab — each opponent mills three (landfall)", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, opp := range ctx.Opponents() {
					if err := (MillCards{Player: opp, N: 3}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}
