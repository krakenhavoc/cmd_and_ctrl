package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "f04a7763-5fba-42e5-a45a-b3dc2daba902",
		Name:         "Earth Rumble",
		Completeness: CompletenessFull,
		Targets:      EarthbendTargets(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			land := FirstLegalBattlefieldTarget(ctx)
			if land == uuid.Nil {
				return nil
			}
			if err := (Earthbend{Target: land, N: 2}).Apply(ctx); err != nil {
				return err
			}
			return WhenYouDo("Earth Rumble — up to one target creature you control fights target creature an opponent controls",
				earthRumbleFightBody).Apply(ctx)
		},
	})
}

// earthRumbleFightBody is the reflexive fight. An on-disk key.
var earthRumbleFightBody = game.ReflexiveBody("earth-rumble/fight",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		return fightTargetClauses(NewContext(g, item))
	}),
	constTargets(func() *game.TargetSpec {
		return Clauses(
			TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1),
			TargetCreature("target creature an opponent controls", OpponentControls()))
	}))
