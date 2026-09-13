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
// Declared, the Myr Retriever posture: "ANOTHER" is enforced at
// resolution, not in the picker. The Diver is itself an artifact
// card in your graveyard by the time the trigger targets, and
// TargetSpec.CardOK never sees the trigger's source, so the zone
// browser will offer it. Picking it does nothing — a wasted trigger,
// which is weaker than printed; returning it would be the loop the
// printed word exists to forbid.
func init() {
	Register(Spec{
		OracleID:        "08c595da-9305-42a9-b72f-5ccc546edc01",
		Name:            "Junk Diver",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The trigger can be pointed at Junk Diver itself, which returns nothing — pick another artifact card."},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Targets: TargetCardInGraveyard("another target artifact card in your graveyard", Artifact(), YouOwn()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Junk Diver — return an artifact card to hand",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						if item.Targets[0].ID == item.SourceCardID {
							return nil // "another" — see the card comment
						}
						return ReturnFromGraveyard{
							Target: item.Targets[0].ID,
							Dest:   game.ZoneHand,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
