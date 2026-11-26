package cardsdi

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dimir Charm — Instant {U}{B} (Gatecrash, #5565):
//
//	"Choose one —
//	 • Counter target sorcery spell.
//	 • Destroy target creature with power 2 or less.
//	 • Look at the top three cards of target player's library. Put one back
//	   and the rest into that player's graveyard."
//
// Mode 2 is Gurmag Nightwatch's private look at the top three (a shorter
// library offers what it has) with the looker, the caster, naming the library
// they look at: one pick is required and stays on top, and every other card
// still in the library goes to its owner's graveyard (a put, not a mill). A
// card that left the library while the prompt was open is a new object and
// stays where it went (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d99e786-585d-4026-9253-19741fb357de",
		Name:         "Dimir Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Counter target sorcery spell.", effects.TargetSpell("target sorcery spell", effects.Sorcery())),
			effects.ModeWithPurpose(effects.Mode("Destroy target creature with power 2 or less.",
				effects.TargetCreature("target creature with power 2 or less", effects.PowerLE(2))), game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeWithPurpose(effects.Mode("Look at the top three cards of target player's library. Put one back and the rest into that player's graveyard.",
				effects.TargetPlayer("target player")), game.Purpose{Answers: game.AnswerValue}),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			target := item.Targets[0]
			switch {
			case ctx.HasMode(0):
				if target.Kind != game.TargetCard || !dimirCharmLegal(ctx, target) {
					return nil
				}
				return effects.CounterTarget{StackID: target.ID}.Apply(ctx)
			case ctx.HasMode(1):
				if target.Kind != game.TargetCard || !dimirCharmLegal(ctx, target) {
					return nil
				}
				return effects.DestroyTarget{Target: target.ID}.Apply(ctx)
			case ctx.HasMode(2):
				if target.Kind != game.TargetPlayer || !dimirCharmLegal(ctx, target) {
					return nil
				}
				looked := ctx.Game.LookAtTopOfPlayersLibraryForEffect(ctx.Controller(), target.ID, 3)
				if len(looked) == 0 {
					return nil
				}
				ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:    ctx.Controller(),
					FromPlayer: target.ID,
					Source:     ctx.Source(),
					Question:   "Dimir Charm — put one of those cards back on top of the library; the rest go into its owner's graveyard",
					Cards:      looked,
					Min:        1,
					Max:        1,
					LiveZone:   game.ZoneLibrary,
					ThenKey:    dimirCharmPick,
					Asked:      game.NewAskedItem(item),
				})
			}
			return nil
		},
	})
}

// dimirCharmLegal reports whether the chosen mode's target is still legal as
// the spell resolves (CR 608.2b).
func dimirCharmLegal(ctx *effects.Context, want game.TargetRef) bool {
	return slices.ContainsFunc(ctx.LegalTargets(), func(t game.TargetRef) bool { return t.Kind == want.Kind && t.ID == want.ID })
}

// dimirCharmPick puts every looked-at card that is still in the library and
// was not chosen into its owner's graveyard. An on-disk key: never renamed,
// never reused.
var dimirCharmPick = effects.AskedPick("choose-cards/s82-old-b415-dimir-charm-rest", func(ctx *effects.Context, r game.CardsChosen) error {
	for _, id := range r.Candidates {
		if slices.Contains(r.Picked, id) || !effects.StdB9InLibrary(ctx.Game, id) {
			continue
		}
		if err := ctx.Game.PutIntoGraveyardForEffect(id); err != nil {
			return err
		}
	}
	return nil
})
