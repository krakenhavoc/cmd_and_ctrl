package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archangel Avacyn // Avacyn, the Purifier — {3}{W}{W} Legendary Creature —
// Angel 4/4 // Legendary Creature — Angel 6/5 (Innistrad Remastered, S82
// GAP, #5948):
//
//	Front: "Flash
//	        Flying, vigilance
//	        When Archangel Avacyn enters, creatures you control gain
//	        indestructible until end of turn.
//	        When a non-Angel creature you control dies, transform
//	        Archangel Avacyn at the beginning of the next upkeep."
//	Back:  "Flying
//	        When this creature transforms into Avacyn, the Purifier, it
//	        deals 3 damage to each other creature and each opponent."
//
// The enters trigger grants indestructible to the creatures you control as
// it resolves (CR 611.2c). The dies trigger looks back at the creature as
// it last existed (CR 603.10a): any creature YOU controlled that was not an
// Angel, tokens included, Avacyn herself never (she is an Angel). It sets
// up a delayed trigger (CR 603.7) for the next upkeep of any player; the
// transform is the permanent turning over in place, and by CR 701.27f a
// delayed trigger transforms it only if it hasn't transformed since the
// trigger was created, so several deaths before one upkeep transform her
// once. The back face's trigger fires when she transforms into the
// Purifier and deals the 3 damage as one instance to every other creature
// (hers included) and each opponent.
//
// No simplification.
func init() {
	const oracle = "432b37a5-d32a-4b78-91ab-860aa026b7cc"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Archangel Avacyn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Archangel Avacyn — creatures you control gain indestructible until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{
						Match:    And(Creature(), YouControl()),
						Keywords: []string{"indestructible"},
						Label:    "Archangel Avacyn — indestructible until end of turn",
					}.Apply(NewContext(g, item))
				}),
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && leftUnderControlOf(ev, dead) == source.Controller && !leftAsSubtype(ev, dead, source.TextSubtype("Angel"))
			}, "Archangel Avacyn — transform Archangel Avacyn at the beginning of the next upkeep", archangelAvacynSchedule),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Avacyn, the Purifier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventTransform, ThisTransformedInto(1),
				"Avacyn, the Purifier — 3 damage to each other creature and each opponent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					self := ctx.Source()
					other := And(Creature(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != self })
					return damageEachMatchingAndEachOpponent(ctx, other, 3)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3, Partial: true}}),
		},
	})
}

// archangelAvacynSchedule is the dies trigger's effect: a delayed trigger
// at the beginning of the next upkeep, whoever's it is, bound to this
// Avacyn object.
func archangelAvacynSchedule(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	if !onBattlefield(g, id) {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:    game.StepUpkeep,
		Label: "Archangel Avacyn — transform Archangel Avacyn",
		Body:  archangelAvacynTransformBody,
	}.Apply(NewContext(g, item))
}

// archangelAvacynTransformBody is the delayed trigger: transform Avacyn if
// she is still the object the ability named and has not transformed since
// the trigger was created (CR 701.27f). An on-disk key: never renamed,
// never reused.
var archangelAvacynTransformBody = game.DelayedBody("archangel-avacyn/transform",
	func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
		return TransformThis{}.Apply(NewContext(g, item))
	})
