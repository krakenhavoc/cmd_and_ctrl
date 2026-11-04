package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invisible Woman, Sue Storm — {4}{W} Legendary Creature — Human Hero 2/5
// (Marvel Super Heroes, #4455):
//
//	"Lifelink
//	 Whenever you put one or more +1/+1 counters on one or more other
//	 Heroes you control, you may create a 0/4 colorless Wall creature
//	 token with defender."
//
// Lifelink is a printed keyword. The trigger is one per event batch
// (CR 603.2c): the clause names "one or more other Heroes", not each
// one, so it is OncePerBatch with no per-object key, and counters put on
// three Heroes at once make one Wall. The Hero is read through the
// layered view, so a changeling counts; the source itself is excluded,
// and only counters you put count. The "may" is a real prompt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e9d20b6d-fa4f-473b-9866-230fd92a1319",
		Name:            "Invisible Woman, Sue Storm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			Optional(OncePerBatch(On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if source == nil || ev.Label != game.CounterPlusOne || ev.Target == source.InstanceID {
					return false
				}
				if _, ok := b12CountersPlacedByYou(ev, source, g); !ok {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.Target)
				return ok && c.IsCreature() && c.Controller == source.Controller && c.HasSubtypeWord("Hero")
			}, "Invisible Woman, Sue Storm — create a 0/4 Wall", Do(CreateToken{
				Template: TokenCard("0/4 colorless Wall with defender"),
				N:        1,
			}))), "Invisible Woman, Sue Storm — create a 0/4 colorless Wall with defender?"),
		},
	})
}
