package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mindshrieker — {1}{U} Creature — Spirit Bird 1/1 (Innistrad, S82 OLD b446,
// #5572):
//
//	"Flying
//	 {2}: Target player mills a card. This creature gets +X/+X until end of
//	 turn, where X is the milled card's mana value."
//
// The mill is the keyed mill, so the boost waits for a replacement's prompt
// and reads only the card that actually reached the graveyard (CR 400.7); with
// an empty library there is no milled card and X is 0. The card's mana value is
// read as it was in the library (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f638ceaf-6b44-4d99-95f3-0aa97b1738fe",
		Name:            "Mindshrieker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Target player mills a card. This creature gets +X/+X until end of turn, where X is the milled card's mana value.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return effects.MillToZone{Player: t.ID, N: 1, ThenTail: ctx.ItemTail(mindshriekerMilled)}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}

// mindshriekerMilled is the boost, by the milled card's mana value. An
// on-disk key: never renamed, never reused.
var mindshriekerMilled = effects.CardTail("mindshrieker/boost-by-milled-mana-value", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || !effects.B15OnBattlefield(ctx.Game, ctx.Item.SourceCardID) || ctx.IsNewSourceObject(ctx.Item.SourceCardID) {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	x := c.ManaValue()
	return effects.BoostUntilEOT{Target: ctx.Item.SourceCardID, Power: x, Toughness: x, Label: "Mindshrieker — +X/+X"}.Apply(ctx)
})
