package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Specimen Freighter — Artifact — Spacecraft {5}{U}, 4/7 (Edge of
// Eternities, #4461):
//
//	"When this Spacecraft enters, return up to two target non-Spacecraft
//	 creatures to their owners' hands.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 9+.)
//	 9+ | Flying
//	 Whenever this Spacecraft attacks, defending player mills four cards."
//
// The enters trigger is Horses of the Bruinen's return, the creatures
// going together; a Spacecraft that is a creature is not a legal target,
// and the targets still legal as it resolves are returned (CR 608.2b).
// Station and the 9+ line are Susurian Dirgecraft's helpers (CR 721.2).
// The attack trigger is Nemesis of Reason's: the defending player is a
// fact of the attack declaration, not a target, and one who has left
// the game by resolution mills nothing.
//
// No simplification.
func init() {
	const millLabel = "Specimen Freighter — defending player mills four cards"
	Register(Spec{
		OracleID:     "fb84b94c-dc8c-4182-9b96-abd092b20981",
		Name:         "Specimen Freighter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Specimen Freighter — return up to two target non-Spacecraft creatures to their owners' hands",
				func(g *game.Game, item *game.StackItem) error {
					g.BounceCardsToHandForEffect(legalTargetCardIDs(NewContext(g, item)))
					return nil
				}),
				TargetCreature("up to two target non-Spacecraft creatures", Not(Subtype("Spacecraft"))).WithCount(0, 2)),
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source)
				},
				Key: millLabel,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, millLabel)
					item.Params.Player = b17DefendingPlayer(g, ev)
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					defender := item.Params.Player
					if defender == uuid.Nil || g.PlayerByIDForEffect(defender) == nil {
						return nil
					}
					return MillCards{Player: defender, N: 4}.Apply(NewContext(g, item))
				},
			},
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(9, 4, 7),
			ThresholdKeywords(9, "flying"),
		),
	})
}
