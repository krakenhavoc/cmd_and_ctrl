package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grixis Slavedriver — Creature — Zombie Giant {5}{B}, 4/4 (Conflux,
// #5587):
//
//	"When this creature leaves the battlefield, create a 2/2 black
//	 Zombie creature token.
//	 Unearth {3}{B} ({3}{B}: Return this card from your graveyard to the
//	 battlefield. It gains haste. Exile it at the beginning of the next
//	 end step or if it would leave the battlefield. Unearth only as a
//	 sorcery.)"
//
// A leaves-the-battlefield trigger (CR 603.6c) on any exit, dying or
// not, so an unearthed Slavedriver that is exiled makes its Zombie too.
// The unearth ability is the shared one (unearth.go, CR 702.84a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9087928a-4732-4661-b3ed-d0b1ea06edf6",
		Name:         "Grixis Slavedriver",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisLeaves("Grixis Slavedriver — create a 2/2 black Zombie creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{3}{B}")},
	})
}
