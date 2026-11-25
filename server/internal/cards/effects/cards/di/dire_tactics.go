package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dire Tactics — Instant {W}{B} (Ikoria: Lair of Behemoths, #5512):
//
//	"Exile target creature. If you don't control a Human, you lose life
//	 equal to that creature's toughness."
//
// The toughness is read BEFORE the exile, as the layered view gave it on the
// battlefield (CR 608.2h, Swords to Plowshares' shape); the Human check is
// read after the exile, where the clause sits in the text, so a Human target
// of your own that was just exiled no longer counts. A target gone in
// response does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c8cb2ec-5b58-497c-a6e7-2d25512d5a44",
		Name:         "Dire Tactics",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !ctx.IsTargetLegal(t) {
				return nil
			}
			toughness := 0
			if info, ok := ctx.Game.PermanentForEffect(t.Object()); ok {
				toughness = info.Toughness
			}
			return effects.ExileTarget{Target: t.ID, ThenTail: ctx.TailN(direTacticsLoseLife, toughness)}.Apply(ctx)
		},
	})
}

// direTacticsLoseLife is the second sentence, run once the exile has settled
// (whether or not the creature reached exile, as the text has no "if you do"):
// the Human check is read now, the toughness is the one carried from the
// battlefield. An on-disk key.
var direTacticsLoseLife = effects.CardTail("then/dire-tactics-lose-life", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Tail.N <= 0 {
		return nil
	}
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() && c.HasSubtype(ctx.TextSubtype("Human")) {
			return nil
		}
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -o.Tail.N)
})
