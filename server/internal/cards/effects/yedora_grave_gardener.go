package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yedora, Grave Gardener — Legendary Creature — Treefolk Druid {4}{G},
// 5/5:
//
//	"Whenever another nontoken creature you control dies, you may
//	 return it to the battlefield face down under its owner's control.
//	 It's a Forest land. (It has no other types or abilities.)"
//
// Roadmap batch 30 (#393), skipped there on "#656 face-down objects"
// and then on CR 708.2's LISTED characteristics (#1270), which is the
// card that shows a listing REPLACES the CR 708.2a body rather than
// decorating it: a face-down Forest is not a creature at all. The
// body is ForestLandBody(); its "{T}: Add {G}" is CR 305.6's, which the
// engine derives from the Forest subtype.
//
// The return is game.ReturnFromGraveyardFaceDownForEffect — the
// reanimation pipeline with the face-down state riding the entry
// event, so no ETB trigger and no "as enters" hook runs (CR 708.2a),
// and the controller is the only player who may look (CR 708.5). The
// kind is `turned`: the effect, not a keyword, put it here, so a card
// with morph or disguise can still be turned up for its own cost
// (CR 702.37e) and anything else stays a Forest for good (CR 708.7).
//
// "Under its OWNER's control": uuid.Nil controller. The trigger is
// "you control", read off the creature as it died, so Yedora sees its
// controller's creatures, including a stolen one — which then comes
// back to its owner.
//
// "You may" is asked when the trigger fires rather than at resolution
// — the engine's standard optional-trigger shape. If the card has left
// the graveyard by the time the trigger resolves, it is a new object
// (CR 400.7) and nothing happens.
func init() {
	Register(Spec{
		OracleID:     "1fa20b05-cc06-4fb7-ae76-677a29e2185a",
		Name:         "Yedora, Grave Gardener",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventLTB},
				Key:     "Yedora, Grave Gardener — return it face down as a Forest land",
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b19AnotherNontokenCreatureYouControlDied(ev, source, g)
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return yedoraReturnFaceDown(g, item.Trigger.Event.CardID)
				},
			}, "Yedora, Grave Gardener — return it to the battlefield face down as a Forest land?"),
		},
	})
}

// yedoraReturnFaceDown is the trigger's resolution: the card is
// returned from its owner's graveyard, face down, as a Forest land. A
// card no longer in a graveyard is the CR 400.7 "nothing happens".
func yedoraReturnFaceDown(g *game.Game, died uuid.UUID) error {
	_, err := g.ReturnFromGraveyardFaceDownForEffect(died, uuid.Nil, false, ForestLandBody())
	if errors.Is(err, game.ErrCardNotFound) {
		return nil
	}
	return err
}
