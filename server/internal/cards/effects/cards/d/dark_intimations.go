package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark Intimations — {2}{U}{B}{R} Sorcery (Aether Revolt, #5536):
//
//	"Each opponent sacrifices a creature or planeswalker of their choice,
//	 then discards a card. You return a creature or planeswalker card from
//	 your graveyard to your hand, then draw a card.
//	 When you cast a Bolas planeswalker spell, exile this card from your
//	 graveyard. That planeswalker enters with an additional loyalty counter
//	 on it."
//
// The spell is Cruel Ultimatum's walk: every opponent sacrifices (their own
// choice), then every opponent discards, then the controller picks a card to
// take back (no target, mandatory while a creature or planeswalker card is
// there) and draws. The second ability is a graveyard cast trigger, the shape
// Kozilek's Return uses; it is not optional, so the card is exiled whenever a
// Bolas planeswalker spell is cast, and that spell is marked to enter with
// one more loyalty counter (CR 614.1c) through the cast-trigger entry mark.
//
// No simplification.
func init() {
	const label = "Dark Intimations — exile this card from your graveyard; that planeswalker enters with an additional loyalty counter"
	bolasPlaneswalker := effects.And(effects.Planeswalker(), effects.HasSubtype("Bolas"))
	cast := effects.WheneverYouCastThatEntersWithCounters("Dark Intimations", bolasPlaneswalker, game.CounterLoyalty,
		effects.EntersWithOneMore, label)
	mark := cast.Effect
	cast.Effect = func(g *game.Game, item *game.StackItem) error {
		if z := g.FindCardZoneForEffect(item.SourceCardID); z != nil && z.Kind == game.ZoneGraveyard {
			if err := (effects.ExileTarget{Target: item.SourceCardID}).Apply(effects.NewContext(g, item)); err != nil {
				return err
			}
		}
		return mark(g, item)
	}
	effects.Register(effects.Spec{
		OracleID:     "228cb63f-0b93-4be5-94e5-4eaae9112eeb",
		Name:         "Dark Intimations",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.EachOpponentSacrificesACreatureOrPlaneswalker(ctx.Game, ctx.Item)
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.EachOpponentDiscardsOne(ctx.Game, ctx.Item) }),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ReturnChosenFromGraveyard{
					Player:   ctx.Controller(),
					Question: "Dark Intimations — return a creature or planeswalker card from your graveyard to your hand",
					Match:    func(c game.Card) bool { return c.IsCreature() || c.IsPlaneswalker() },
					Min:      1,
					Max:      1,
					ToHand:   true,
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.DrawCards{N: 1}.Apply(ctx) }),
		),
		Triggered: []game.TriggeredAbility{effects.InGraveyard(cast)},
	})
}
