package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Herald of Ilharg — Creature — Boar Beast {2}{R}{G}, 3/3 (Ravnica: Clue
// Edition, #5929):
//
//	"Trample
//	 Whenever you cast a creature spell, put two +1/+1 counters on this
//	 creature. If that spell has mana value 5 or greater, this creature
//	 deals damage equal to the number of counters on it to each
//	 opponent."
//
// One trigger: Animar's creature-spell watch, two counters, then the
// spell's mana value (Saruman's read) gates the damage, which counts the
// +1/+1 counters the Herald has after the new ones landed. The damage
// is the Herald's (a source that left deals it as it last existed, CR
// 608.2h, and the counters then read as zero only if it never had any).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "80d1866f-b5a5-42c6-b81a-851a34ec0cd9",
		Name:            "Herald of Ilharg",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return creatureSpellCastByYou(ev, source, g)
			}, "Herald of Ilharg — put two +1/+1 counters on it; if that spell has mana value 5 or greater, it deals damage equal to the number of counters on it to each opponent",
				s82BnsB12HeraldCast),
		},
	})
}
