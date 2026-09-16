package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sword of the Animist — Legendary Artifact — Equipment for {2}
// (EDHREC rank 236, the highest-ranked Equipment in this batch):
//
//	"Equipped creature gets +1/+1.
//	 Whenever equipped creature attacks, you may search your library
//	 for a basic land card, put it onto the battlefield tapped, then
//	 shuffle.
//	 Equip {2}"
//
// Ramp stapled to an attack, which is why it is played in decks that
// do not otherwise want Equipment at all: put it on any creature that
// can attack safely and it is a Rampant Growth every turn.
//
// THE FIRST "WHENEVER EQUIPPED CREATURE ATTACKS" IN THE CATALOG. The
// condition is attachedCreatureAttacked, and EventAttack's shape does
// the work: it carries the attacking creature in CardID and fires
// once per creature on its FIRST declaration, so one Sword on one
// creature is exactly one trigger per combat. No batching guard is
// needed — the Curse of Opulence problem is the mirror image (that
// card watches the defender and sees one event per attacker).
//
// "YOU MAY SEARCH" IS THE PROMPT, NOT AN ASSUMPTION.
// SearchLibrary.Optional forces the search prompt even when the
// choice looks free, so the searcher can decline the land AND the
// shuffle — which is a real decision for a deck that has stacked its
// top. TappedOnEntry is the card's own "tapped" clause.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d79cbc61-6c15-48ea-bbba-3cffb819ccba",
		Name:         "Sword of the Animist",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attachedCreatureAttacked(ev, source)
			}, "Sword of the Animist — search for a basic land", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					TappedOnEntry: true,
					Optional:      true,
					Shuffle:       true,
					Reason:        "Sword of the Animist — search for a basic land card",
				}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
