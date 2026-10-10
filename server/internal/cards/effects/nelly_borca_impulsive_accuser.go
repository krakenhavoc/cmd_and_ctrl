package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nelly Borca, Impulsive Accuser — Legendary Creature — Human
// Detective {2}{R}{W}, 2/4:
//
//	"Vigilance
//	 Whenever Nelly Borca attacks, suspect target creature. Then goad
//	 all suspected creatures. (A suspected creature has menace and
//	 can't block.)
//	 Whenever one or more creatures an opponent controls deal combat
//	 damage to one or more of your opponents, you and the controller of
//	 those creatures each draw a card."
//
// The attack trigger suspects its target, then goads every suspected
// creature on the battlefield, yours included: a creature you goad
// must attack each combat and attack a player other than you, which
// it would anyway. A target that is gone by resolution counters the
// ability (CR 608.2b), so nothing is goaded.
//
// The damage trigger is a "one or more" trigger keyed per DEALING
// player (#2733): one opponent's creatures hitting your opponents in a
// combat damage step is one trigger, however many creatures and
// players; two opponents' creatures connecting in the same step are two
// triggers, one for each controller. The dealing creature's controller
// is the damage event's Actor, stamped when the damage is dealt, so it
// is still known after the creature dies to that same combat damage.
// First-strike and regular damage are two steps, so two triggers.
//
// No simplification.
func init() {
	const drawLabel = "Nelly Borca — you and the controller of those creatures each draw a card"
	Register(Spec{
		OracleID:        "7ef5b2b6-86da-4e4a-9456-dcbb878936c4",
		Name:            "Nelly Borca, Impulsive Accuser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Nelly Borca — suspect target creature, then goad all suspected creatures", nellyBorcaAccuse),
				TargetCreature("target creature"),
			),
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return creatureDealtCombatDamageToAnOpponentOf(ev, source.Controller, g) &&
						ev.Actor != uuid.Nil && ev.Actor != source.Controller
				},
				Key:          drawLabel,
				OncePerBatch: true,
				BatchKey: func(ev game.Event, _ *game.Card, _ *game.Game) string {
					return ev.Actor.String()
				},
				Effect: nellyBorcaBothDraw,
			},
		},
	})
}

// nellyBorcaAccuse suspects the target, then goads every suspected
// creature.
func nellyBorcaAccuse(g *game.Game, item *game.StackItem) error {
	if err := SuspectEachLegalTarget(g, item); err != nil {
		return err
	}
	return GoadAllMatching(NewContext(g, item), func(_ *game.Game, c game.Card) bool { return c.Suspected })
}

// nellyBorcaBothDraw draws a card for Nelly's controller and one for
// the player whose creatures dealt the damage, read off the carried
// trigger event.
func nellyBorcaBothDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: ctx.Controller(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	return DrawCards{Player: ctx.Trigger().Event.Actor, N: 1}.Apply(ctx)
}
