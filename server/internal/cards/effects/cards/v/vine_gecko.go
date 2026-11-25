package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vine Gecko — Creature — Elemental Lizard {1}{G}, 2/2 (Zendikar Rising,
// S82 OLD #5508):
//
//	"The first kicked spell you cast each turn costs {1} less to cast.
//	 Whenever you cast a kicked spell, put a +1/+1 counter on this
//	 creature."
//
// The discount reads the turn's cast tally of kicked spells (CastTally.
// Kicked), which is bumped after a spell is priced (CR 601.2f), so a first
// kicked spell sees none before it. The kicker is read from the announcement
// being priced, so the discount applies to the spell that is kicked and to
// no other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9596b05-e0bb-4e70-82fe-5db60f87bf3d",
		Name:         "Vine Gecko",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "The first kicked spell you cast each turn costs {1} less to cast.",
				effects.YourSpell(), effects.S82OldB092FirstKickedSpellThisTurn()),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.S82OldB089YouCastAKickedSpell,
				"Vine Gecko — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
