package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shared by the Phyrexia: All Will Be One oil-counter creatures of S82 batch 9.

// s82OneB09Oil is the name of the counter every oil card counts.
const S82OneB09Oil = "oil"

// s82OneB09OilOnNoncreatureCast is "Whenever you cast a noncreature spell,
// put an oil counter on this creature." The spell is read off the stack by
// b10NoncreatureSpellCastByYou; the counter lands only while the creature is
// still the permanent that triggered (CounterOnThis).
func S82OneB09OilOnNoncreatureCast(name string) game.TriggeredAbility {
	return On(game.EventCast, B10NoncreatureSpellCastByYou,
		name+" — put an oil counter on it",
		Do(CounterOnThis{Kind: S82OneB09Oil, N: 1}))
}

// s82OneB09OilCount is the oil counters on a permanent.
func S82OneB09OilCount(c *game.Card) int {
	if c == nil {
		return 0
	}
	return c.Counters[S82OneB09Oil]
}
