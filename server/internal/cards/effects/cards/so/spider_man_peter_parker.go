package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spider-Man, Peter Parker — Legendary Creature — Spider Human Hero
// {3}{W}{W}, 4/4 (Marvel's Spider-Man, #4460):
//
//	"Flying
//	 Whenever you gain life, put a +1/+1 counter on target creature you
//	 control. It gains indestructible until end of turn."
//
// One trigger per life-gain event (CR 119.9), targeting a creature its
// controller controls. If the target is illegal on resolution nothing
// happens (CR 608.2b), so both halves go through FirstLegalBattlefieldTarget.
func init() {
	const label = "Spider-Man, Peter Parker — +1/+1 counter on target creature you control; it gains indestructible"
	effects.Register(effects.Spec{
		OracleID:        "5be3243a-6840-4601-96c8-ffb756041d49",
		Name:            "Spider-Man, Peter Parker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverYouGainLife(label, func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					target := effects.FirstLegalBattlefieldTarget(ctx)
					if target == uuid.Nil {
						return nil
					}
					if err := (effects.AddCounter{Target: target, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{
						Target:   target,
						Keywords: []string{"indestructible"},
						Label:    "Spider-Man, Peter Parker — indestructible",
					}.Apply(ctx)
				}),
				effects.TargetCreature("target creature you control", effects.YouControl()),
			),
		},
	})
}
