package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fire Lord Ozai — Legendary Creature — Human Noble {3}{B}, 4/4 (Avatar:
// The Last Airbender Eternal, #5075):
//
//	"Whenever Fire Lord Ozai attacks, you may sacrifice another
//	 creature. If you do, add an amount of {R} equal to the sacrificed
//	 creature's power. Until end of combat, you don't lose this mana as
//	 steps end.
//	 {6}: Exile the top card of each opponent's library. Until end of
//	 turn, you may play one of those cards without paying its mana
//	 cost."
//
// The attack trigger asks at resolution, over the creatures you control
// then (maySacrificeOne), and the mana is the sacrificed creature's power
// as it last existed (CR 608.2h), anthems and counters included. It is
// firebending's mana: red, kept until end of combat.
//
// The activation exiles one card off each opponent's library in APNAP
// order and grants ONE play over all of them: one permission per library
// tied into one budget (game.CastPermission.BudgetGroup with CastsLeft
// 1), so casting or playing any of them spends the rest. The play is
// free ({0} in place of the mana cost; a land is simply played with the
// turn's land drop), follows the card's normal timing (the ruling) and
// ends with the turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5945306d-e792-4eae-999b-dfb44b10f062",
		Name:         "Fire Lord Ozai",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Fire Lord Ozai — you may sacrifice another creature to add {R} equal to its power",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Fire Lord Ozai — sacrifice another creature to add {R} equal to its power (it lasts until end of combat)?",
						Creature(), true, fireLordOzaiSacrificed)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{6}: Exile the top card of each opponent's library. Until end of turn, you may play one of those cards without paying its mana cost.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{6}"),
			Effect:  fireLordOzaiExile,
		}},
	})
}

// fireLordOzaiSacrificed is the "if you do": red mana equal to the
// sacrificed creature's last-known power, kept until end of combat.
var fireLordOzaiSacrificed = RegisterSacrificedOne("prompt-run/fire-lord-ozai-add-red", func(ctx *Context, sacrificed uuid.UUID) error {
	n := departedCreaturePower(ctx.Game, sacrificed)
	if n <= 0 {
		return nil
	}
	return redUntilEndOfCombat(n).Apply(ctx)
})

// fireLordOzaiExile is the activation's resolution.
func fireLordOzaiExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	perm := ExileTopWithPermission{Free: true}.permission()
	perm.CastsLeft = 1
	perm.BudgetGroup = uuid.New()
	perm.Source = item.SourceCardID
	perm.SourceName = "Fire Lord Ozai"
	perm.Label = "Fire Lord Ozai — play one of those cards without paying its mana cost"
	step := game.KeyedTail{Ref: exileTopEachStep, Source: ctx.Source(), N: 1, Permission: &perm, Player: item.Controller}
	return exileTopEachFrom(g, step, apnapOrder(g, ctx.Opponents()), nil)
}
