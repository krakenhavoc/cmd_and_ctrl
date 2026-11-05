package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood-Sport Champion — Creature — Klingon Warrior {1}{B}, 2/2 (Star
// Trek Commander, #4974):
//
//	"Kicker {2} (You may pay an additional {2} as you cast this spell.)
//	 Deathtouch
//	 When this creature enters, if it was kicked, other creatures you
//	 control gain deathtouch until end of turn."
//
// The kicked check is an intervening "if" (CR 603.4), read off the
// permanent's carried kicker record when the ability would trigger
// (Gatekeeper of Malakir's shape, CR 702.33c), so an unkicked Champion
// puts nothing on the stack. The creatures are locked as the ability
// resolves (CR 611.2c): one that enters later this turn does not gain
// deathtouch. "Other" is the source object (CR 400.7), so a Champion
// that left and came back is a new object and is affected like any other
// creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9d08cc9b-2b26-49de-80aa-962a358bfc1a",
		Name:            "Blood-Sport Champion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		OptionalCosts:   []game.AdditionalCost{Kicker("{2}")},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: AllOf(Self, func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return game.CardKickedTimes(*source) > 0
			}),
			Key: "Blood-Sport Champion — kicked, other creatures you control gain deathtouch",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl(), trekB11OtherThanSourceObject(item)),
					Keywords: []string{"deathtouch"},
					Label:    "Blood-Sport Champion — deathtouch",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
