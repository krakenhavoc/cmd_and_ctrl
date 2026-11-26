package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dead Reckoning — Sorcery {1}{B}{B} (Worldwake, S82 OLD batch 489, #5582):
//
//	"You may put target creature card from your graveyard on top of your
//	 library. If you do, Dead Reckoning deals damage equal to that card's
//	 power to target creature."
//
// Two target clauses announced together. As it resolves the card is read
// (its power, CR 608.2h) and put on top of the library; only if that
// happened is the damage dealt, to the second target if it is still legal.
// The tuck is a yes/no question asked of the caster.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a4b77f0-81ac-48a6-961b-bd63e99f74cf",
		Name:         "Dead Reckoning",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
			effects.TargetCreature("target creature")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.MayChoice{
				Question: "Dead Reckoning — put the creature card on top of your library? (If you do, it deals damage equal to its power to the other target.)",
				Then:     deadReckoningAnswer,
			}.Apply(ctx)
		},
	})
}

var deadReckoningAnswer = effects.MayYes("confirm/dead-reckoning-tuck", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	cards := ctx.ClauseTargets(0)
	if len(cards) == 0 || !ctx.IsTargetLegal(cards[0]) {
		return nil
	}
	card, ok := ctx.Game.LookupCardForEffect(cards[0].ID)
	if !ok {
		return nil
	}
	// The card's power is read before it moves (CR 608.2h) and rides the tail.
	power := ctx.Game.PowerAnywhereForEffect(card)
	return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{card.InstanceID}, game.TuckOptions{},
		effects.WithN(ctx.ItemTail(deadReckoningTucked), power))
})

// deadReckoningTucked is "If you do, Dead Reckoning deals damage equal to that
// card's power to target creature": only when the card really reached the
// library, to the second target if it is still legal.
var deadReckoningTucked = effects.CardTail("wwk/dead-reckoning-tucked", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || o.Tail.N <= 0 {
		return nil
	}
	for _, t := range ctx.ClauseTargets(1) {
		if ctx.IsTargetLegal(t) {
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: o.Tail.N}.Apply(ctx)
		}
	}
	return nil
})
