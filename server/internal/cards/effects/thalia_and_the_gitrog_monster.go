package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thalia and The Gitrog Monster — Legendary Creature — Human Frog Horror
// {1}{W}{B}{G}, 4/4 (March of the Machine Commander):
//
//	"First strike, deathtouch
//	 You may play an additional land on each of your turns.
//	 Creatures and nonbasic lands your opponents control enter tapped.
//	 Whenever Thalia and The Gitrog Monster attacks, sacrifice a creature or
//	 land, then draw a card."
//
// The extra land drop is The Gitrog Monster's AdditionalLandPlays; the
// enter-tapped replacement is Thalia, Heretic Cathar's. The attack trigger is a
// mandatory sacrifice (it may be Thalia and The Gitrog Monster itself); the
// draw is the sacrifice's tail, so it happens even with nothing to sacrifice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "422a76e1-97a2-4a90-b07d-f487a2d8c236",
		Name:                "Thalia and The Gitrog Monster",
		Completeness:        CompletenessFull,
		PrintedKeywords:     []string{"first strike", "deathtouch"},
		AdditionalLandPlays: 1,
		Purpose:             game.Purpose{ExtraLandDrops: 1},
		Replacements: []game.ReplacementEffect{
			opponentsCreaturesAndNonbasicLandsEnterTapped("Thalia and The Gitrog Monster: enters tapped"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Thalia and The Gitrog Monster — sacrifice a creature or land, then draw a card",
				thaliaGitrogAttack),
		},
	})
}

func thaliaGitrogAttack(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return SacrificeChoice{
		Player:     item.Controller,
		Candidates: permanentsControlledByMatching(g, item.Controller, Or(Creature(), Land())),
		Question:   "Thalia and The Gitrog Monster — sacrifice a creature or land",
		ThenTail:   ctx.ItemTail(thaliaGitrogSacrificed),
	}.Apply(ctx)
}

var thaliaGitrogSacrificed = CardTail("then/thalia-and-the-gitrog-monster-draw", func(ctx *Context, _ game.TailOutcome) error {
	return DrawCards{N: 1}.Apply(ctx)
})
