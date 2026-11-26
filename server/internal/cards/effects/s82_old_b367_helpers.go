package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b367_helpers.go — the "you may sacrifice this Aura. If you do,
// destroy target …" shape of Flamespeaker's Will and Mortal Obstinacy (OLD
// batch 367, #5558, S82).

// S82OldB367DestroyTargetTail registers the "if you do" half: destroy the
// trigger's target, if it is still legal as the answer lands (CR 608.2b), and
// only when the Aura really went to the graveyard.
func S82OldB367DestroyTargetTail(key string) game.TailRef {
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		if !OneWentThisWay(o) {
			return nil
		}
		return DestroyEachLegalTarget(nil, ctx)
	})
}

// S82OldB367SacrificeAuraAnswer registers the "you may sacrifice this Aura"
// answer: on a yes the source is sacrificed and the tail told whether it left.
func S82OldB367SacrificeAuraAnswer(key string, then game.TailRef) game.ConfirmThen {
	return MayYes(key, func(ctx *Context, _ game.ConfirmAnswer) error {
		if !OnBattlefield(ctx.Game, ctx.Source()) {
			return nil
		}
		return SacrificePermanent{Target: ctx.Source(), ThenTail: ctx.ItemTail(then)}.Apply(ctx)
	})
}
