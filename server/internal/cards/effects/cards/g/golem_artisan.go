package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Golem Artisan — Artifact Creature — Golem {5}, 3/3 (Jumpstart 2022):
//
//	"{2}: Target artifact creature gets +1/+1 until end of turn.
//	 {2}: Target artifact creature gains your choice of flying, trample, or
//	 haste until end of turn."
//
// Two targeted abilities on artifact creatures (the Artisan included). The
// keyword is chosen as the second one resolves (CR 608.2d), with the target
// riding the prompt as data; a target that has left the battlefield or
// stopped being an artifact creature is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8537fc2c-c027-4440-bff4-d046fc198428",
		Name:         "Golem Artisan",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}: Target artifact creature gets +1/+1 until end of turn.",
				Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)), game.AnswerPump),
				Cost:    effects.ManaCost("{2}"),
				Targets: effects.TargetCreature("target artifact creature", effects.Artifact()),
				Effect:  effects.BoostTargetUntilEOT("Golem Artisan — +1/+1", 1, 1),
			},
			{
				Label:   "{2}: Target artifact creature gains your choice of flying, trample, or haste until end of turn.",
				Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)), game.AnswerCombatGrant),
				Cost:    effects.ManaCost("{2}"),
				Targets: effects.TargetCreature("target artifact creature", effects.Artifact()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.PickOption{
						Question: "Golem Artisan — the target gains which until end of turn?",
						Options:  []game.ChoiceOption{{Label: "Flying"}, {Label: "Trample"}, {Label: "Haste"}},
						ThenKey:  golemArtisanKeyword,
						Carry:    []uuid.UUID{id},
					}.Apply(ctx)
				},
			},
		},
	})
}

// golemArtisanKeyword is the keyword chosen, keyed. An on-disk key: never
// renamed, never reused.
var golemArtisanKeyword = effects.OptionPickThen("option-pick/golem-artisan-keyword", func(ctx *effects.Context, r game.OptionPicked) error {
	keywords := []string{"flying", "trample", "haste"}
	if r.Index < 0 || r.Index >= len(keywords) || len(r.Carry) == 0 {
		return nil
	}
	id := r.Carry[0]
	if id == uuid.Nil || !effects.OnBattlefield(ctx.Game, id) {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{keywords[r.Index]}, Label: "Golem Artisan — " + keywords[r.Index]}.Apply(ctx)
})
