package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Beast, Deathless Prince — Legendary Creature — Demon {2}{B}{R}, 6/6
// (Doctor Who, #5948):
//
//	"When you cast this spell, gain control of target creature until end of
//	 turn. Untap it. It gains menace and haste until end of turn.
//	 The Beast enters tapped with six stun counters on it. (If it would
//	 become untapped, remove a stun counter from it instead.)
//	 Whenever a creature deals combat damage to its owner, untap The Beast and
//	 draw a card."
//
// The cast trigger is Abby's shape: it goes on the stack above the spell and
// the CASTER controls it. Its body is Act of Treason's (control, untap, menace
// and haste until end of turn). The entry is the shared enters-tapped-with-
// counters row. The third ability is one trigger per creature's combat damage
// to the player who OWNS it (CR 120.3, read off the damage event; the dealer's
// owner is last-known if it has since left), and untapping The Beast spends a
// stun counter instead while it has one (CR 122.1g, engine-side).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca187570-3634-4296-92b2-1fcd99d150e3",
		Name:         "The Beast, Deathless Prince",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedWithCounters(game.CounterStun, 6)},
		Triggered: []game.TriggeredAbility{
			{
				FromStack: true,
				Watches:   []game.EventKind{game.EventCast},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				Targets: TargetCreature("target creature"),
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "The Beast — gain control of target creature until end of turn, untap it, it gains menace and haste")
					item.Controller, item.Owner = ev.Actor, ev.Actor
					return item
				},
				Effect: theBeastSteal,
			},
			On(game.EventDealDamage, theBeastCreatureDamagedItsOwner,
				"The Beast — untap The Beast and draw a card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if onBattlefield(g, item.SourceCardID) && !sourceIsNewObject(g, item) {
						if err := (UntapTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
							return err
						}
					}
					return DrawCards{N: 1}.Apply(ctx)
				}),
		},
	})
}

func theBeastSteal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (GainControl{Target: id, Controller: item.Controller, Duration: DurationUntilEndOfTurn(ctx), Label: "The Beast — gain control until end of turn"}).Apply(ctx); err != nil {
		return err
	}
	if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: id, Keywords: []string{"menace", "haste"}, Label: "The Beast — menace and haste until end of turn"}.Apply(ctx)
}

// theBeastCreatureDamagedItsOwner is "a creature deals combat damage to its
// owner": combat damage from a creature to the player who owns that creature.
func theBeastCreatureDamagedItsOwner(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	return ok && dealer.IsCreature() && dealer.Owner == ev.Target
}
