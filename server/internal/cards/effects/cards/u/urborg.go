package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urborg — Legendary Land (Tales of Middle-earth Commander):
//
//	"{T}: Add {B}.
//	 {T}: Target creature loses first strike or swampwalk until end of turn."
//
// The second ability is a choice as it resolves (Ballroom Brawlers' shape): the
// controller picks first strike or swampwalk for the target, which then loses
// that one in layer 6 with its own timestamp (CR 613.1f), so a grant with a
// later timestamp gives it back. It does nothing to a target that is gone
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6114962-035e-4e7f-9009-4739bf83a05a",
		Name:         "Urborg",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature loses first strike or swampwalk until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.PickOption{
					Question: "Urborg — the target loses which until end of turn?",
					Options:  []game.ChoiceOption{{Label: "First strike"}, {Label: "Swampwalk"}},
					ThenKey:  s82GapB120UrborgLoses,
					Carry:    []uuid.UUID{id},
				}.Apply(ctx)
			},
		}},
	})
}

var s82GapB120UrborgLoses = effects.OptionPickThen("option-pick/urborg-loses", func(ctx *effects.Context, r game.OptionPicked) error {
	keywords := []string{"first strike", "swampwalk"}
	if r.Index < 0 || r.Index >= len(keywords) || len(r.Carry) == 0 || !effects.OnBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	return effects.UntilEndOfTurn(ctx, r.Carry[0], nil, "Urborg — loses "+keywords[r.Index]+" until end of turn",
		game.RemoveKeywordsMod(keywords[r.Index]))
})
