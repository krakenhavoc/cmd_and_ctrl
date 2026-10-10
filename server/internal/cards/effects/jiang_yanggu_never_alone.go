package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jiang Yanggu, Never Alone — Legendary Creature — Human Druid {3}{G},
// 2/2:
//
//	"When Jiang Yanggu enters, create Mowu, a legendary 3/3 green Dog
//	 creature token.
//	 At the beginning of your end step, untap all tokens you control."
//
// Mowu is a legendary token, so a second one is subject to the legend
// rule like any other. The end-step untap reads the token set as the
// trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ba661d7-1cb8-4fec-bfaa-5f2797235fcf",
		Name:         "Jiang Yanggu, Never Alone",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Jiang Yanggu, Never Alone — create Mowu, a legendary 3/3 Dog",
				Do(CreateToken{Template: TokenCard("3/3 green Legendary Dog named Mowu"), N: 1})),
			AtYourEndStep("Jiang Yanggu, Never Alone — untap all tokens you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsToken() {
							ids = append(ids, c.InstanceID)
						}
					}
					for _, id := range ids {
						if err := (UntapTarget{Target: id}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
