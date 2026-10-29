package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Head of the Hunt — Creature — Wolf {2}{B}{B}, 4/3 (The Hobbit, #4454):
//
//	"Flash
//	 If a creature an opponent controls would die, exile it instead. When
//	 you do, create a 2/2 green Wolf creature token."
//
// The replacement is Liesa's "if a creature an opponent controls would
// die, exile it instead" (rfOpponentCreatureWouldDie). Tokens are
// creatures too, so they are exiled as well. The creature never died
// (CR 700.4), so its dies-triggers do not fire.
//
// "When you do" is a CR 603.12 reflexive trigger whose creator is the
// replacement (#4517): it is declared on the move as the replacement
// applies (QueueReflexiveTriggerAfterReplacedMoveForEffect) and created
// once the creature has landed in exile, under Head of the Hunt's
// controller, going on the stack the next time a player would receive
// priority. Each creature exiled this way makes its own trigger, so a
// board wipe makes one Wolf per opponent creature. When the controller
// of a dying creature applies Rest in Peace first instead, this
// replacement never applies and there is no Wolf; the trigger watches
// this replacement, not the exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0523e271-5e0b-4e4b-9eac-86f915d18952",
		Name:            "Head of the Hunt",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: rfOpponentCreatureWouldDie,
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				return g.QueueReflexiveTriggerAfterReplacedMoveForEffect(ev, src, game.ReflexiveTrigger{
					Label: "Head of the Hunt — create a 2/2 green Wolf creature token",
					Body:  headOfTheHuntWolfBody,
				})
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Head of the Hunt: exile instead of dying",
		}},
	})
}

// headOfTheHuntWolfBody is the reflexive trigger: one 2/2 green Wolf for
// the trigger's controller.
var headOfTheHuntWolfBody = game.ReflexiveBody("head-of-the-hunt/wolf", simpleBody(func(g *game.Game, item *game.StackItem) error {
	return CreateToken{Template: TokenCard("2/2 green Wolf"), N: 1}.Apply(NewContext(g, item))
}), nil)
