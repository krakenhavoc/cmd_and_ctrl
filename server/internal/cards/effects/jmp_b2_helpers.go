package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jmp_b2_helpers.go — shared bodies for the Jumpstart Dimir deck request,
// second batch (#4892).

// jmpB2DestroyTargetCreatureItsControllerLosesTwo is "Destroy target
// creature. Its controller loses 2 life." — Bitter Downfall's body, shared
// with Launch Party. The creature's controller is read before the destroy,
// so the life loss still lands when the creature is indestructible or
// regenerated, and the spell does nothing when its target is gone
// (CR 608.2b).
func jmpB2DestroyTargetCreatureItsControllerLosesTwo(_ *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			continue
		}
		controller := c.Controller
		if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), controller, -2)
	}
	return nil
}

// jmpB2CastFromGraveyardBySacrificing is "You may cast this card from your
// graveyard by paying <mana> and sacrificing <n> <label> rather than paying
// its mana cost" (Scourge of Nel Toth): SacrificeInstead's sacrifice, a
// printed mana price, and bound to the graveyard, so a cast from hand may
// not claim it and a cast from the graveyard must.
func jmpB2CastFromGraveyardBySacrificing(mana string, n int, label string, preds ...CardPredicate) game.AlternativeCost {
	ac := SacrificeInstead(n, label, 0, preds...)
	ac.Key = "graveyard_sacrifice"
	ac.Label = "Cast this card from your graveyard by paying " + mana + " and sacrificing " + label
	ac.ManaCost = mana
	ac.FromZone = game.ZoneGraveyard
	return ac
}
