package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Earth Kingdom General — Creature — Human Soldier Ally {3}{G}, 2/2
// (Avatar: The Last Airbender, #4459):
//
//	"When this creature enters, earthbend 2.
//	 Whenever you put one or more +1/+1 counters on a creature, you may
//	 gain that much life. Do this only once each turn."
//
// The entry is the earthbend keyword action with its own target clause.
//
// The second ability is Terrasymbiosis's trigger with a life gain:
// WheneverYouPutCountersOfKindOnACreature, once per creature per
// placement you made (CR 603.2c), any creature's controller. "That much"
// is the +1/+1 counters that placement put, after replacements, read as
// it resolves. "Do this only once each turn" limits the gaining: every
// placement triggers and asks its "you may" as it resolves (CR 603.5);
// the first "yes" gains and uses up the turn, a "no" does not (the
// ruling), and once the turn's gain is done no further trigger is made.
// The earthbend's own counters are a placement you made, so the entry
// can feed it.
//
// No simplification.
func init() {
	gain := WheneverYouPutCountersOfKindOnACreature(game.CounterPlusOne, nil, earthKingdomGeneralLabel,
		func(g *game.Game, item *game.StackItem) error {
			if b28DoneThisTurn(g, item.SourceCardID, earthKingdomGeneralDone) {
				return nil
			}
			n := plusOneCountersPutInTheTriggeringBatch(g, item)
			if n <= 0 {
				return nil
			}
			b28NoteDoneThisTurn(g, item.SourceCardID, earthKingdomGeneralDone)
			return GainLife{Player: item.Controller, Amount: n}.Apply(NewContext(g, item))
		})
	when := gain.AppliesTo
	gain.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		return when(ev, source, ch, g) && !b28DoneThisTurn(g, source.InstanceID, earthKingdomGeneralDone)
	}
	Register(Spec{
		OracleID:     "e43c31ee-1bac-4f96-8b74-334b3238ab73",
		Name:         "Earth Kingdom General",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			avB14WhenEntersEarthbend("Earth Kingdom General — earthbend 2", 2),
			Optional(gain, "Earth Kingdom General — gain that much life?"),
		},
	})
}

const (
	// earthKingdomGeneralLabel is the life trigger's stack label.
	earthKingdomGeneralLabel = "Earth Kingdom General — gain that much life"
	// earthKingdomGeneralDone is the once-each-turn record, a label of
	// its own (b28DoneThisTurn says why).
	earthKingdomGeneralDone = "Earth Kingdom General — life gained this turn"
)
