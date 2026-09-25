package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Terastodon — Creature — Elephant {6}{G}{G}, 9/9 (EDHREC rank
// 1397):
//
//	"When this creature enters, you may destroy up to three target
//	 noncreature permanents. For each permanent put into a graveyard
//	 this way, its controller creates a 3/3 green Elephant creature
//	 token."
//
// Green's Vindicate-on-a-body: three of anything that is not a
// creature, and the usual line is to eat your own lands for three
// 3/3s. "You may … up to three" is a zero-to-three target clause
// chosen when the trigger goes on the stack — picking nothing is
// declining — and each target still legal at resolution (CR 608.2b)
// is destroyed through the single-target destroy verb, the one that
// honours indestructible (the Sylvan Reclamation shape, one target
// at a time). Each permanent's controller is read BEFORE it is
// destroyed and its zone AFTER, so only what actually reached a
// graveyard earns an Elephant: an indestructible permanent, or a
// commander that went to the command zone instead, earns nothing, as
// printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17975a20-2c13-4325-be1e-0bcda5063a2d",
		Name:         "Terastodon",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetPermanent("up to three target noncreature permanents", Noncreature()).WithCount(0, 3),
			Key:     "Terastodon — destroy up to three noncreature permanents, Elephants for their controllers",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				var victims []b12Victim
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					controller, ok := controllerOfTarget(ctx, t.ID)
					if !ok {
						continue
					}
					victims = append(victims, b12Victim{ID: t.ID, Controller: controller})
				}
				for _, v := range victims {
					if err := (DestroyTarget{Target: v.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return b12ElephantsForTheDestroyed(ctx, victims)
			},
		}},
	})
}
