package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Giver of Runes — Creature — Kor Cleric, {W}, 1/2:
//
//	"{T}: Another target creature you control gains protection from
//	 colorless or from the color of your choice until end of turn."
//
// Mother of Runes with "another" and the colorless answer: the target is
// chosen at activation, the color at resolution (Angelic Intervention's
// ChooseColorOrColorlessTail), and the grant is the layer-6 keyword.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ec85850-f274-4c0c-9a03-0488267caa14",
		Name:         "Giver of Runes",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Another target creature you control gains protection from colorless or from the color of your choice until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection"))},
			Cost:    game.AbilityCost{Tap: true},
			Targets: effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					then := ctx.ItemTail(giverOfRunesThen)
					then.Carry = []uuid.UUID{t.ID}
					effects.ChooseColorOrColorlessTail(game.ColorForProtection, g, item.Controller, ctx.Source(),
						"Giver of Runes — choose colorless or a color", then)
					return nil
				}
				return nil
			},
		}},
	})
}

var giverOfRunesThen = effects.CardTail("then/giver-of-runes-protection", func(ctx *effects.Context, o game.TailOutcome) error {
	token := game.ProtectionFromColorOrColorless(o.Tail.Word)
	if token == "" || len(o.Tail.Carry) == 0 {
		return nil
	}
	return effects.GrantKeywordUntilEOT{
		Target:   o.Tail.Carry[0],
		Keywords: []string{token},
		Label:    "Giver of Runes — " + token,
	}.Apply(ctx)
})
