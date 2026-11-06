package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b14_helpers.go — shared helpers for Avatar: The Last Airbender,
// batch 14 (#4459).

// avB14Earthbend is "earthbend N" as a triggered ability's effect: the
// first still-legal target land (CR 608.2b) is earthbent with the
// count `count` reads as the ability resolves. The trigger's target
// clause is EarthbendTargets(); pair them with avB14EarthbendTrigger.
func avB14Earthbend(count func(*Context) int) Effect {
	body := EarthbendFirstTarget(count)
	return func(g *game.Game, item *game.StackItem) error {
		return body(item, NewContext(g, item))
	}
}

// avB14WhenEntersEarthbend is "When this creature enters, earthbend N"
// (Badgermole Cub, Earth Kingdom General): the keyword's own "target
// land you control", chosen as the trigger is put on the stack
// (CR 603.3d).
func avB14WhenEntersEarthbend(label string, n int) game.TriggeredAbility {
	return Targeting(WhenThisEnters(label, avB14Earthbend(EarthbendCount(n))), EarthbendTargets())
}
