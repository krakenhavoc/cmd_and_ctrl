package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b540_helpers.go — Shadowmoor batch 540 (#5591).

// OldB540CounterThenExtract is "Counter target spell. Search its controller's
// graveyard, hand, and library for all cards with the same name as that spell
// and exile them. Then that player shuffles." (Counterbore): the spell's name
// and controller are read before it is countered, while it is still on the
// stack (a split card cast as one half has only that half's name, CR 709.3b),
// then the counter, then the extraction (ExtractionSearch, #4033). A spell
// that can't be countered stays on the stack and the search still happens; a
// target that has left the stack makes the spell do nothing (CR 608.2b).
func OldB540CounterThenExtract(card string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		target, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		spell := ctx.Game.StackItemForEffect(target)
		c, found := ctx.Game.LookupCardForEffect(target)
		if spell == nil || !found {
			return nil
		}
		victim, names := spell.Controller, ExtractionNames(c)
		if err := (CounterTarget{StackID: target}).Apply(ctx); err != nil {
			return err
		}
		return ExtractionSearch(ctx, card, victim, names, true)
	}
}

// OldB540ColorCreaturesYouControlHave is "<Color> creatures you control have
// <keyword>" as a layer 6 static (Bloodmark Mentor's first strike, Roughshod
// Mentor's trample). The colour is read from each creature's current colours
// (CR 105.1), the way the colour anthems read it, so a creature recoloured
// in layer 5 counts.
func OldB540ColorCreaturesYouControlHave(color, keyword string) game.StaticAbility {
	grant := KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller && target.HasColor(source.TextColor(color))
	}, keyword)
	name := game.ColorName(color)
	return Labelled(grant, strings.ToUpper(name[:1])+name[1:]+" creatures you control have "+keyword+".")
}
