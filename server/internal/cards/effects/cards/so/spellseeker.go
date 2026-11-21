package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spellseeker — Creature — Human Wizard {2}{U}, 1/1 (EDHREC rank
// 1043):
//
//	"When this creature enters, you may search your library for an
//	 instant or sorcery card with mana value 2 or less, reveal it,
//	 put it into your hand, then shuffle."
//
// The cheap-spell tutor on a blinkable body. "You may search" is the
// search's Optional flag (CR 701.23b — the searcher may decline the
// card and the shuffle), so the prompt always opens and "fail to
// find" is always an answer; the mana value is the printed one,
// which is what a card in a library has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "47a785ed-8095-4685-8daa-02c4e2b0ffcd",
		Name:         "Spellseeker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Spellseeker — search for an instant or sorcery with mana value 2 or less", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.B09IsCheapInstantOrSorceryCard,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Spellseeker — an instant or sorcery card with mana value 2 or less, revealed, to hand",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
