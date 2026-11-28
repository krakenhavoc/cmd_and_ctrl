package cardsfl

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forgotten Lore — Sorcery {G} (Ice Age, S82 OLD, #5636):
//
//	"Target opponent chooses a card in your graveyard. You may pay {G}. If you do, repeat this process except
//	 that opponent can't choose a card already chosen for Forgotten Lore. Then put the last chosen card into
//	 your hand."
//
// A chain of keyed prompts: the opponent picks a card in your graveyard, you may pay {G} (CR 118.12), and a "yes"
// asks again with the cards already chosen left out. When you decline, or no unchosen card is left, the last card
// chosen goes to your hand if it is still in your graveyard (CR 400.7). An empty graveyard does nothing.
//
// No simplification.
func init() {
	// The two continuations name each other through forgottenLoreAsk, so they are registered here rather than in
	// their declarations (an initialization cycle otherwise), as manifestDreadAgain is.
	forgottenLoreChose = effects.AskedPick("choose-cards/s82-old-b844-forgotten-lore-chose", forgottenLoreChoseBody)
	forgottenLoreRepeat = effects.PayUnlessBranches("pay-unless/s82-old-b844-forgotten-lore-repeat",
		func(ctx *effects.Context, r game.PayUnlessAnswer) error {
			if len(r.Carry) == 0 {
				return nil
			}
			return forgottenLoreAsk(ctx, r.Carry[0], slices.Clone(r.Carry[1:]))
		},
		func(ctx *effects.Context, r game.PayUnlessAnswer) error {
			if len(r.Carry) < 2 {
				return nil
			}
			return forgottenLoreFinish(ctx, slices.Clone(r.Carry[1:]))
		})
	effects.Register(effects.Spec{
		OracleID:     "ca9f538d-4191-448a-870d-f645b7c37c1b",
		Name:         "Forgotten Lore",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve:    forgottenLoreStart,
	})
}

// forgottenLoreChose is the opponent's choice: the card joins those already chosen, then the caster may pay {G}
// to repeat. A graveyard with nothing left to choose ends the process.
var forgottenLoreChose game.ChooseCardsThen

// forgottenLoreRepeat: paid, the opponent chooses again (not a card already chosen); declined, the last chosen
// card goes to the caster's hand.
var forgottenLoreRepeat game.PayUnlessElse

// forgottenLoreStart starts the process: the target opponent chooses a card in the caster's graveyard.
func forgottenLoreStart(_ *game.StackItem, ctx *effects.Context) error {
	opp := effects.TargetedPlayer(ctx)
	if opp == uuid.Nil {
		return nil
	}
	return forgottenLoreAsk(ctx, opp, nil)
}

// forgottenLoreAsk is one round: `opp` chooses a card in the caster's graveyard that is not already in `chosen`.
// With none left to choose, the process ends and the last chosen card is taken.
func forgottenLoreAsk(ctx *effects.Context, opp uuid.UUID, chosen []uuid.UUID) error {
	me := ctx.Controller()
	cands := effects.GraveyardCardIDs(ctx, me, func(c game.Card) bool { return !slices.Contains(chosen, c.InstanceID) })
	if len(cands) == 0 {
		return forgottenLoreFinish(ctx, chosen)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    opp,
		FromPlayer: me,
		Source:     ctx.Source(),
		Question:   "Forgotten Lore — choose a card in the opponent's graveyard",
		Cards:      cands,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    forgottenLoreChose,
		Carry:      append([]uuid.UUID{opp}, chosen...),
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}

func forgottenLoreChoseBody(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	opp, chosen := r.Carry[0], slices.Clone(r.Carry[1:])
	if len(r.Picked) > 0 {
		chosen = append(chosen, r.Picked[0])
	}
	me := ctx.Controller()
	more := effects.GraveyardCardIDs(ctx, me, func(c game.Card) bool { return !slices.Contains(chosen, c.InstanceID) })
	if len(more) == 0 || len(chosen) == 0 {
		return forgottenLoreFinish(ctx, chosen)
	}
	return effects.MayPay{
		Chooser:  me,
		Cost:     "{G}",
		Question: "Forgotten Lore — pay {G} to have the opponent choose another card?",
		Then:     forgottenLoreRepeat,
		Carry:    append([]uuid.UUID{opp}, chosen...),
	}.Apply(ctx)
}

// forgottenLoreFinish puts the last chosen card into the caster's hand, if it is still in their graveyard.
func forgottenLoreFinish(ctx *effects.Context, chosen []uuid.UUID) error {
	if len(chosen) == 0 {
		return nil
	}
	last := chosen[len(chosen)-1]
	if z := ctx.Game.FindCardZoneForEffect(last); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return effects.ReturnFromGraveyard{Target: last, Dest: game.ZoneHand}.Apply(ctx)
}
