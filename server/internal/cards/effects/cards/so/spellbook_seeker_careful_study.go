package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spellbook Seeker // Careful Study — Creature — Bird Wizard {3}{U}, 3/3 //
// Sorcery {U} (preparation card, CR 722; Secrets of Strixhaven, #4456):
//
//	"Flying
//	 This creature enters prepared. (While it's prepared, you may cast a copy
//	 of its spell. Doing so unprepares it.)"
//
//	Careful Study — "Draw two cards, then discard two cards."
//
// Flying is a printed keyword and the permanent enters prepared
// (SelfEntersPrepared, Honorbound Page). The prepare spell is the loot
// every "draw N, then discard N" card shares (lootOne): the discard opens
// only once both draws have resolved, so a drawn card may be discarded.
//
// No simplification.
func init() {
	const id = "df5e4100-8c83-42d6-af96-1a1ec84a9210"
	effects.Register(effects.Spec{
		OracleID:        id,
		Name:            "Spellbook Seeker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersPrepared()},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Careful Study",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Discards: 2},
		OnResolve:    socB35CarefulStudyResolve,
	})
}

// socB35CarefulStudyResolve is Careful Study: draw two cards, then discard
// two cards.
func socB35CarefulStudyResolve(item *game.StackItem, ctx *effects.Context) error {
	return effects.LootOne(ctx.Game, item, 2)
}
