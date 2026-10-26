package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Transpose — Instant {2}{B}:
//
//	"Draw a card, then discard a card. You lose 1 life. If this spell
//	 was cast from your hand, create a 0/1 black Wizard creature token
//	 with "Whenever you cast a noncreature spell, this token deals 1
//	 damage to each opponent."
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// The life loss and the token wait for the discard (DiscardPrompt.Then),
// in printed order. "Cast from your hand" is the zone the spell was
// cast from (StackItem.CastFromZone), so the rebound cast from exile
// makes no Wizard. The Wizard is the catalog's printed token
// (NoncreatureCastWizardToken), ability and all.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "77d5f298-06b4-49d1-9b33-f1f176665ba2",
		Name:            "Transpose",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		OnResolve: ResolveSteps(
			// The discard waits for a prompt the draw asks (CR 608.2c, #4155).
			DrawCards{N: 1},
			Instruction(transposeDiscard),
		),
	})
}

// transposeDiscard is the discard, and the life loss and Wizard it leads to.
// Those two are continuation steps (#4162): the Wizard waits for a prompt the
// life loss asks (a CR 616 order of two life-loss replacements).
func transposeDiscard(ctx *Context) error {
	item := ctx.Item
	// Whether the spell was cast from the hand rides the prompt as data: the
	// source card is the carry when it was, and nothing is when it was not
	// (cast from exile by rebound).
	var fromHand []uuid.UUID
	if item.CastFromZone == game.ZoneHand {
		fromHand = []uuid.UUID{item.SourceCardID}
	}
	return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player: ctx.Controller(),
		Source: item.SourceCardID,
		N:      1,
	}, transposeDiscarded, fromHand)
}

// transposeDiscarded is the clause after the discard, keyed (#4253): a
// table waiting on the discard is a restore point. It runs against a stand-in
// item that remembers only what the wizard clause reads, the zone the spell
// was cast from.
var transposeDiscarded = game.RegisterDiscardThen("choose-cards/transpose-after-discard", func(g *game.Game, r game.DiscardsDone) error {
	ctx := keyedContext(g, r.Player, r.Source)
	if len(r.Carry) > 0 {
		ctx.Item.CastFromZone = game.ZoneHand
	}
	return transposeAfterDiscard.Run(ctx)
})

// transposeAfterDiscard is "You lose 1 life. If this spell was cast from your
// hand, create a 0/1 black Wizard creature token …".
var transposeAfterDiscard = RegisterThenSteps("then-steps/transpose-after-discard",
	Instruction(func(ctx *Context) error {
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -1)
	}),
	Instruction(func(ctx *Context) error {
		if ctx.Item.CastFromZone != game.ZoneHand {
			return nil
		}
		return CreateToken{Controller: ctx.Controller(), Template: NoncreatureCastWizardToken(), N: 1}.Apply(ctx)
	}),
)
