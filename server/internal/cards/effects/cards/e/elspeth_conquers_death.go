package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elspeth Conquers Death — Enchantment — Saga {3}{W}{W} (Theros Beyond Death, #5514):
//
//	"I — Exile target permanent an opponent controls with mana value 3 or greater.
//	 II — Noncreature spells your opponents cast cost {2} more to cast until your
//	 next turn.
//	 III — Return target creature or planeswalker card from your graveyard to the
//	 battlefield. Put a +1/+1 counter or a loyalty counter on it."
//
// Chapter II is a scoped record (game.ModNoncreatureSpellsCostMoreForOpponents,
// #5514, beside Tax Collector's ModSpellsCostMoreForOpponents) read by the cast
// pricer, which outlives the Saga and ends as the controller's next turn begins.
// Chapter III returns the card, and once it has entered gives it a +1/+1 counter
// if it is a creature or a loyalty counter if it is a planeswalker: no printed
// card is both, so the "or" is never a choice the controller makes.
//
// No simplification.
func init() {
	const name = "Elspeth Conquers Death"
	effects.Register(effects.Spec{
		OracleID:     "40c88392-2aa7-4e12-86a4-8d680219e670",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "exile target permanent an opponent controls with mana value 3 or greater"),
				effects.TargetPermanent("target permanent an opponent controls with mana value 3 or greater", effects.OpponentControls(), effects.ManaValueGE(3)),
				effects.ExileFirstLegalCardTarget),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "noncreature spells your opponents cast cost {2} more until your next turn"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					g.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeGame, ctx.Controller(),
						[]game.Mod{game.NoncreatureSpellsCostMoreForOpponentsMod(2)},
						effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
						"Elspeth Conquers Death — noncreature spells your opponents cast cost {2} more to cast")
					return nil
				}),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "return target creature or planeswalker card from your graveyard to the battlefield; put a +1/+1 counter or a loyalty counter on it"),
				effects.TargetCardInGraveyard("target creature or planeswalker card from your graveyard", effects.YouOwn(), effects.Or(effects.Creature(), effects.Planeswalker())),
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnFromGraveyardTogether{
						Targets:  effects.LegalTargetCardIDs(effects.NewContext(g, item)),
						ThenTail: elspethConquersDeathCounter,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var elspethConquersDeathCounter = effects.TogetherTail("elspeth-conquers-death/counter-on-it",
	func(ctx *effects.Context, entered []uuid.UUID, _ int) error {
		for _, id := range entered {
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || !effects.OnBattlefield(ctx.Game, id) {
				continue
			}
			kind := game.CounterPlusOne
			if !c.IsCreature() {
				kind = game.CounterLoyalty
			}
			if err := (effects.AddCounter{Target: id, Kind: kind, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
