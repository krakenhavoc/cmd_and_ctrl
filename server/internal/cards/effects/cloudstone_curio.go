package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cloudstone Curio — Artifact {3} (#387, #1223):
//
//	"Whenever a nonartifact permanent you control enters, you may
//	 return another permanent you control that shares a permanent
//	 type with it to its owner's hand."
//
// The second card on the trigger-data row, and the other half of what
// that row means. Scrap Trawler's clause reads a NUMBER off the
// event; this one reads a TYPE LINE — "shares a permanent type WITH
// IT", where "it" is the permanent that just entered and nothing in
// a static clause can name it.
//
// `TargetsFrom` builds the clause from the trigger's own CR 603.10
// snapshot: the entering permanent's permanent types (CR 110.4a),
// intersected against each candidate. An artifact creature entering
// offers both the artifacts and the creatures, which is the card's
// whole engine — a Bounce-and-replay loop only closes when the two
// ends share a type.
//
// Sandbox simplification, declared (the Whitemane Lion posture): "you
// may return another permanent you control" is a resolution-time
// choice, not a target, and the pick_target prompt is the one picker
// the engine has for choosing among permanents — so the permanent is
// chosen as the trigger goes on the stack, with a count of "up to
// one" carrying the "you may". Weaker than printed on two counts:
// opponents see the choice before the trigger resolves, and a
// permanent of yours with shroud cannot be the one returned.
func init() {
	Register(Spec{
		OracleID:     "5cd2fd32-4da2-40eb-b003-c0b9a9ec91c1",
		Name:         "Cloudstone Curio",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"You pick the permanent to return when the trigger goes on the stack rather than as it resolves, so opponents can respond to the choice, and a permanent of yours with shroud can't be picked.",
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: cloudstoneNonartifactEntered,
			// TargetsFrom, not Targets: "shares a permanent type with
			// it" is a fact about what entered. See trigger_event.go.
			TargetsFrom: cloudstoneSharesATypeClause,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source,
					"Cloudstone Curio — return another permanent you control to its owner's hand",
					bounceChosenTarget)
			},
		}},
	})
}

// cloudstoneNonartifactEntered is the trigger condition: a
// NONARTIFACT permanent entered the battlefield under the Curio's
// controller's control.
//
// The Curio is itself an artifact, so it never triggers on its own
// entry, and the exclusion is the printed one rather than an
// "another" clause — a second nonartifact permanent entering does
// trigger it.
func cloudstoneNonartifactEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsArtifact() && c.Controller == source.Controller
}

// cloudstoneSharesATypeClause builds the clause from the triggering
// event: "another permanent you control that shares a permanent type
// with it", as an up-to-one so declining is an answer.
//
// "Another" excludes the permanent that entered, by instance id —
// which is also what keeps a creature from bouncing itself the moment
// it lands and makes the Curio a two-card loop rather than a one-card
// one.
//
// An event with no object snapshot yields a clause nothing satisfies
// rather than nil, for the reason Scrap Trawler's does: nil would say
// "this trigger targets nothing" and put an ability on the stack that
// returns nothing.
func cloudstoneSharesATypeClause(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	entered := tc.Object
	return TargetPermanent(
		"another permanent you control that shares a permanent type with the permanent that entered",
		YouControl(),
		func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return entered != nil && c.InstanceID != entered.ID && entered.SharesPermanentTypeWith(c)
		},
	).WithCount(0, 1)
}
