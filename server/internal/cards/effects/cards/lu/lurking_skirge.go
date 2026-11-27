package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lurking Skirge — Enchantment {1}{B}:
//
//	"When a creature is put into an opponent's graveyard from the
//	 battlefield, if this permanent is an enchantment, it becomes a 3/2
//	 Phyrexian Imp creature with flying."
//
// "An opponent's graveyard" is its owner's (CR 404.1): the creature may have
// been controlled by anyone. Whether it was a creature is read as it last
// existed (CR 603.10a). The intervening "if" (CR 603.4) is rechecked as the
// ability resolves; the creature it becomes is not an enchantment, so it
// triggers once.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21e90420-aedb-4a84-ba57-d5eae497b09b",
		Name:         "Lurking Skirge",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && source.HasCardType("enchantment") && dead.Owner != source.Controller
			}, "Lurking Skirge — it becomes a 3/2 Phyrexian Imp creature with flying",
				effects.OldB717BecomesACreature([]string{"Phyrexian", "Imp"}, 3, 2, "Lurking Skirge — a 3/2 Phyrexian Imp creature with flying",
					game.AddKeywordsMod("flying"))),
		},
	})
}
