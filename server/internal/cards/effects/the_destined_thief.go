package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Destined Thief — Legendary Creature — Human Rogue {2}{U}, 2/2 (Final
// Fantasy, #4462):
//
//	"The Destined Thief can't be blocked.
//	 {U}, {T}: Another target creature you control can't be blocked this
//	 turn.
//	 Whenever one or more creatures you control deal combat damage to one
//	 or more players, draw a card, then discard a card. If you have a full
//	 party, instead draw three cards."
//
// The batch trigger is OncePerBatch over a creature of yours dealing
// combat damage to a player (archmages_newt's read), so a combat damage
// step is one trigger however many connect. The party is counted as it
// resolves (CR 700.8c): three cards and no discard with a full party, else
// the loot.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7857c7b6-9676-47ac-bb28-d0b4cbda5fa7",
		Name:         "The Destined Thief",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "The Destined Thief — draw a card, then discard a card, or draw three cards with a full party",
				Do(
					If{Cond: finB9FullParty, Then: []Applier{DrawCards{N: 3}}},
					If{Cond: finB9NoFullParty, Then: []Applier{finB9RunEffect(func(g *game.Game, item *game.StackItem) error {
						return b16DrawThenDiscard(g, item, 1)
					})}},
				))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{U}, {T}: Another target creature you control can't be blocked this turn.",
			Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    Plus(ManaCost("{U}"), TapCost()),
			Targets: Another(TargetCreature("another target creature you control", YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return RestrictUntilEOT{Target: target, Restrictions: game.CantBeBlocked, Label: "The Destined Thief — can't be blocked"}.Apply(ctx)
			},
		}},
	})
}
