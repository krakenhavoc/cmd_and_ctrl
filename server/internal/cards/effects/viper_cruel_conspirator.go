package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Viper, Cruel Conspirator — Legendary Creature — Human Assassin Villain
// {1}{B}, 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"{B}: Target creature that's attacking alone gets +1/+1 until end of
//	 turn.
//	 {B}: Target creature that's attacking alone gains your choice of
//	 deathtouch or lifelink until end of turn."
//
// "Attacking alone" is CR 506.5 read live (AttackingAlone), on any
// player's creature. The first ability is Crowd of True Believers'
// target. The second chooses as it resolves (Gideon Blackblade's
// PickOption shape; the target rides the prompt as data), and does
// nothing if its target is gone (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6824a6ed-37d0-4db5-b2d4-9c737d0c49f9",
		Name:         "Viper, Cruel Conspirator",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{B}: Target creature that's attacking alone gets +1/+1 until end of turn.",
				Cost:    ManaCost("{B}"),
				Purpose: ForTargets(BenefitUntilEOT(0, 1, 1)),
				Targets: TargetCreature("target creature that's attacking alone", AttackingAlone()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Viper, Cruel Conspirator — +1/+1"}.Apply(ctx)
				},
			},
			{
				Label:   "{B}: Target creature that's attacking alone gains your choice of deathtouch or lifelink until end of turn.",
				Cost:    ManaCost("{B}"),
				Purpose: ForTargets(BenefitUntilEOT(0, 0, 0)),
				Targets: TargetCreature("target creature that's attacking alone", AttackingAlone()),
				Effect:  viperChoosesAKeyword,
			},
		},
	})
}

// viperKeywords are the two keywords the second ability offers, in printed
// order.
var viperKeywords = []string{"deathtouch", "lifelink"}

// viperChoosesAKeyword asks for the keyword as the ability resolves.
func viperChoosesAKeyword(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	options := make([]game.ChoiceOption, len(viperKeywords))
	for i, kw := range viperKeywords {
		options[i] = game.ChoiceOption{Label: "Gains " + kw + " until end of turn"}
	}
	return PickOption{
		Question: "Viper, Cruel Conspirator — the creature gains your choice of deathtouch or lifelink until end of turn",
		Options:  options,
		ThenKey:  viperKeywordChosen,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
}

// viperKeywordChosen grants the chosen keyword to the carried target, keyed
// (#4253).
var viperKeywordChosen = OptionPickThen("option-pick/viper-cruel-conspirator-keyword", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(viperKeywords) || len(r.Carry) == 0 {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   r.Carry[0],
		Keywords: []string{viperKeywords[r.Index]},
		Label:    "Viper, Cruel Conspirator — gains " + viperKeywords[r.Index] + " until end of turn",
	}.Apply(ctx)
})
