package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Windscouter — Creature — Human Scout {3}{U}, 3/3 (Prophecy, S82 OLD batch 673,
// #5617):
//
//	"Flying
//	 When this creature attacks or blocks, return it to its owner's hand at end
//	 of combat. (Return it only if it's on the battlefield.)"
//
// Brackwater Elemental's attack-or-block trigger (one printed ability, one
// registered trigger; it triggers once however many attackers it blocks,
// CR 509.3a) over Cinder Wall's delayed trigger at the beginning of the end of
// combat step (CR 511.2), returning the permanent to its owner's hand instead
// of destroying it. The return happens only if it is still the same object on
// the battlefield (CR 400.7).
//
// No simplification.
func init() {
	const label = "Windscouter — return it to its owner's hand at end of combat"
	effects.Register(effects.Spec{
		OracleID:        "58232c49-f880-41ef-b808-62a6e804c715",
		Name:            "Windscouter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)
			}, label, effects.S82OldB673ReturnThisAtEndOfCombat(label)),
		},
	})
}
