package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Terashi's Grasp — Sorcery — Arcane {2}{W} (Betrayers of Kamigawa,
// #5602):
//
//	"Destroy target artifact or enchantment. You gain life equal to its
//	 mana value."
//
// The mana value is read off the target before it is destroyed, so the
// life gain is the permanent's mana value as it last existed on the
// battlefield (CR 608.2h). It is gained even if the permanent is not
// destroyed (an indestructible one), and not at all if the target has
// become illegal, when the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d4738552-3a5e-43c5-a975-8b77618bacaf",
		Name:         "Terashi's Grasp",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			mv, _ := ctx.Game.PermanentManaValueForEffect(id)
			if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Amount: mv}.Apply(ctx)
		},
	})
}
