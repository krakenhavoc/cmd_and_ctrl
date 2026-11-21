package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armory Guard — Creature — Giant Soldier {3}{W}, 2/5 (Ravnica Remastered,
// #5948):
//
//	"This creature has vigilance as long as you control a Gate."
//
// Grabby Tabby's self-only keyword grant (layer 6, CR 613.1f) while the
// controller controls a Gate: the permanents are read live, through the
// layers, so a Gate that is turned into something else or leaves ends it.
//
// No simplification.
func init() {
	vigilance := effects.KeywordGrant(effects.SelfWhileYouControlA(effects.HasSubtype("Gate")), "vigilance")
	vigilance.Label = "This creature has vigilance as long as you control a Gate."
	effects.Register(effects.Spec{
		OracleID:     "96c67950-df59-430c-94fd-234086c6c8f2",
		Name:         "Armory Guard",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{vigilance},
	})
}
