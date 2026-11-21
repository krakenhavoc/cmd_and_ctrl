package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rhys the Redeemed — Legendary Creature — Elf Warrior {G/W}, 1/1
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"{2}{G/W}, {T}: Create a 1/1 green and white Elf Warrior creature
//	 token.
//	 {4}{G/W}{G/W}, {T}: For each creature token you control, create a
//	 token that's a copy of that creature."
//
// The second ability is Second Harvest narrowed to creature tokens (read
// through the layered view, so an animated token counts and a
// non-creature one doesn't). The set is fixed as the ability resolves
// and the copies are one creation instruction (CR 608.2f), so they enter
// together and are not themselves copied.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7252190-ad24-4ba3-a644-2790dd1d680d",
		Name:         "Rhys the Redeemed",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}{G/W}, {T}: Create a 1/1 green and white Elf Warrior creature token.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
				Cost:    effects.Plus(effects.ManaCost("{2}{G/W}"), effects.TapCost()),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green and white Elf Warrior"), N: 1}),
			},
			{
				Label:   "{4}{G/W}{G/W}, {T}: For each creature token you control, create a token that's a copy of that creature.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    effects.Plus(effects.ManaCost("{4}{G/W}{G/W}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsToken() && c.IsCreature() {
							ids = append(ids, c.InstanceID)
						}
					}
					return effects.CreateTokenCopiesTogether{Controller: item.Controller, Copies: ids}.Apply(ctx)
				},
			},
		},
	})
}
