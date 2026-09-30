package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Death Tyrant — Creature — Beholder Skeleton {4}{B}, 4/6 (EDHREC rank
// 8784):
//
//	"Menace
//	 Negative Energy Cone — Whenever an attacking creature you control
//	 or a blocking creature an opponent controls dies, create a 2/2
//	 black Zombie creature token.
//	 {5}{B}: Return this card from your graveyard to the battlefield
//	 tapped."
//
// The trigger is two conditions on one ability, each read off the
// dying creature's leaves-the-battlefield event (#1661, CR 603.10a):
// diedWhileAttacking for "an attacking creature you control" and
// diedWhileBlocking for "a blocking creature an opponent controls".
// The exit clears both from the card before the event fires, so the
// event is the only place either fact survives. The controller is the
// one the creature had as it left, which a zone change does not reset
// (the same read every "creature you control dies" trigger makes).
//
// Crossing the clauses matters and is pinned by a test: YOUR blocker
// dying is not "a blocking creature an opponent controls", and an
// OPPONENT's attacker dying is not "an attacking creature you
// control". A Death Tyrant that dies attacking sees its own death —
// the LTB harvest reads its own event the same way.
//
// The graveyard return is Reassembling Skeleton's ability at a higher
// price: the same ZoneGraveyard activation and the same
// returnThisFromGraveyardTapped body, entering tapped through the
// CR 614 entry.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fa944633-e078-41bb-9c7d-4f6b669c27a5",
		Name:            "Death Tyrant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, deathTyrantCombatDeath,
				"Death Tyrant — Negative Energy Cone: create a 2/2 black Zombie",
				func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: BlackZombieToken(), N: 1}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label: "{5}{B}: Return this card from your graveyard to the battlefield tapped.",
			Cost:  ManaCost("{5}{B}"),
			Zones: []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return returnThisFromGraveyardTapped(g, item)
			},
		}},
	})
}

// deathTyrantCombatDeath is Negative Energy Cone's condition: an
// attacking creature the source's controller controlled died, or a
// blocking creature someone else controlled did.
func deathTyrantCombatDeath(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if dead, ok := diedWhileAttacking(ev, g); ok && leftUnderControlOf(ev, dead) == source.Controller {
		return true
	}
	dead, ok := diedWhileBlocking(ev, g)
	return ok && leftUnderControlOf(ev, dead) != source.Controller
}
