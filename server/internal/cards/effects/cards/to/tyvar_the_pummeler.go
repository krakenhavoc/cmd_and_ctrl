package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tyvar, the Pummeler — Legendary Creature — Elf Warrior {1}{G}{G}, 3/3
// (Duskmourn: House of Horror, #4466):
//
//	"Tap another untapped creature you control: Tyvar gains indestructible
//	 until end of turn. Tap it.
//	 {3}{G}{G}: Creatures you control get +X/+X until end of turn, where X
//	 is the greatest power among creatures you control."
//
// The first ability is Vanguard of the Rose's "gains indestructible until
// end of turn. Tap it." with Ranger's Hawk's cost (TapAnotherUntapped):
// tapping another creature is not a {T} cost, so a creature that came
// under your control this turn can pay it (CR 302.6), and so can Tyvar
// when it is itself summoning sick. "It" is Tyvar, which taps as the
// ability resolves if it is still the same permanent.
//
// The second is Night of the Sweets' Revenge's team pump with Orcish
// Siegemaster's X (b42GreatestPowerControlledBy): the greatest power among
// creatures you control, read once as the ability resolves (CR 608.2h),
// and every creature you control then gets +X/+X until end of turn
// (CR 611.2c), so a creature that arrives later gets nothing. X is never
// less than 0 (CR 107.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a242eab-204c-4411-860f-c0f3582a8f38",
		Name:         "Tyvar, the Pummeler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Tap another untapped creature you control: Tyvar gains indestructible until end of turn. Tap it.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.TapAnotherUntapped("another untapped creature you control", effects.Creature()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.B09SourceStillOnBattlefield(g, item) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					if err := (effects.GrantKeywordUntilEOT{
						Target:   item.SourceCardID,
						Keywords: []string{"indestructible"},
						Label:    "Tyvar, the Pummeler — indestructible",
					}).Apply(ctx); err != nil {
						return err
					}
					return effects.TapTarget{Target: item.SourceCardID}.Apply(ctx)
				},
			},
			{
				Label:   "{3}{G}{G}: Creatures you control get +X/+X until end of turn, where X is the greatest power among creatures you control.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{3}{G}{G}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					x := effects.B42GreatestPowerControlledBy(g, item.Controller)
					if x <= 0 {
						return nil
					}
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.Creature(), effects.ControlledBy(item.Controller)),
						Power:     x,
						Toughness: x,
						Label:     "Tyvar, the Pummeler — creatures you control get +X/+X",
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
