package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Disorienting Choice — Sorcery {3}{G} (Duskmourn Commander):
//
//	"For each opponent, choose up to one target artifact or enchantment
//	 that player controls. For each permanent chosen this way, its
//	 controller may exile it. Then if one or more of the chosen
//	 permanents are still on the battlefield, you search your library for
//	 up to that many land cards, put them onto the battlefield tapped,
//	 then shuffle."
//
// One "up to one" target per opponent (perOpponentClauses), each bound to
// its opponent. Each chosen permanent's controller is asked in turn order
// whether to exile it; a decline leaves it. "That many" counts the chosen
// permanents still on the battlefield once everyone has answered (and a
// permanent that left in response is not a legal target at all).
func init() {
	Register(Spec{
		OracleID:     "7aa34ef4-5e11-491b-8dbe-6ef80e0ad26c",
		Name:         "Disorienting Choice",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1, Lands: 1, Answers: game.AnswerRemove},
		Targets:      perOpponentClauses(game.MaxPlayers-1, "artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var picks []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					picks = append(picks, t.ID)
				}
			}
			return disorientingChoiceAsk(ctx, picks, picks)
		},
	})
}

// disorientingChoiceAsk asks the controller of the next chosen permanent
// whether to exile it; `all` rides along as the chosen set.
func disorientingChoiceAsk(ctx *Context, pending, all []uuid.UUID) error {
	for len(pending) > 0 {
		c, ok := disorientingChoiceOnBattlefield(ctx, pending[0])
		if ok {
			return MayChoice{
				Player:   c.Controller,
				Question: "Disorienting Choice — exile " + c.Name + "? (If you don't, the caster searches for a land.)",
				Then:     disorientingChoiceAnswer,
				Carry:    packIDLists(pending[:1], pending[1:], all),
			}.Apply(ctx)
		}
		pending = pending[1:]
	}
	return disorientingChoiceSearch(ctx, all)
}

var disorientingChoiceAnswer game.ConfirmThen

func init() {
	disorientingChoiceAnswer = MayAnswer("confirm/disorienting-choice-exile", func(ctx *Context, r game.ConfirmAnswer) error {
		lists := unpackIDLists(r.Carry)
		if len(lists) < 3 {
			return nil
		}
		current, rest, all := lists[0], lists[1], lists[2]
		if r.Accepted && len(current) == 1 {
			if err := (ExileTarget{Target: current[0]}).Apply(ctx); err != nil {
				return err
			}
		}
		return disorientingChoiceAsk(ctx, rest, all)
	})
}

// disorientingChoiceSearch counts the chosen permanents still on the
// battlefield and searches for up to that many lands.
func disorientingChoiceSearch(ctx *Context, all []uuid.UUID) error {
	n := 0
	for _, id := range all {
		if _, ok := disorientingChoiceOnBattlefield(ctx, id); ok {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:        ctx.Controller(),
		Source:        ctx.Source(),
		Pred:          b03IsLandCard,
		Dest:          game.ZoneBattlefield,
		Limit:         n,
		UpTo:          true,
		TappedOnEntry: true,
		Shuffle:       true,
		Reason:        "Disorienting Choice — up to that many lands onto the battlefield tapped",
	})
}

func disorientingChoiceOnBattlefield(ctx *Context, id uuid.UUID) (game.Card, bool) {
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}
