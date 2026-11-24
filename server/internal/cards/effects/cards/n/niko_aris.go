package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Niko Aris — Legendary Planeswalker — Niko {X}{W}{U}{U}, loyalty 3:
//
//	"When Niko Aris enters, create X Shard tokens. (They're enchantments with
//	 '{2}, Sacrifice this token: Scry 1, then draw a card.')
//	 +1: Up to one target creature you control can't be blocked this turn.
//	 Whenever that creature deals damage this turn, return it to its owner's
//	 hand.
//	 −1: Niko Aris deals 2 damage to target tapped creature for each card
//	 you've drawn this turn.
//	 −1: Create a Shard token."
//
// X is the cast's X, stamped on the enters trigger by the engine (XFromCast,
// CR 107.3m), so a Niko that was not cast makes none. The +1's second sentence
// is a CR 603.7 delayed trigger on the creature chosen (a "this turn" duration,
// ended by cleanup): the first time that creature deals damage it goes back to
// its owner's hand if it is still that object on the battlefield; with no
// target nothing is scheduled. The first −1 counts the cards its controller
// has drawn this turn when it resolves (CR 608.2h), Niko dealing the damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "781162c4-9c0f-436c-b9a5-214e60e009e4",
		Name:            "Niko Aris",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		XMatters:        true,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.Self,
			Key:       "Niko Aris — create X Shard tokens",
			XFromCast: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if ctx.X() <= 0 {
					return nil
				}
				return effects.CreateToken{Controller: ctx.Controller(), Template: effects.ShardToken(), N: ctx.X()}.Apply(ctx)
			},
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Up to one target creature you control can't be blocked this turn. Whenever that creature deals damage this turn, return it to its owner's hand.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
				Targets: effects.UpToOneTargetCreature("up to one target creature you control", effects.YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBeBlocked,
							Label: "Niko Aris — can't be blocked"}).Apply(ctx); err != nil {
							return err
						}
						return effects.DelayedOnEvent{
							Label:     "Niko Aris — return that creature to its owner's hand",
							On:        []game.EventKind{game.EventDealDamage},
							Condition: effects.OldB071TheListedObjectDealtDamageCondition,
							Cards:     []uuid.UUID{t.ID},
							Body:      effects.OldB071ReturnTheListedObjectToHandBody,
						}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "−1: Niko Aris deals 2 damage to target tapped creature for each card you've drawn this turn.",
				Cost:    effects.LoyaltyCost(-1),
				Targets: effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := effects.B15CardsDrawnThisTurn(g)[ctx.Controller()]
					if n <= 0 {
						return nil
					}
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2 * n}.Apply(ctx)
						}
					}
					return nil
				},
			},
			{
				Label:   "−1: Create a Shard token.",
				Cost:    effects.LoyaltyCost(-1),
				Purpose: game.Purpose{Tokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.ShardToken(), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
