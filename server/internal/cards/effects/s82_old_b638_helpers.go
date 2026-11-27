package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB638BlocksOrIsBlockedByBlueOrBlack is Amphibious Kavu's trigger condition:
// "Whenever this creature blocks or becomes blocked by one or more blue and/or
// black creatures". It is one trigger per block (CR 509.3a, 509.3c): the
// Kavu's first EventBlock when at least one attacker it blocks is blue or
// black, or its one EventBecomesBlocked when at least one of the creatures
// blocking it is.
func OldB638BlocksOrIsBlockedByBlueOrBlack(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	blueOrBlack := func(c *game.Card) bool {
		return c.IsCreature() && (c.HasColorWord("U") || c.HasColorWord("B"))
	}
	switch ev.Kind {
	case game.EventBlock:
		if ev.CardID != source.InstanceID || ev.Amount > 1 {
			return false
		}
		for _, a := range source.BlockedAttackers() {
			if c, ok := g.LookupCardForEffect(a); ok && blueOrBlack(&c) {
				return true
			}
		}
	case game.EventBecomesBlocked:
		if ev.CardID != source.InstanceID {
			return false
		}
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.IsBlockingAttacker(source.InstanceID) && blueOrBlack(&c) {
				return true
			}
		}
	}
	return false
}

// OldB638BecomesChosenColorUntilEOT is "{G}: This creature becomes the color of
// your choice until end of turn" (Caldera Kavu), Spiritmonger's row under its
// own continuation key: the colour is asked as the ability resolves, then set
// with the colour-setting mod, so the creature is that colour and no other
// (CR 105.3) until the cleanup step (CR 514.2).
func OldB638BecomesChosenColorUntilEOT(name string, chosen game.TailRef) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		ChooseColorTail(game.ColorForBenefit, g, item.Controller, item.SourceCardID,
			name+" — choose a color", NewContext(g, item).ItemTail(chosen))
		return nil
	}
}

// OldB638ColorChosenTail registers the continuation OldB638BecomesChosenColorUntilEOT
// hands the answer to. The key is an on-disk identity: never renamed.
func OldB638ColorChosenTail(key, name string) game.TailRef {
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		color := o.Tail.Word
		if color == "" || !SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
			return nil
		}
		return ScopedEffectFor{
			Target:   ctx.Item.SourceCardID,
			Mods:     []game.Mod{game.SetColorsMod(color)},
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    name + " — " + game.ColorName(color),
		}.Apply(ctx)
	})
}
