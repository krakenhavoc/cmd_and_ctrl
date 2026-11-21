package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Patron of the Vein — Creature — Vampire Shaman {4}{B}{B}, 4/4
// (EDHREC rank 3514):
//
//	"Flying
//	 When this creature enters, destroy target creature an opponent
//	 controls.
//	 Whenever a creature an opponent controls dies, exile it and put a
//	 +1/+1 counter on each Vampire you control."
//
// The Vampire deck's top end: removal on entry, and every opposing
// death — its own entry kill first — exiles the card and grows the
// team. The entry trigger is mandatory and targeted, so it is dropped
// with no legal target (CR 603.3d). The dies trigger is diedCreature
// under another player's control; its body exiles the dead card only
// if it is still in a graveyard (a card already reanimated, or an
// opponent's commander that took the command zone, is left alone)
// and then puts the counter on every Vampire the controller controls
// — the Patron himself included, as printed — whether or not the
// exile happened.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8dbc8fdb-36ce-4f30-b679-9c2029fcd9c6",
		Name:            "Patron of the Vein",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.B06SelfETB,
				Targets:   effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				Key:       "Patron of the Vein — destroy target creature an opponent controls",
				Effect:    effects.DestroyFirstLegalTarget,
			},
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B33OpponentsCreatureDied(ev, source, g)
				},
				Key: "Patron of the Vein — exile the creature and put a +1/+1 counter on each Vampire you control",
				Effect: func(g *game.Game, item *game.StackItem) error {
					dead := item.Trigger.Event.CardID
					if z := g.FindCardZoneForEffect(dead); z != nil && z.Kind == game.ZoneGraveyard {
						if err := (effects.ExileTarget{Target: dead}).Apply(effects.NewContext(g, item)); err != nil {
							return err
						}
					}
					return effects.B17PutCounterOnEachVampireYouControl(g, item)
				},
			},
		},
	})
}
