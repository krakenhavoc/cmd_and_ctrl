package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Startling Development — Instant {1}{U} (Jumpstart 2022, #5948):
//
//	"Until end of turn, target creature becomes a blue Serpent with base
//	 power and toughness 4/4.
//	 Cycling {1}"
//
// Turn to Frog's shape without losing abilities: the creature keeps its
// card types and every ability, replaces its creature types with Serpent
// (CR 205.1b), its colours with blue (CR 105.3) and its base power and
// toughness with 4/4 (layer 7b). Cycling {1} from hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8747ca90-ac1d-45fd-89fe-aac385cd66b4",
		Name:         "Startling Development",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{1}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id := effects.FirstLegalBattlefieldTarget(ctx)
			if id == [16]byte{} {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, id, nil, "Startling Development — becomes a blue Serpent with base power and toughness 4/4",
				game.LoseAllCreatureTypesMod(),
				game.AddSubtypesMod(ctx.TextSubtypes("Serpent")...),
				game.SetColorsMod(ctx.TextColor("U")),
				game.SetBasePowerMod(4),
				game.SetBaseToughnessMod(4),
			)
		},
	})
}
