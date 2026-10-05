package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stinkweed Imp — Creature — Imp {2}{B}, 1/2:
//
//	"Flying
//	 Whenever this creature deals combat damage to a creature, destroy
//	 that creature.
//	 Dredge 5 (If you would draw a card, you may mill five cards
//	 instead. If you do, return this card from your graveyard to your
//	 hand.)"
//
// The trigger is Kaldra Compleat's shape with destroy for exile: it
// names the damaged creature as the OBJECT it was when the trigger is
// built, so a creature that died to the damage, or died and came back,
// is not destroyed (CR 400.7), while one that survived (indestructible
// aside) is. Dredge 5 is dredge.go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e005bc76-4985-4ec6-b9f6-cf6d0d9f5df4",
		Name:            "Stinkweed Imp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{Dredge(5)},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: thisDealtCombatDamageToACreature,
			Key:       stinkweedImpDestroyLabel,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, stinkweedImpDestroyLabel)
				item.Params.Object = damagedCreatureRef(g, ev.Target)
				return item
			},
			Effect: destroyDamagedCreature,
		}},
	})
}

const stinkweedImpDestroyLabel = "Stinkweed Imp — destroy the creature it dealt combat damage to"

// destroyDamagedCreature is "destroy that creature": the object named on
// item.Params.Object, and only while it is still that object on the
// battlefield.
func destroyDamagedCreature(g *game.Game, item *game.StackItem) error {
	ref := item.Params.Object
	c, ok := g.LookupCardForEffect(ref.ID)
	if !ok || c.ObjectEpoch != ref.Epoch {
		return nil
	}
	if z := g.FindCardZoneForEffect(ref.ID); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return DestroyTarget{Target: ref.ID}.Apply(NewContext(g, item))
}
