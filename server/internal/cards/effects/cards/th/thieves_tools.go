package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thieves' Tools — Artifact — Equipment, {1}{B}:
//
//	"When this Equipment enters, create a Treasure token.
//	 Equipped creature can't be blocked as long as its power is 3 or
//	 less.
//	 Equip {2}"
//
// The evasion clause is a CONDITIONAL "can't be blocked" (#750, ADR
// 0045 addendum Decision 11): CantBeBlockedWhile on OnAttached, with
// the condition read off the equipped creature's live power when
// blockers are declared (CR 509.1b). It is a block rule rather than a
// Restriction bit behind a layer-6 condition because the condition is
// the creature's POWER, which layer 7 finishes computing after layer 6
// has run: a layer-6 static would read the power from before this
// turn's pumps and anthems. A block rule reads it after every layer,
// at the moment blocks are declared. Pumping the creature above 3
// after blocks are declared changes nothing, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7fe361ef-a168-4847-92a2-21c1661aac06",
		Name:         "Thieves' Tools",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Thieves' Tools — create a Treasure",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedWhile(effects.OnAttached(), effects.PowerLE(3)),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}
