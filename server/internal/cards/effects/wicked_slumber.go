package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wicked Slumber — Instant {3}{U} (March of the Machine):
//
//	"Convoke
//	 Tap up to two target creatures. Put a stun counter on either of them.
//	 Then put a stun counter on either of them."
//
// The creatures still legal are tapped (CR 608.2b), then the controller
// chooses where each of two stun counters goes: "either of them" is asked
// twice, so both may land on one creature or one on each. With a single
// creature left, both go on it without a question. The second counter
// waits for the first choice (an on-disk key per step).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b27b794-580c-472d-b63e-99b271eb115d",
		Name:         "Wicked Slumber",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Targets:      TargetCreature("up to two target creatures", Creature()).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var tapped []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard || slices.Contains(tapped, t.ID) {
					continue
				}
				if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				tapped = append(tapped, t.ID)
			}
			switch len(tapped) {
			case 0:
				return nil
			case 1:
				return AddCounter{Target: tapped[0], Kind: game.CounterStun, N: 2}.Apply(ctx)
			}
			wickedSlumberAsk(ctx, tapped, wickedSlumberFirst,
				"Wicked Slumber — put a stun counter on which of them?")
			return nil
		},
	})
}

// wickedSlumberAsk asks which of the tapped creatures gets the next stun
// counter.
func wickedSlumberAsk(ctx *Context, tapped []uuid.UUID, then game.ChooseCardsThen, question string) {
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: question,
		Cards:    tapped,
		Min:      1,
		Max:      1,
		ThenKey:  then,
		Carry:    tapped,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(ctx.Item),
	})
}

// wickedSlumberStunPicked puts a stun counter on the creature picked, if it
// is still one of the two the spell tapped and still on the battlefield.
func wickedSlumberStunPicked(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 || !slices.Contains(r.Carry, r.Picked[0]) || !onBattlefield(ctx.Game, r.Picked[0]) {
		return nil
	}
	return AddCounter{Target: r.Picked[0], Kind: game.CounterStun, N: 1}.Apply(ctx)
}

// wickedSlumberSecond is "Then put a stun counter on either of them". An
// on-disk key.
var wickedSlumberSecond = askedPick("choose-cards/wicked-slumber-second-stun", wickedSlumberStunPicked)

// wickedSlumberFirst is "Put a stun counter on either of them", then asks
// for the second. An on-disk key.
var wickedSlumberFirst = askedPick("choose-cards/wicked-slumber-first-stun", func(ctx *Context, r game.CardsChosen) error {
	if err := wickedSlumberStunPicked(ctx, r); err != nil {
		return err
	}
	wickedSlumberAsk(ctx, r.Carry, wickedSlumberSecond, "Wicked Slumber — then put a stun counter on which of them?")
	return nil
})
