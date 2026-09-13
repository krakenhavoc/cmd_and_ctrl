package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Juri, Master of the Revue — Legendary Creature — Human Shaman
// {B}{R}, 1/1 (EDHREC rank 1824):
//
//	"Whenever you sacrifice a permanent, put a +1/+1 counter on Juri.
//	 When Juri dies, it deals damage equal to its power to any
//	 target."
//
// The Rakdos sacrifice commander. Two triggers:
//
//   - The counter is "whenever you sacrifice a permanent" —
//     EventSacrifice with the controller as Actor, one per permanent
//     (Mayhem Devil's shape). Sacrificing Juri himself fires it, but
//     the counter finds nothing to land on; the dies trigger is the
//     one that matters then.
//   - The dies trigger is Omnath's any-target shape, with the damage
//     read as last-known power: the harvester's LKI characteristic
//     (layers applied) plus the counters Juri carried when he left,
//     which b13LastKnownPower reads back off the event log because
//     the LKI carries no counter math.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3364bad6-6d0a-4141-b410-86e3d9e1916e",
		Name:         "Juri, Master of the Revue",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventSacrifice},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor == source.Controller && ev.CardID != uuid.Nil
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Juri, Master of the Revue — put a +1/+1 counter on Juri",
						func(g *game.Game, item *game.StackItem) error {
							if !b15OnBattlefield(g, item.SourceCardID) {
								return nil
							}
							return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return cardDied(ev, source)
				},
				Targets: TargetAny(),
				Build: func(_ game.Event, source *game.Card, lki game.Characteristic, g *game.Game) *game.StackItem {
					power := b13LastKnownPower(g, source.InstanceID, lki)
					return game.NewTriggeredItem(source, "Juri, Master of the Revue — damage equal to its power to any target",
						func(g *game.Game, item *game.StackItem) error {
							ctx := NewContext(g, item)
							for _, t := range ctx.LegalTargets() {
								return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: power}.Apply(ctx)
							}
							return nil
						})
				},
			},
		},
	})
}
