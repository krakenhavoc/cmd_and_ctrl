package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Earth Rumble — Sorcery {3}{G} (Avatar: The Last Airbender, #4459):
//
//	"Earthbend 2. When you do, up to one target creature you control
//	 fights target creature an opponent controls."
//
// The earthbend is the keyword action on the spell's own target land. If
// that land is no longer a legal target as the spell resolves, the spell
// doesn't resolve (CR 608.2b, the ruling) and nothing is earthbent, so no
// "when you do" trigger is made. Otherwise the fight is a reflexive
// triggered ability (CR 603.12): it goes on the stack after the spell
// has resolved and its two targets are chosen then, so the land just
// earthbent can be the creature you control. It is a fight between the
// two still-legal picks (fightTargetClauses); with no creature of yours
// chosen, or either target illegal as it resolves, neither creature
// deals or is dealt damage (CR 701.14b, the ruling).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f04a7763-5fba-42e5-a45a-b3dc2daba902",
		Name:         "Earth Rumble",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EarthbendTargets(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			land := effects.FirstLegalBattlefieldTarget(ctx)
			if land == uuid.Nil {
				return nil
			}
			if err := (effects.Earthbend{Target: land, N: 2}).Apply(ctx); err != nil {
				return err
			}
			return effects.WhenYouDo("Earth Rumble — up to one target creature you control fights target creature an opponent controls",
				earthRumbleFightBody).Apply(ctx)
		},
	})
}

// earthRumbleFightBody is the reflexive fight. An on-disk key.
var earthRumbleFightBody = game.ReflexiveBody("earth-rumble/fight",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		return effects.FightTargetClauses(effects.NewContext(g, item))
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.Clauses(
			effects.TargetCreature("up to one target creature you control", effects.YouControl()).WithCount(0, 1),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()))
	}))
