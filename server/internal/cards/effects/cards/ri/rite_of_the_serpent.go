package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rite of the Serpent — Sorcery {4}{B}{B} (Khans of Tarkir):
//
//	"Destroy target creature. If that creature had a +1/+1 counter on it,
//	 create a 1/1 green Snake creature token."
//
// The counters are read before the destroy, from the creature as it stands as
// the spell resolves (CR 608.2h's last-known information would say the same):
// a creature with a +1/+1 counter that survives (indestructible, regenerated)
// still pays out the Snake, as the ruling says, and a creature that left in
// response is an illegal target so the whole spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f009ef70-d718-45cf-bb22-e560484bbfe5",
		Name:         "Rite of the Serpent",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				had := false
				if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok {
					had = c.Counters[game.CounterPlusOne] > 0
				}
				if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if had {
					return effects.CreateToken{Template: effects.TokenCard("1/1 green Snake"), N: 1}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
