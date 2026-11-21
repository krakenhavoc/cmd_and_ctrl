package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moonblade Shinobi — Creature — Human Ninja {3}{U}, 3/2:
//
//	"Ninjutsu {2}{U} ({2}{U}, Return an unblocked attacker you control
//	 to hand: Put this card onto the battlefield from your hand tapped
//	 and attacking.)
//	 Whenever this creature deals combat damage to a player, create a
//	 1/1 blue Illusion creature token with flying."
//
// #1227's third ninjutsu proof, and the one whose payoff is a body
// rather than a card: the Illusion arrives in the combat damage step,
// after blockers, so it is a flier for NEXT combat — and a flier is
// exactly what the next ninjutsu activation wants to swing with.
//
// Both halves are shared machinery. The keyword is Ninjutsu("{2}{U}")
// (ninjutsu.go); the trigger is the self-only combat damage shape; the
// token is the table entry Meloku already mints.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "046e25ee-c96d-4cff-93be-f7e3379d713c",
		Name:         "Moonblade Shinobi",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Ninjutsu("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(
				"Moonblade Shinobi — create a 1/1 blue Illusion with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Illusion with flying"), N: 1}),
			),
		},
	})
}
