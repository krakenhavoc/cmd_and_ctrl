package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b510_helpers.go — Shards of Alara, "Activated abilities" (S82 OLD
// batch 510, #5588). Append-only.

// S82OldB510ThisGainsRow is "<cost>: This creature gains <keyword> until end
// of turn." (Cloudheath Drake, Shore Snapper, Vectis Silencers).
func S82OldB510ThisGainsRow(cost, keyword string) ActivatedAbility {
	label := cost + ": This creature gains " + keyword + " until end of turn."
	return ActivatedAbility{
		Label:   label,
		Cost:    ManaCost(cost),
		Purpose: game.Purpose{Answers: game.AnswerSurvive},
		Effect:  ThisGainsKeywordUntilEndOfTurn(keyword, label),
	}
}

// S82OldB510TargetGainsRow is "<cost>: Target creature <with ...> gains
// <keyword> until end of turn." The cost may carry {T}. Clause is the target
// label without "target creature", e.g. "with power 5 or greater".
func S82OldB510TargetGainsRow(cost game.AbilityCost, costText, clause, keyword string, preds ...CardPredicate) ActivatedAbility {
	target := "target creature"
	if clause != "" {
		target += " " + clause
	}
	label := costText + ": Target creature"
	if clause != "" {
		label += " " + clause
	}
	label += " gains " + keyword + " until end of turn."
	return ActivatedAbility{
		Label:   label,
		Cost:    cost,
		Targets: TargetCreature(target, preds...),
		Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, keyword)),
		Effect:  EbTargetCreatureGainsUntilEOT(keyword, label),
	}
}

// S82OldB510RegenerateRow is "<cost>, <extra cost>: Regenerate this creature."
// for a row with a non-mana cost part (Skeletal Kathari, Viashino Skeleton).
func S82OldB510RegenerateRow(label string, cost game.AbilityCost) ActivatedAbility {
	return ActivatedAbility{
		Label:   label,
		Purpose: game.Purpose{Answers: game.AnswerSurvive},
		Cost:    cost,
		Effect:  RegenerateThisSource,
	}
}
