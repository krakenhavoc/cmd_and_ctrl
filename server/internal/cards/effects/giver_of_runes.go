package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "4ec85850-f274-4c0c-9a03-0488267caa14",
		Name:         "Giver of Runes",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Another target creature you control gains protection from colorless or from the color of your choice until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect, Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "protection"))},
			Cost:    game.AbilityCost{Tap: true},
			Targets: Another(TargetCreature("another target creature you control", YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					then := ctx.ItemTail(giverOfRunesThen)
					then.Carry = []uuid.UUID{t.ID}
					ChooseColorOrColorlessTail(game.ColorForProtection, g, item.Controller, ctx.Source(),
						"Giver of Runes — choose colorless or a color", then)
					return nil
				}
				return nil
			},
		}},
	})
}

var giverOfRunesThen = CardTail("then/giver-of-runes-protection", func(ctx *Context, o game.TailOutcome) error {
	token := game.ProtectionFromColorOrColorless(o.Tail.Word)
	if token == "" || len(o.Tail.Carry) == 0 {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   o.Tail.Carry[0],
		Keywords: []string{token},
		Label:    "Giver of Runes — " + token,
	}.Apply(ctx)
})
