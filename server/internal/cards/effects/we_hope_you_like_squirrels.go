package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// We Hope You Like Squirrels — Plane — Secret Lair (pssc, #4870):
//
//	"At the beginning of each player's upkeep and whenever a player casts
//	 a spell, that player creates a 1/1 green Squirrel creature token.
//	 Whenever chaos ensues, put a +1/+1 counter on each Squirrel you
//	 control."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The first is ONE ability with two trigger conditions (OnAny, as Antarctic
// Research Base's is); "that player" is the event's actor, the upkeep's
// player or the spell's caster, who creates the token, not the planar
// controller. The chaos ability (CR 311.7) counts "you" as the planar
// controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7aac9628-f112-4840-91b4-7bb5cd896685",
		Name:         "We Hope You Like Squirrels",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventBeginUpkeep, game.EventCast}, AnyPlayer,
				"We Hope You Like Squirrels — that player creates a 1/1 green Squirrel creature token",
				func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: triggeringActor(item), Template: TokenCard("1/1 green Squirrel"), N: 1}.Apply(NewContext(g, item))
				}),
			WheneverChaosEnsues("We Hope You Like Squirrels — a +1/+1 counter on each Squirrel you control",
				avB30aPutCounterOnEachOfSubtypeYouControl("Squirrel")),
		},
	})
}
