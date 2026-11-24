package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kardur's Vicious Return — Enchantment — Saga {2}{B}{R} (Kaldheim, #5505):
//
//	"I — You may sacrifice a creature. When you do, this Saga deals 3
//	 damage to any target.
//	 II — Each player discards a card.
//	 III — Return target creature card from your graveyard to the
//	 battlefield. Put a +1/+1 counter on it. It gains haste until your next
//	 turn."
//
// Chapter I is Grist's sacrifice with a reflexive trigger (CR 603.12): the
// target is chosen only if a creature was sacrificed, and the Saga is the
// damage source. Chapter II is every player's choice of discard, active
// player first. Chapter III is a one-card return whose tail adds the counter
// and the haste (until the controller's next turn) to the creature that
// actually entered.
//
// No simplification.
func init() {
	const name = "Kardur's Vicious Return"
	effects.Register(effects.Spec{
		OracleID:     "6ca1d5ca-cc97-4783-9c0a-f47097f35ba8",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "you may sacrifice a creature; when you do, this Saga deals 3 damage to any target"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, name+" — sacrifice a creature to deal 3 damage to any target?",
						effects.Creature(), false, kardursViciousReturnSacrificed)
				}),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "each player discards a card"),
				effects.TspB1EachPlayerDiscardsACard(name+" — discard a card")),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "return target creature card from your graveyard to the battlefield; put a +1/+1 counter on it; it gains haste until your next turn"),
				effects.TargetCardInGraveyard("target creature card in your graveyard", effects.YouOwn(), effects.Creature()),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.ReturnFromGraveyardTogether{
						Targets:    []uuid.UUID{id},
						Controller: ctx.Controller(),
						ThenTail:   kardursViciousReturnBoost,
					}.Apply(ctx)
				}),
		},
	})
}

var kardursViciousReturnSacrificed = effects.RegisterSacrificedOne("prompt-run/kardurs-vicious-return-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.WhenYouDo("Kardur's Vicious Return — 3 damage to any target", kardursViciousReturnDamageBody).Apply(ctx)
})

var kardursViciousReturnDamageBody = game.ReflexiveBody("kardurs-vicious-return/damage", effects.SimpleBody(effects.DealDamageToTheTarget(3)),
	effects.ConstTargets(effects.TargetAny))

// kardursViciousReturnBoost puts the +1/+1 counter on the creature that
// entered and gives it haste until its controller's next turn.
var kardursViciousReturnBoost = effects.TogetherTail("together/kardurs-vicious-return-boost", func(ctx *effects.Context, entered []uuid.UUID, _ int) error {
	for _, id := range entered {
		if !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
		if err := (effects.ScopedEffectFor{
			Target:   id,
			Mods:     []game.Mod{game.AddKeywordsMod("haste")},
			Duration: effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
			Label:    "Kardur's Vicious Return — haste until your next turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
