package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Myr Retriever — Artifact Creature — Myr {2}, 1/1 (EDHREC rank 878):
//
//	"When this creature dies, return another target artifact card
//	 from your graveyard to your hand."
//
// The artifact deck's recursion piece, and half of a famous loop
// (two Retrievers and a sacrifice outlet). A dies trigger with a
// graveyard target, Sun Titan's shape.
//
// "ANOTHER" is exact. The clause is built per trigger through
// AnotherTarget (TriggeredAbility.TargetsFrom, which is handed the
// source), so the picker excludes THIS Retriever by instance rather
// than by name: a second Myr Retriever — a Clone or a token copy of
// this one — is a legal target, as printed, and the Retriever whose
// trigger it is never is. The resolution re-check (CR 608.2b) runs
// the same clause.
func init() {
	Register(Spec{
		OracleID:     "d07d3be3-f69d-4484-8467-cffd43871788",
		Name:         "Myr Retriever",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			TargetsFrom: AnotherTarget(func(other CardPredicate) *game.TargetSpec {
				return TargetCardInGraveyard("another target artifact card in your graveyard", Artifact(), YouOwn(), other)
			}),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Myr Retriever — return an artifact card to hand",
					returnTargetedCardToHand)
			},
		}},
	})
}
