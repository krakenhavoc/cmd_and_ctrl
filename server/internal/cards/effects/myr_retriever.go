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
// "ANOTHER" is enforced at resolution, not in the picker: the
// Retriever is itself an artifact card in your graveyard by the time
// the trigger targets, and TargetSpec.CardOK never sees the trigger's
// source, so the zone browser will offer it. Picking it does nothing
// — a wasted trigger, which is weaker than printed; returning it
// would be the loop the printed word exists to forbid, and that is
// the direction #259 rules out. Declared here rather than hidden.
func init() {
	Register(Spec{
		OracleID:     "d07d3be3-f69d-4484-8467-cffd43871788",
		Name:         "Myr Retriever",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The trigger can be pointed at Myr Retriever itself, which returns nothing — pick another artifact card."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Targets: TargetCardInGraveyard("another target artifact card in your graveyard", Artifact(), YouOwn()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Myr Retriever — return an artifact card to hand",
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
