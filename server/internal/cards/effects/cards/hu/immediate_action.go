package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Immediate Action — Conspiracy (Conspiracy, #6102):
//
//	"Hidden agenda (Start the game with this conspiracy face down in the
//	 command zone and secretly choose a card name. You may turn this
//	 conspiracy face up any time and reveal that name.)
//	 Creatures you control with the chosen name have haste."
//
// Hidden agenda is the engine's (ADR 0227): the card starts face down in
// its owner's command zone, the name is chosen before the hands are dealt,
// and the reveal_agenda special action turns it face up. Face down it has
// no abilities (CR 315.5b); face up, a layer-6 haste grant to the
// creatures its controller controls with that name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a9c1c2d-d4d7-4df3-a853-f1b81c67b3e3",
		Name:         "Immediate Action",
		Completeness: effects.CompletenessFull,
		HiddenAgenda: effects.HiddenAgendaOneName,
		Static: []game.StaticAbility{func() game.StaticAbility {
			ab := effects.KeywordGrant(effects.CreaturesYouControlWithTheChosenName, "haste")
			ab.Label = "Creatures you control with the chosen name have haste."
			return ab
		}()},
	})
}
