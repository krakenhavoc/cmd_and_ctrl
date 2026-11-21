package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ral, Monsoon Mage // Ral, Leyline Prodigy — Modern Horizons 3 (#5948), a
// transforming legendary creature whose back face is a planeswalker.
//
// Front face, Legendary Creature — Human Wizard {1}{R}, 1/3:
//
//	"Instant and sorcery spells you cast cost {1} less to cast.
//	 Whenever you cast an instant or sorcery spell during your turn, flip a
//	 coin. If you lose the flip, Ral deals 1 damage to you. If you win the
//	 flip, you may exile Ral. If you do, return him to the battlefield
//	 transformed under his owner's control."
//
// Back face, Legendary Planeswalker — Ral, loyalty 2:
//
//	"Ral enters with an additional loyalty counter on him for each instant
//	 and sorcery spell you've cast this turn.
//	 +1: Until your next turn, instant and sorcery spells you cast cost {1}
//	     less to cast.
//	 −2: Ral deals 2 damage divided as you choose among one or two targets.
//	     Draw a card if you control a blue permanent other than Ral.
//	 −8: Exile the top eight cards of your library. You may cast instant and
//	     sorcery spells from among them this turn without paying their mana
//	     costs."
//
// The back face registers under "<oracle_id>#1" (game.CatalogKey). The cost
// reduction is Goblin Electromancer's; the +1 is Spellbinding Soprano's
// promise on an until-your-next-turn duration. The coin is a keyed flip
// (Goblin Psychopath's shape): a lost flip is 1 damage to you, a won one asks
// whether to exile Ral, and the return is ExileAndReturnTransformed, a new
// object (CR 400.7) that enters with its printed loyalty plus a counter for
// each instant and sorcery spell cast this turn — the spell that triggered the
// flip is on the stack and counts (CR 601.2i). The −2 is Twin Bolt's division,
// and the −8 is Hazoret's-Undying-Fury style: each exiled card gets its own
// free, cast-only permission lasting the turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     ralMonsoonMageOracle,
		Name:         "Ral, Monsoon Mage",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.",
				YourSpell(), InstantOrSorcerySpell()),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, AllOf(YouCast(Or(Instant(), Sorcery())),
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return IsYourTurn(g, source.Controller)
				}),
				"Ral, Monsoon Mage — flip a coin", ralMonsoonMageFlip),
		},
	})

	Register(Spec{
		OracleID:        ralMonsoonMageOracle + "#1",
		Name:            "Ral, Leyline Prodigy",
		Completeness:    CompletenessFull,
		StartingLoyalty: 2,
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterLoyalty, func(g *game.Game, src *game.Card) int {
				return ralInstantsAndSorceriesCastThisTurn(g, src.Controller)
			}, "Ral, Leyline Prodigy: enters with an additional loyalty counter for each instant and sorcery spell you've cast this turn"),
		},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Until your next turn, instant and sorcery spells you cast cost {1} less to cast.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					g.GrantNextSpellPromiseForEffect(ctx.Controller(), game.NextSpellPromise{
						Filter:             game.PermissionFilter{InstantOrSorceryOnly: true},
						Reduce:             1,
						EveryMatchingSpell: true,
						Text:               "Instant and sorcery spells you cast cost {1} less to cast.",
					}, "Ral, Leyline Prodigy", ctx.Source(), DurationUntilYourNextTurn(ctx, ctx.Controller()))
					return nil
				},
			},
			{
				Label:   "−2: Ral deals 2 damage divided as you choose among one or two targets. Draw a card if you control a blue permanent other than Ral.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetAny().WithCount(1, 2).Dividing(Divide(2)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := DealDividedDamage(ctx); err != nil {
						return err
					}
					if !ralControlsBlueOther(g, ctx.Controller(), ctx.Source()) {
						return nil
					}
					return DrawCards{N: 1}.Apply(ctx)
				},
			},
			{
				Label: "−8: Exile the top eight cards of your library. You may cast instant and sorcery spells from among them this turn without paying their mana costs.",
				Cost:  LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					window := g.UntilEndOfTurnDuration()
					return ExileTopWithPermission{
						From:     ctx.Controller(),
						GrantTo:  ctx.Controller(),
						N:        8,
						CastOnly: true,
						Free:     true,
						Window:   &window,
						Filter:   game.PermissionFilter{InstantOrSorceryOnly: true},
					}.Apply(ctx)
				},
			},
		},
	})
}

// ralMonsoonMageOracle is shared by both faces.
const ralMonsoonMageOracle = "16d40f56-691b-42a7-a5ab-157568921e03"

// ralInstantsAndSorceriesCastThisTurn counts the instant and sorcery spells
// `player` has cast this turn, the one still on the stack included.
func ralInstantsAndSorceriesCastThisTurn(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventCast || ev.Actor != player {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && (c.IsInstant() || c.IsSorcery()) {
			n++
		}
	}
	return n
}

// ralControlsBlueOther is "you control a blue permanent other than Ral".
func ralControlsBlueOther(g *game.Game, player, ral uuid.UUID) bool {
	g.RecomputeLayersIfStaleLocked()
	for _, c := range g.Battlefield.Cards {
		if c.Controller != player || c.InstanceID == ral {
			continue
		}
		for _, col := range c.EffectiveColors() {
			if col == "U" {
				return true
			}
		}
	}
	return false
}

// ralMonsoonMageFlip flips the coin; the outcome is a keyed continuation, so a
// table waiting on the flip is a restore point.
func ralMonsoonMageFlip(g *game.Game, item *game.StackItem) error {
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Ral, Monsoon Mage — flip a coin",
		ThenKey:  ralMonsoonMageFlipped,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// ralMonsoonMageFlipped: a lost flip deals 1 damage to the controller; a won
// one offers the exile-and-transform. On-disk key.
var ralMonsoonMageFlipped = CoinFlipThen("coin-flip/ral-monsoon-mage", func(ctx *Context, r game.CoinFlipped) error {
	if !wonFirstFlip(r) {
		return DealDamage{Source: ctx.Source(), Target: ctx.Controller(), Amount: 1}.Apply(ctx)
	}
	return MayChoice{
		Question: "Ral, Monsoon Mage — exile Ral, then return him transformed?",
		YesLabel: "Exile Ral and return him transformed",
		NoLabel:  "Don't",
		Then:     ralMonsoonMageTransform,
	}.Apply(ctx)
})

var ralMonsoonMageTransform = MayYes("confirm/ral-monsoon-mage-transform", func(ctx *Context, _ game.ConfirmAnswer) error {
	return ExileAndReturnTransformed{Target: ctx.Source()}.Apply(ctx)
})
