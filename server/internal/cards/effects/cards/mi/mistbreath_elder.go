package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mistbreath Elder — Creature — Frog Warrior {G}, 2/2 (Bloomburrow, #4467):
//
//	"At the beginning of your upkeep, return another creature you control
//	 to its owner's hand. If you do, put a +1/+1 counter on this creature.
//	 Otherwise, you may return this creature to its owner's hand."
//
// Asked as the trigger resolves, not as it goes on the stack (it targets
// nothing). With another creature under your control the return is not
// optional: you choose which one (a choose_cards prompt over the other
// creatures, checked against the battlefield as answered), and only if it
// really reached a hand does the Elder get its counter. "Otherwise" is the
// case where you control no other creature (or it could not be returned):
// then it is a yes/no, and a yes returns the Elder itself (the rulings).
// Both ride registered keys, so a table waiting on the answer is a
// restore point.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf9221f5-a4b2-4074-901c-0fa3c2bda80c",
		Name:         "Mistbreath Elder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Mistbreath Elder — return another creature you control to its owner's hand; if you do, +1/+1 counter on this creature; otherwise you may return this creature to its owner's hand",
				mistbreathElderUpkeep),
		},
	})
}

func mistbreathElderUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !effects.OnBattlefield(g, item.SourceCardID) || g.AbilitySourceIsNewObjectForEffect(item) {
		return nil
	}
	var others []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() && c.InstanceID != item.SourceCardID {
			others = append(others, c.InstanceID)
		}
	}
	if len(others) == 0 {
		return mistbreathElderMayReturnSelf(ctx)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     item.SourceCardID,
		Question:   "Mistbreath Elder — return another creature you control to its owner's hand",
		Cards:      others,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    mistbreathElderPicked,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

func mistbreathElderMayReturnSelf(ctx *effects.Context) error {
	return effects.MayChoice{
		Question: "Mistbreath Elder — return this creature to its owner's hand?",
		Then:     mistbreathElderAnswered,
	}.Apply(ctx)
}

// mistbreathElderPicked returns the chosen creature and, if it reached a
// hand, puts the counter on the Elder; if it did not, "otherwise" applies.
// An on-disk key.
var mistbreathElderPicked = effects.AskedPick("choose-cards/mistbreath-elder-return", func(ctx *effects.Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if effects.OnBattlefield(ctx.Game, id) {
			return effects.BounceToHand{Target: id, ThenTail: ctx.ItemTail(mistbreathElderBounced)}.Apply(ctx)
		}
	}
	return mistbreathElderMayReturnSelf(ctx)
})

// mistbreathElderBounced is "if you do, put a +1/+1 counter on this
// creature" once the bounce settled, and "otherwise" when the creature did
// not reach a hand. An on-disk key.
var mistbreathElderBounced = effects.CardTail("mistbreath-elder/after-bounce", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return mistbreathElderMayReturnSelf(ctx)
	}
	if !effects.OnBattlefield(ctx.Game, ctx.Source()) || ctx.Game.AbilitySourceIsNewObjectForEffect(ctx.Item) {
		return nil
	}
	return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
})

// mistbreathElderAnswered is the "otherwise, you may" answer: a yes returns
// the Elder. An on-disk key.
var mistbreathElderAnswered = effects.MayYes("confirm/mistbreath-elder-return-self", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	if !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return effects.BounceToHand{Target: ctx.Source()}.Apply(ctx)
})
