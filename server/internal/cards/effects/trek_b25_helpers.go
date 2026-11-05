package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b25_helpers.go — shared bodies for Star Trek batch 25 (#4973).

// trekKlingonWarriorSacrificedAtNextEndStep is "create a 2/1 red Klingon
// Warrior creature token with haste. Sacrifice it at the beginning of the
// next end step." as a trigger's whole Effect — D-7 Battle Cruiser's 3+
// line and L'Rell, Fierce Visionary's combat trigger. The token is created
// through the keyed creation door, and `then` (a card's own registered
// tail, built with trekSacrificeTokensAtNextEndStep) is handed exactly the
// tokens that landed, so a replacement that makes more (or none) is
// followed.
func trekKlingonWarriorSacrificedAtNextEndStep(then game.TailRef) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		return g.CreateTokensKeyedForEffect(game.TokenCreation{
			Controller: item.Controller,
			Source:     item.SourceCardID,
			Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("2/1 red Klingon Warrior with haste")), Count: 1}},
		}, ctx.ItemTail(then))
	}
}

// trekSacrificeTokensAtNextEndStep is the tail body that schedules a CR
// 603.7 delayed trigger at the beginning of the next end step sacrificing
// the tokens the creation made (the tail's IDs), labelled `label`.
func trekSacrificeTokensAtNextEndStep(label string) func(ctx *Context, o game.TailOutcome) error {
	return func(ctx *Context, o game.TailOutcome) error {
		if len(o.IDs) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{
			Label: label,
			Cards: o.IDs,
			Body:  sacrificeListedCardsBody,
		}.Apply(ctx)
	}
}

// trekB25ArtifactCreaturesYouControl is "artifact creatures you control"
// for a static's AppliesTo: the source itself included, read through the
// layered types.
func trekB25ArtifactCreaturesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.IsCreature() && target.IsArtifact()
}

// trekB25ExceptItHasMenace is the "except it has menace" exception clause
// of a token copy (CR 707.9), Mirror Universe Intrusion's: menace rides
// the token's own keywords, as TokenCopyGainsHaste's haste does.
func trekB25ExceptItHasMenace(t *game.Card) {
	for _, kw := range t.Keywords {
		if kw == "menace" {
			return
		}
	}
	t.Keywords = append(t.Keywords, "menace")
}
