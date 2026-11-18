package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Strefan, Maurer Progenitor — Legendary Creature — Vampire Noble
// {2}{B}{R}, 3/2 (Innistrad: Crimson Vow, #5495):
//
//	"Flying
//	 At the beginning of your end step, create a Blood token for each
//	 player who lost life this turn.
//	 Whenever Strefan attacks, you may sacrifice two Blood tokens. If you
//	 do, you may put a Vampire card from your hand onto the battlefield
//	 tapped and attacking. It gains indestructible until end of turn."
//
// The end-step trigger counts the players (you included) who lost life
// this turn, as it resolves; with none, nothing is made. The attack
// trigger asks "you may sacrifice two Blood tokens" first (Optional), then,
// as it resolves, which two (a controller with fewer than two is not
// asked, so "if you do" does not happen). The put is Nick Fury's
// HandAttackerPutter: the Vampire enters tapped and attacking a defender
// its controller chooses, was never declared as an attacker, and gains
// indestructible until end of turn only if it entered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3cc745d2-5021-43ff-ace8-35011011e62a",
		Name:            "Strefan, Maurer Progenitor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Strefan, Maurer Progenitor — create a Blood token for each player who lost life this turn",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, p := range g.Seats {
						if p != nil && !p.Eliminated && b18LifeLostThisTurn(g, p.ID) > 0 {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return CreateToken{Template: BloodToken(), N: n}.Apply(NewContext(g, item))
				}),
			Optional(WheneverThisAttacks("Strefan, Maurer Progenitor — sacrifice two Blood tokens; if you do, you may put a Vampire card from your hand onto the battlefield tapped and attacking",
				strefanAttacks),
				"Strefan, Maurer Progenitor — sacrifice two Blood tokens?"),
		},
	})
}

// strefanPutter is the put: keys "choose-cards/strefan-maurer-progenitor-put"
// and "option-pick/strefan-maurer-progenitor-attacks".
var strefanPutter = NewHandAttackerPutter("strefan-maurer-progenitor", "Strefan, Maurer Progenitor", strefanIndestructible)

// strefanIndestructible is "It gains indestructible until end of turn", on
// the card that entered. An on-disk key.
var strefanIndestructible = EnteredTail("strefan-maurer-progenitor/indestructible", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil // nothing entered
	}
	return GrantKeywordUntilEOT{
		Target:   entered,
		Keywords: []string{"indestructible"},
		Label:    "Strefan, Maurer Progenitor — indestructible until end of turn",
	}.Apply(ctx)
})

// strefanSacrificed is "If you do": two Blood tokens went.
var strefanSacrificed = RegisterSacrificedOne("prompt-run/strefan-maurer-progenitor-put", func(ctx *Context, _ uuid.UUID) error {
	strefanPutter.Offer(ctx.Game, ctx.Item, handCardsMatching(ctx.Game, ctx.Controller(), HasSubtype("Vampire")),
		"Strefan, Maurer Progenitor — you may put a Vampire card from your hand onto the battlefield tapped and attacking")
	return nil
})

// strefanAttacks asks for exactly two of the Blood tokens the controller
// has now; fewer than two is no question and no "if you do".
func strefanAttacks(g *game.Game, item *game.StackItem) error {
	return ChoosePermanents{
		Question: "Strefan, Maurer Progenitor — sacrifice two Blood tokens",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			out := permanentsControlledByMatching(g, of, HasSubtype("Blood"))
			if len(out) < 2 {
				return nil, 0, 0
			}
			return out, 2, 2
		},
		Keyed: strefanSacrificed.pick,
	}.Apply(NewContext(g, item))
}
