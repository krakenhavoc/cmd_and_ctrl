package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stalwarts of Osgiliath — Creature — Human Soldier {4}{W}, 4/3:
//
//	"When this creature enters, the Ring tempts you.
//	 Whenever you draw your second card each turn, put a +1/+1 counter
//	 on this creature."
//
// The second ability is Thopter Fabricator's trigger: it fires on the
// second card drawn in each turn, anyone's turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d41b2a4f-e137-4e61-99eb-6261ab3edd03",
		Name:         "Stalwarts of Osgiliath",
		Purpose:      game.Purpose{RingTempts: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Stalwarts of Osgiliath — the Ring tempts you", effects.Do(effects.TheRingTemptsYou{})),
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Stalwarts of Osgiliath — put a +1/+1 counter on this creature", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
