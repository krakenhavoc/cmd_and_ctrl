package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soaring Stoneglider — Creature — Elephant Cleric {2}{W}, 4/3:
//
//	"As an additional cost to cast this spell, exile two cards from your
//	 graveyard or pay {1}{W}.
//	 Flying, vigilance"
//
// An either/or additional cost (ADR 0100 §2) whose first branch is the
// graveyard-exile cast cost (#1948, ADR 0189's 2026-10-16 amendment):
// the caster announces which branch with the cast (cost_branch, CR
// 601.2b), and the chosen branch is the plan's mandatory entry, so the
// two cards are named on exile_ids and exiled with the spell on the
// stack exactly as Abhorrent Oculus's six are, while the {1}{W} branch is
// priced with the mana cost (CR 601.2f). A graveyard of fewer than two
// cards leaves only the mana branch (CR 118.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "62d18d2b-fa68-49cd-b62d-803728ddfa8b",
		Name:            "Soaring Stoneglider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		AdditionalCost: effects.EitherCost(
			effects.ExileFromGraveyardCost(2, "two cards", nil).Keyed("exile"),
			effects.ManaAdditionalCost("{1}{W}").Keyed("mana"),
		),
	})
}
