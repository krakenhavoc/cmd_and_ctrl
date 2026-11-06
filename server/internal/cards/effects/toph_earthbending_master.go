package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Toph, Earthbending Master — Legendary Creature — Human Warrior Ally
// {3}{G}, 2/4 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Landfall — Whenever a land you control enters, you get an
//	 experience counter.
//	 Whenever you attack, earthbend X, where X is the number of
//	 experience counters you have."
//
// The experience counter goes on the player (CR 122.1), through Aang,
// Airbending Master's youGetAnExperienceCounter, and stays if Toph dies.
// "Whenever you attack" triggers once for a declaration in which you
// attack with one or more creatures (CR 508.3d), whatever they attack:
// OncePerBatch over that declaration's attack events. X is read as the
// earthbend resolves, so it counts every experience counter you have by
// then, however you got it (the ruling). Earthbend is the shared verb
// (CR 701.66a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad5d3e7e-bf8c-4fa3-a81e-1523aec2585f",
		Name:         "Toph, Earthbending Master",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Toph, Earthbending Master — you get an experience counter", youGetAnExperienceCounter),
			OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Toph, Earthbending Master — earthbend X, where X is the number of experience counters you have",
					avB14Earthbend(func(ctx *Context) int { return experienceCounters(ctx.Game, ctx.Controller()) })),
				EarthbendTargets())),
		},
	})
}
