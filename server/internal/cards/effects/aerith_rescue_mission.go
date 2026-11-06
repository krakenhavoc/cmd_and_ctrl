package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aerith Rescue Mission — Sorcery {3}{W} (Final Fantasy, #4462):
//
//	"Choose one —
//	 • Take the Elevator — Create three 1/1 colorless Hero creature
//	   tokens.
//	 • Take 59 Flights of Stairs — Tap up to three target creatures. Put
//	   a stun counter on one of them."
//
// The second mode taps each target still legal as it resolves (CR
// 608.2b), then puts one stun counter on one of THOSE creatures: with one
// left there is nothing to choose, with two or three the caster picks.
// A target that left in response is neither tapped nor offered. The
// stun counter is the engine's (CR 122.1d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "14c3e48b-4c4e-4787-8c50-0e3aee19aa23",
		Name:         "Aerith Rescue Mission",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Take the Elevator — Create three 1/1 colorless Hero creature tokens."),
			Mode("Take 59 Flights of Stairs — Tap up to three target creatures. Put a stun counter on one of them.",
				TargetCreature("up to three target creatures").WithCount(0, 3)),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				return CreateToken{Template: TokenCard(jobSelectHeroToken), N: 3}.Apply(ctx)
			}
			return aerithRescueMissionStairs(ctx)
		},
	})
}

// aerithRescueMissionStairs taps the legal targets and puts a stun
// counter on one of them.
func aerithRescueMissionStairs(ctx *Context) error {
	var tapped []uuid.UUID
	for _, t := range ctx.ModeTargets(0) {
		if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) || slices.Contains(tapped, t.ID) {
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
		return AddCounter{Target: tapped[0], Kind: game.CounterStun, N: 1}.Apply(ctx)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Aerith Rescue Mission — put a stun counter on which of them?",
		Cards:    tapped,
		Min:      1,
		Max:      1,
		ThenKey:  aerithRescueMissionStun,
		Carry:    tapped,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// aerithRescueMissionStun puts the stun counter on the creature picked,
// if it is still one of the creatures the spell tapped. An on-disk key.
var aerithRescueMissionStun = askedPick("choose-cards/aerith-rescue-mission-stun", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 || !slices.Contains(r.Carry, r.Picked[0]) || !onBattlefield(ctx.Game, r.Picked[0]) {
		return nil
	}
	return AddCounter{Target: r.Picked[0], Kind: game.CounterStun, N: 1}.Apply(ctx)
})
