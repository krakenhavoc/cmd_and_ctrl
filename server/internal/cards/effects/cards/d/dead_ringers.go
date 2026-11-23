package cardsd

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dead Ringers — Sorcery {4}{B} (Mystery Booster 2, #5948):
//
//	"Destroy two target nonblack creatures unless either one is a color the
//	 other isn't. They can't be regenerated."
//
// Two distinct nonblack creatures. As the spell resolves the two are compared
// by colour: they are destroyed only when they have exactly the same colours
// (two colourless creatures match, a colourless creature and a green one do
// not). A target that has left is read as it last existed (CR 608.2h); a
// target that became illegal is not destroyed, and the other still is when
// the colours match (CR 608.2b). The destruction ignores regeneration
// (CR 701.19c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb34e487-f7a9-48ad-8f42-00295a80a160",
		Name:         "Dead Ringers",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("two target nonblack creatures", effects.NonBlack()).WithCount(2, 2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) != 2 {
				return nil
			}
			var colours [2][]string
			for i, t := range item.Targets {
				if c, ok := effects.DisorientingChoiceOnBattlefield(ctx, t.ID); ok {
					colours[i] = slices.Clone(c.EffectiveColors())
				} else if info, ok := ctx.Game.LastKnownPermanentForEffect(t.ID); ok {
					colours[i] = slices.Clone(info.Characteristic.Colors)
				} else {
					return nil
				}
				slices.Sort(colours[i])
			}
			if !slices.Equal(colours[0], colours[1]) {
				return nil
			}
			for _, t := range ctx.LegalTargets() {
				if err := (effects.DestroyTarget{Target: t.ID, CantBeRegenerated: true}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
