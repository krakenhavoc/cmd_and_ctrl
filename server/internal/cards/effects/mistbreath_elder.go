package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "bf9221f5-a4b2-4074-901c-0fa3c2bda80c",
		Name:         "Mistbreath Elder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Mistbreath Elder — return another creature you control to its owner's hand; if you do, +1/+1 counter on this creature; otherwise you may return this creature to its owner's hand",
				mistbreathElderUpkeep),
		},
	})
}

func mistbreathElderUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !onBattlefield(g, item.SourceCardID) || g.AbilitySourceIsNewObjectForEffect(item) {
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

func mistbreathElderMayReturnSelf(ctx *Context) error {
	return MayChoice{
		Question: "Mistbreath Elder — return this creature to its owner's hand?",
		Then:     mistbreathElderAnswered,
	}.Apply(ctx)
}

// mistbreathElderPicked returns the chosen creature and, if it reached a
// hand, puts the counter on the Elder; if it did not, "otherwise" applies.
// An on-disk key.
var mistbreathElderPicked = askedPick("choose-cards/mistbreath-elder-return", func(ctx *Context, r game.CardsChosen) error {
	g := ctx.Game
	returned := false
	for _, id := range r.Picked {
		if !onBattlefield(g, id) {
			continue
		}
		if err := g.BounceToHandForEffect(id); err != nil {
			return err
		}
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneHand {
			returned = true
		}
	}
	if returned {
		if !onBattlefield(g, ctx.Source()) {
			return nil
		}
		return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}
	return mistbreathElderMayReturnSelf(ctx)
})

// mistbreathElderAnswered is the "otherwise, you may" answer: a yes returns
// the Elder. An on-disk key.
var mistbreathElderAnswered = MayYes("confirm/mistbreath-elder-return-self", func(ctx *Context, _ game.ConfirmAnswer) error {
	g := ctx.Game
	if !onBattlefield(g, ctx.Source()) {
		return nil
	}
	return g.BounceToHandForEffect(ctx.Source())
})
