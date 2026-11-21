package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invasion of Vryn // Overloaded Mage-Ring — March of the Machine, a
// transforming double-faced card:
//
//	Invasion of Vryn — Battle — Siege {3}{U}, defense 4
//	  "(As a Siege enters, choose an opponent to protect it. You and
//	    others can attack it. When it's defeated, exile it, then cast it
//	    transformed.)
//	   When this Siege enters, draw three cards, then discard a card."
//	Overloaded Mage-Ring — Artifact
//	  "{1}, {T}, Sacrifice this artifact: Copy target spell you control.
//	   You may choose new targets for the copy. (A copy of a permanent
//	   spell becomes a token.)"
//
// The Siege under the bare oracle ID, the Mage-Ring under "<oracle>#1".
// The protector, the defense counters and the defeat are the engine's,
// keyed on the card type (game/battle.go). The draw is Meat Locker's: the
// discard prompt opens after the three draws, so a drawn card can be the
// one discarded. The Mage-Ring is Slick Imitator's copy body, with the
// sacrifice and tap paid at announce.
//
// No simplification.
const invasionOfVrynOracleID = "73c7fdb3-4742-4d08-8658-daa4435c49b1"

func init() {
	effects.Register(effects.Spec{
		OracleID:     invasionOfVrynOracleID,
		Name:         "Invasion of Vryn",
		Completeness: effects.CompletenessFull,
		Battle: &effects.BattleSpec{
			Defense: 4,
			Subtype: effects.BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Invasion of Vryn — draw three cards, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return effects.B16DrawThenDiscard(g, item, 3) }),
			effects.DefeatedTrigger("Invasion of Vryn — defeated: exile it, then cast Overloaded Mage-Ring", effects.SiegeDefeated()),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     invasionOfVrynOracleID + "#1",
		Name:         "Overloaded Mage-Ring",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: Copy target spell you control. You may choose new targets for the copy.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetSpell("target spell you control", effects.YouControl()),
			Effect:  effects.CopyTargetedSpell,
		}},
	})
}
