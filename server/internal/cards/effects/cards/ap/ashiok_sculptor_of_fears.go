package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashiok, Sculptor of Fears — Legendary Planeswalker — Ashiok {4}{U}{B},
// loyalty 4 (Theros Beyond Death, #5514):
//
//	"+2: Draw a card. Each player mills two cards.
//	 −5: Put target creature card from a graveyard onto the battlefield under
//	 your control.
//	 −11: Gain control of all creatures target opponent controls."
//
// The +2 draws, then every player mills two as a second instruction of the
// same ability (Extract from Darkness's mill). The −5 targets a creature card
// in any graveyard and puts it onto the battlefield under the activator's
// control; a target that left in response does nothing (CR 608.2b). The −11
// takes every creature the target opponent controls as the ability resolves,
// the set locked then (CR 611.2c), with no stated duration (CR 611.2a): the
// theft lasts until the game ends or something else changes control, and a
// creature that arrives later is not taken.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ce79bbd2-edf0-4d5e-9aaf-717ec3960c48",
		Name:            "Ashiok, Sculptor of Fears",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Draw a card. Each player mills two cards.",
				Cost:    effects.LoyaltyCost(2),
				Purpose: game.Purpose{Draws: 1},
				Effect: effects.Do(
					effects.DrawCards{N: 1},
					effects.Instruction(func(ctx *effects.Context) error {
						return effects.B02EachPlayerMills(ctx.Game, ctx.Item, 2)
					}),
				),
			},
			{
				Label:   "−5: Put target creature card from a graveyard onto the battlefield under your control.",
				Cost:    effects.LoyaltyCost(-5),
				Targets: effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
				Effect:  effects.S82Mom16PutTargetOntoTheBattlefield,
			},
			{
				Label:   "−11: Gain control of all creatures target opponent controls.",
				Cost:    effects.LoyaltyCost(-11),
				Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					victim := t.ID
					taken := effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.Controller == victim
					}))
					for _, c := range taken {
						if err := (effects.GainControl{
							Target:     c.InstanceID,
							Controller: ctx.Controller(),
							Duration:   game.IndefiniteDuration(),
							Label:      "Ashiok, Sculptor of Fears — gain control (no stated duration)",
						}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})
}
