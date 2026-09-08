package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// modes.go — S20 sub-PR 4: constructors for Spec.Modes so a modal
// card file reads like its oracle text:
//
//	Modes: ChooseOne(
//		Mode("Exile target player's graveyard.", TargetPlayer("target player")),
//		Mode("Destroy target artifact.", TargetPermanent("target artifact", Artifact())),
//		Mode("Each creature deals 1 damage to its controller."),
//	),
//
// and its OnResolve is a run of `if ctx.HasMode(i)` blocks in
// printed order. The engine validates the choice at announce and
// routes the chosen option's TargetSpec through the S20 legality
// gate; the client shows the picker between the X prompt and
// targeting.

// Mode declares one option. Pass a TargetSpec when the bullet
// targets; omit it otherwise.
func Mode(label string, targets ...*game.TargetSpec) game.ModeOption {
	o := game.ModeOption{Label: label}
	if len(targets) > 0 {
		o.Targets = targets[0]
	}
	return o
}

// ChooseOne — "Choose one —".
func ChooseOne(options ...game.ModeOption) *game.ModeSpec {
	return &game.ModeSpec{Prompt: "Choose one", Options: options, Min: 1, Max: 1}
}

// ChooseN — "Choose two —" (n, n) or "choose one or both" (1, 2).
func ChooseN(prompt string, min, max int, options ...game.ModeOption) *game.ModeSpec {
	return &game.ModeSpec{Prompt: prompt, Options: options, Min: min, Max: max}
}

// ManaValueGE passes when the card's mana value is ≥ n (Austere
// Command's "mana value 4 or greater").
func ManaValueGE(n int) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		cost, err := game.ParseCost(c.ManaCost)
		if err != nil {
			return false
		}
		return cost.Generic+len(cost.Required) >= n
	}
}
