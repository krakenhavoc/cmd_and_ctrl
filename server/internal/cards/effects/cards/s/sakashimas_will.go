package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sakashima's Will — Sorcery {3}{U} (Commander Legends):
//
//	"Choose one. If you control a commander as you cast this spell, you may
//	 choose both instead.
//	 • Target opponent chooses a creature they control. You gain control of it.
//	 • Choose a creature you control. Each other creature you control becomes a
//	   copy of that creature until end of turn."
//
// The count is the Will cycle's (ChooseOne(...).OrUpToIf(2, YouControlACommander)).
// The first bullet asks the targeted opponent (an opponent with one creature
// is not asked; one with none loses nothing); the creature they pick is
// gained for good (CR 611.2a). The second asks you which of your creatures is
// the model, then every other creature you control, read as the copy begins
// and locked then (CR 611.2c), becomes a copy of it until end of turn (one
// duration copy over many permanents, Mirrorweave's). The bullets run in
// printed order, so a stolen creature is among "each other creature you
// control" when both are chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c3c3bec-00f0-4de2-b929-73f1c9276ec4",
		Name:         "Sakashima's Will",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Target opponent chooses a creature they control. You gain control of it.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				sakashimasWillSteal),
			effects.ModeDoing("Choose a creature you control. Each other creature you control becomes a copy of that creature until end of turn.", nil,
				sakashimasWillCopy),
		).OrUpToIf(2, effects.YouControlACommander),
	})
}

// sakashimasWillCreaturesOf is the creatures `who` controls on the
// battlefield.
func sakashimasWillCreaturesOf(g *game.Game, who uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == who && c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

// sakashimasWillSteal is the first bullet: the opponent chooses.
func sakashimasWillSteal(_ *game.StackItem, ctx *effects.Context, occ int) error {
	refs := ctx.ModeTargets(occ)
	if len(refs) == 0 {
		return nil
	}
	opp := refs[0].ID
	cands := sakashimasWillCreaturesOf(ctx.Game, opp)
	switch len(cands) {
	case 0:
		return nil
	case 1:
		return sakashimasWillGain(ctx, cands[0])
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  opp,
		Source:   ctx.Source(),
		Question: "Sakashima's Will — choose a creature you control; its controller gains control of it",
		Cards:    cands,
		Min:      1,
		Max:      1,
		Zone:     game.ZoneBattlefield,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  sakashimasWillStealPick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

func sakashimasWillGain(ctx *effects.Context, id uuid.UUID) error {
	return effects.GainControl{
		Target:     id,
		Controller: ctx.Controller(),
		Duration:   game.IndefiniteDuration(),
		Label:      "Sakashima's Will — you gain control of the chosen creature",
	}.Apply(ctx)
}

// sakashimasWillStealPick is the opponent's answer. An on-disk key: never
// renamed, never reused.
var sakashimasWillStealPick = effects.AskedPick("choose-cards/sakashimas-will-steal", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return sakashimasWillGain(ctx, r.Picked[0])
})

// sakashimasWillCopy is the second bullet: you choose the model.
func sakashimasWillCopy(_ *game.StackItem, ctx *effects.Context, _ int) error {
	cands := sakashimasWillCreaturesOf(ctx.Game, ctx.Controller())
	switch len(cands) {
	case 0:
		return nil
	case 1:
		return sakashimasWillBecome(ctx, cands[0])
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Sakashima's Will — choose a creature you control; each other creature you control becomes a copy of it",
		Cards:    cands,
		Min:      1,
		Max:      1,
		Zone:     game.ZoneBattlefield,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  sakashimasWillCopyPick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

func sakashimasWillBecome(ctx *effects.Context, model uuid.UUID) error {
	var others []uuid.UUID
	for _, id := range sakashimasWillCreaturesOf(ctx.Game, ctx.Controller()) {
		if id != model {
			others = append(others, id)
		}
	}
	return effects.BecomeCopy{
		Targets: others,
		Of:      model,
		Label:   "Sakashima's Will — each other creature you control becomes a copy until end of turn",
	}.Apply(ctx)
}

// sakashimasWillCopyPick is your answer. An on-disk key: never renamed,
// never reused.
var sakashimasWillCopyPick = effects.AskedPick("choose-cards/sakashimas-will-copy", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return sakashimasWillBecome(ctx, r.Picked[0])
})
