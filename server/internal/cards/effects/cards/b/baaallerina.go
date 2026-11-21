package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baaallerina — Creature — Sheep Performer {3}{U}, 3/3:
//
//	"Flying
//	 When this creature enters, you may put a name sticker on a nonland
//	 permanent you own.
//	 {2}{U}: Another target creature with a name sticker on it gains
//	 flying until end of turn."
//
// The name sticker is the engine's keyed questions with only name stickers
// offered (ADR 0208 §4); its word goes where its controller says (CR
// 123.6b). The target clause reads the sticker data (WithSticker).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3bf3e503-0628-4550-84a9-7c1dd0a97457",
		Name:            "Baaallerina",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Baaallerina — you may put a name sticker on a nonland permanent you own",
				effects.Do(effects.PutASticker{Kinds: []game.StickerKind{game.StickerName}})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Another target creature with a name sticker on it gains flying until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    effects.ManaCost("{2}{U}"),
			Targets: effects.Another(effects.TargetCreature("another target creature with a name sticker on it", effects.WithSticker(game.StickerName))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Baaallerina — flying"}.Apply(ctx)
			},
		}},
	})
}
