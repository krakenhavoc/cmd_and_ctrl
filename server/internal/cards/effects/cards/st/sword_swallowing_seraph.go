package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sword-Swallowing Seraph — Creature — Angel Performer {4}{W}{W}, 4/4:
//
//	"Flying, vigilance
//	 When this creature enters, you may put a name sticker on a nonland
//	 permanent you own.
//	 {1}{W}, {T}: Put a +1/+1 counter on another target creature with a
//	 name sticker on it. Activate only as a sorcery."
//
// The name sticker is the engine's keyed questions with only name stickers
// offered (ADR 0208 §4). The target clause reads the sticker data
// (WithSticker).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "01efde9a-a20d-4972-b17f-c2d9eb024b0b",
		Name:            "Sword-Swallowing Seraph",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sword-Swallowing Seraph — you may put a name sticker on a nonland permanent you own",
				effects.Do(effects.PutASticker{Kinds: []game.StickerKind{game.StickerName}})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{W}, {T}: Put a +1/+1 counter on another target creature with a name sticker on it. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			SorcerySpeed: true,
			Targets:      effects.Another(effects.TargetCreature("another target creature with a name sticker on it", effects.WithSticker(game.StickerName))),
			// A +1/+1 counter on the announced creature, if still legal:
			// Animal Sanctuary's body.
			Effect: effects.B36CounterOnChosenAnimal,
		}},
	})
}
