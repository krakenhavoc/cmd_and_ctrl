package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Everything Pizza — Artifact — Food {2} (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"When this artifact enters, search your library for a basic land
//	 card, reveal it, put it into your hand, then shuffle.
//	 {2}{W}{U}{B}{R}{G}, {T}, Sacrifice this artifact: Target player gains
//	 3 life and draws a card. Each of your opponents discards a card. This
//	 artifact deals 3 damage to any target. Put three +1/+1 counters on up
//	 to one target creature."
//
// It is a Food without the usual Food ability: the oracle text prints
// none, so it has none.
//
// The activated ability has three target clauses (#764): target player,
// any target, and up to one target creature, each re-checked on its own
// at resolution (CR 608.2b), so a clause whose target has gone is
// skipped and the rest still happens. Its five instructions run in
// printed order as steps (Do), so the damage waits for the opponents'
// discard prompts to be answered (CR 608.2c). The artifact has been
// sacrificed, so the damage is dealt by it as it last existed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa481983-8f1a-4519-9371-5e7edcd3274f",
		Name:         "Everything Pizza",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Everything Pizza — search your library for a basic land card and put it into your hand",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Everything Pizza — a basic land card",
						Source:    item.SourceCardID,
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}{W}{U}{B}{R}{G}, {T}, Sacrifice this artifact: Target player gains 3 life and draws a card. Each of your opponents discards a card. This artifact deals 3 damage to any target. Put three +1/+1 counters on up to one target creature.",
			Cost:  Plus(ManaCost("{2}{W}{U}{B}{R}{G}"), TapCost(), SacrificeThis()),
			Targets: Clauses(
				TargetPlayer("target player"),
				TargetAny(),
				TargetCreature("up to one target creature").WithCount(0, 1),
			),
			Effect: Do(
				Instruction(everythingPizzaLifeAndDraw),
				Instruction(func(ctx *Context) error { return eachOpponentDiscardsOne(ctx.Game, ctx.Item) }),
				Instruction(everythingPizzaDamage),
				Instruction(everythingPizzaCounters),
			),
		}},
	})
}

// everythingPizzaLifeAndDraw is "target player gains 3 life and draws a
// card".
func everythingPizzaLifeAndDraw(ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	if err := (GainLife{Player: t.ID, Amount: 3}).Apply(ctx); err != nil {
		return err
	}
	return DrawCards{Player: t.ID, N: 1}.Apply(ctx)
}

// everythingPizzaDamage is "this artifact deals 3 damage to any target".
func everythingPizzaDamage(ctx *Context) error {
	t, ok := ctx.ClauseTarget(1)
	if !ok {
		return nil
	}
	return ctx.Game.DamageInstanceForEffect(func() error {
		return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}.Apply(ctx)
	})
}

// everythingPizzaCounters is "put three +1/+1 counters on up to one
// target creature".
func everythingPizzaCounters(ctx *Context) error {
	t, ok := ctx.ClauseTarget(2)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 3}.Apply(ctx)
}
