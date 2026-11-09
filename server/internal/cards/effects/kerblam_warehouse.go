package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kerblam! Warehouse — Plane — Kandoka (Doctor Who, #4870):
//
//	"Whenever one or more creatures you control deal combat damage to a
//	 player, create a Treasure token.
//	 Whenever chaos ensues, until your next turn, noncreature artifacts
//	 you control gain "{T}, Sacrifice this artifact: Flip a coin. If you
//	 win the flip, this artifact deals 3 damage to any target.""
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The first ability is one trigger per combat damage step and player. The
// chaos ability (CR 311.7) grants a catalog bundle (ADR 0093) to the
// noncreature artifacts the planar controller controls as it resolves, a
// set locked then (CR 611.2c), until that player's next turn. The granted
// ability's target is chosen as it is activated, before the flip (the
// ruling), and the sacrificed artifact deals the damage as it last existed
// on the battlefield (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71bff56b-733c-4d74-9e0c-85371a8b3787",
		Name:         "Kerblam! Warehouse",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: kerblamWarehouseGrant,
			Activated: []ActivatedAbility{{
				Label:   "{T}, Sacrifice this artifact: Flip a coin. If you win the flip, this artifact deals 3 damage to any target.",
				Cost:    Plus(TapCost(), SacrificeThis()),
				Targets: TargetAny(),
				Purpose: ForTargets(DamageToTarget(0, 3)),
				Effect:  kerblamWarehouseFlip,
			}},
			Text: "{T}, Sacrifice this artifact: Flip a coin. If you win the flip, this artifact deals 3 damage to any target.",
		}},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Kerblam! Warehouse — create a Treasure", Do(CreateToken{Template: TreasureToken(), N: 1})),
			WheneverChaosEnsues("Kerblam! Warehouse — until your next turn, noncreature artifacts you control gain \"{T}, Sacrifice this artifact: flip a coin; win: 3 damage to any target\"",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return GrantAbilitiesFor{
						Match:    And(Artifact(), Not(Creature()), YouControl()),
						Keys:     []string{kerblamWarehouseGrant},
						Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
						Label:    "Kerblam! Warehouse — {T}, sacrifice: flip a coin, 3 damage",
					}.Apply(ctx)
				}),
		},
	})
}

// kerblamWarehouseGrant is the granted bundle's key. An on-disk key.
const kerblamWarehouseGrant = "kerblam-warehouse/flip-for-3"

// kerblamWarehouseFlip flips the coin; the call is keyed, so a table
// waiting on it is a restore point.
func kerblamWarehouseFlip(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Kerblam! Warehouse — call the coin flip",
		ThenKey:  kerblamWarehouseFlipped,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// kerblamWarehouseFlipped is the settled call: a won flip deals 3 damage
// from the artifact to the ability's target. An on-disk key: never renamed,
// never reused.
var kerblamWarehouseFlipped = CoinFlipThen("coin-flip/kerblam-warehouse", func(ctx *Context, r game.CoinFlipped) error {
	if !wonFirstFlip(r) {
		return nil
	}
	return DealDamageToTheTarget(3)(ctx.Game, ctx.Item)
})
