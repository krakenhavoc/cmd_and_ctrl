package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Junk Diver — Artifact Creature — Bird {3}, 1/1 (EDHREC rank 1962):
//
//	"Flying
//	 When this creature dies, return another target artifact card
//	 from your graveyard to your hand."
//
// Myr Retriever with wings: the artifact deck's recursion piece and
// the other half of the Retriever loop. A dies trigger with a
// graveyard target, Sun Titan's shape.
//
// "ANOTHER" is exact. The clause is built per trigger through
// AnotherTarget (TriggeredAbility.TargetsFrom, which is handed the
// source), so the picker excludes THIS Diver by instance rather than
// by name: a second Junk Diver — a Clone or a token copy of this one
// — is a legal target, as printed, and the Diver whose trigger it is
// never is. The resolution re-check (CR 608.2b) runs the same clause.
func init() {
	Register(Spec{
		OracleID:        "08c595da-9305-42a9-b72f-5ccc546edc01",
		Name:            "Junk Diver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			TargetsFrom: AnotherTarget(func(other CardPredicate) *game.TargetSpec {
				return TargetCardInGraveyard("another target artifact card in your graveyard", Artifact(), YouOwn(), other)
			}),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Junk Diver — return an artifact card to hand",
					returnTargetedCardToHand)
			},
		}},
	})
}
