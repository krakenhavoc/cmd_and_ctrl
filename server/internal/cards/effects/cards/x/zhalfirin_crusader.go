package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zhalfirin Crusader — Creature — Human Knight {1}{W}{W}, 2/2:
//
//	"Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)
//	 {1}{W}: The next 1 damage that would be dealt to this creature this
//	 turn is dealt to any target instead."
//
// ADR 0108 §9 (#1905): flanking and a 1-point charged redirection (CR
// 615.7) from the Crusader to the target. Flanking is the keyword token,
// which the engine triggers from (game/flanking.go, #5784); the row
// beside it is the retired pre-#5784 trigger, kept for restore
// (RetiredFlankingRow).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9e60c102-412e-4956-a7bb-a2cd737d6692",
		Name:            "Zhalfirin Crusader",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{effects.RetiredFlankingRow("Zhalfirin Crusader")},
		Activated: []effects.ActivatedAbility{
			effects.RedirectRow("{1}{W}: The next 1 damage that would be dealt to this creature this turn is dealt to any target instead.",
				effects.ManaCost("{1}{W}"), effects.TargetAny(), effects.RedirectDamage{Protect: effects.ShieldThis, Amount: 1, To: effects.RedirectToClause(0)}),
		},
	})
}
