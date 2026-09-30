package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ares, God of War — Legendary Creature — God Warrior Villain {1}{B}{R},
// 4/3 (EDHREC rank 11204):
//
//	"Ares attacks each combat if able.
//	 Whenever an attacking creature you control dies, return that card
//	 to its owner's hand."
//
// The first line is AttacksEachCombat, Zurgo Helmsmasher's CR 508.1d
// requirement.
//
// The second is diedWhileAttacking (#1661): the dying creature's attack
// as it last existed, carried on its leaves-the-battlefield event
// because the exit has already cleared it from the card (CR 603.10a).
// "You control" is the controller it had as it left. Ares dying while
// attacking sees his own death, so he goes back to hand too.
//
// "Return THAT CARD" names one object (CR 400.7), so the trigger
// records the card's graveyard epoch when it fires and returns it only
// if it is still that object in a graveyard when the trigger resolves —
// Edea's shape. A card exiled from the graveyard in response stays
// exiled, and a token has ceased to exist (CR 111.7), so it returns
// nothing. The owner's hand, not the controller's: a stolen attacker
// goes home.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a09d4a2-4796-4a15-ab75-a3a0b717f62d",
		Name:         "Ares, God of War",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedWhileAttacking(ev, g)
				return ok && leftUnderControlOf(ev, dead) == source.Controller
			},
			Key: aresReturnLabel,
			// The dead card's post-move epoch is a board read made at
			// trigger time (ADR 0041 P9's fill-in Build), exactly as
			// Edea's: it is what the resolution compares against to
			// know "that card" is still the same object.
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, aresReturnLabel)
				epoch := -1
				if c, ok := g.LookupCardForEffect(ev.CardID); ok {
					epoch = c.ObjectEpoch
				}
				item.Params.Object = game.ObjectRef{ID: ev.CardID, Epoch: epoch}
				return item
			},
			Effect: aresReturnThatCard,
		}},
	})
}

const aresReturnLabel = "Ares, God of War — return that card to its owner's hand"

// aresReturnThatCard returns the dead card to its owner's hand if it is
// still the same object in a graveyard.
func aresReturnThatCard(g *game.Game, item *game.StackItem) error {
	dead := item.Params.Object
	c, ok := g.LookupCardForEffect(dead.ID)
	if !ok || c.ObjectEpoch != dead.Epoch {
		return nil
	}
	if z := g.FindCardZoneForEffect(dead.ID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ReturnFromGraveyard{Target: dead.ID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
}
