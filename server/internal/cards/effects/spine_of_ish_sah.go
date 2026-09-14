package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spine of Ish Sah — Artifact {7} (EDHREC rank 2568):
//
//	"When this artifact enters, destroy target permanent.
//	 When this artifact is put into a graveyard from the battlefield,
//	 return it to its owner's hand."
//
// Seven mana for a Vindicate that comes back. The entry trigger is a
// targeted ETB over any permanent (destroyChosenTargetTrigger, the
// Acidic Slime shape — the controller picks as it goes on the
// stack, an indestructible target survives). The second is a dies
// trigger: cardDied is the graveyard-only exit, and on resolution
// the Spine goes from the graveyard to its owner's hand if it is
// still there — a Spine exiled from the graveyard in response stays
// exiled (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "02f062f5-8012-4440-ac12-49fc49822106",
		Name:         "Spine of Ish Sah",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Targets:   TargetPermanent("target permanent"),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return destroyChosenTargetTrigger(source, "Spine of Ish Sah — destroy target permanent")
				},
			},
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return cardDied(ev, source)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Spine of Ish Sah — return it to its owner's hand",
						func(g *game.Game, item *game.StackItem) error {
							if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
								return nil
							}
							return ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}
