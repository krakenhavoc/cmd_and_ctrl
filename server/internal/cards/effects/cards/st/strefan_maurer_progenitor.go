package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:        "3cc745d2-5021-43ff-ace8-35011011e62a",
		Name:            "Strefan, Maurer Progenitor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Strefan, Maurer Progenitor — create a Blood token for each player who lost life this turn",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, p := range g.Seats {
						if p != nil && !p.Eliminated && effects.B18LifeLostThisTurn(g, p.ID) > 0 {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.BloodToken(), N: n}.Apply(effects.NewContext(g, item))
				}),
			effects.Optional(effects.WheneverThisAttacks("Strefan, Maurer Progenitor — sacrifice two Blood tokens; if you do, you may put a Vampire card from your hand onto the battlefield tapped and attacking",
				strefanAttacks),
				"Strefan, Maurer Progenitor — sacrifice two Blood tokens?"),
		},
	})
}

// strefanPutter is the put: keys "choose-cards/strefan-maurer-progenitor-put"
// and "option-pick/strefan-maurer-progenitor-attacks".
var strefanPutter = effects.NewHandAttackerPutter("strefan-maurer-progenitor", "Strefan, Maurer Progenitor", strefanIndestructible)

// strefanIndestructible is "It gains indestructible until end of turn", on
// the card that entered. An on-disk key.
var strefanIndestructible = effects.EnteredTail("strefan-maurer-progenitor/indestructible", func(ctx *effects.Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil // nothing entered
	}
	return effects.GrantKeywordUntilEOT{
		Target:   entered,
		Keywords: []string{"indestructible"},
		Label:    "Strefan, Maurer Progenitor — indestructible until end of turn",
	}.Apply(ctx)
})

// strefanSacrificed is "If you do": two Blood tokens went.
var strefanSacrificed = effects.RegisterSacrificedOne("prompt-run/strefan-maurer-progenitor-put", func(ctx *effects.Context, _ uuid.UUID) error {
	strefanPutter.Offer(ctx.Game, ctx.Item, effects.HandCardsMatching(ctx.Game, ctx.Controller(), effects.HasSubtype("Vampire")),
		"Strefan, Maurer Progenitor — you may put a Vampire card from your hand onto the battlefield tapped and attacking")
	return nil
})

// strefanAttacks asks for exactly two of the Blood tokens the controller
// has now; fewer than two is no question and no "if you do".
func strefanAttacks(g *game.Game, item *game.StackItem) error {
	return effects.ChoosePermanents{
		Question: "Strefan, Maurer Progenitor — sacrifice two Blood tokens",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			out := effects.PermanentsControlledByMatching(g, of, effects.HasSubtype("Blood"))
			if len(out) < 2 {
				return nil, 0, 0
			}
			return out, 2, 2
		},
		Keyed: strefanSacrificed.Pick,
	}.Apply(effects.NewContext(g, item))
}
