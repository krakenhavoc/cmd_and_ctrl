package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:            "422a76e1-97a2-4a90-b07d-f487a2d8c236",
		Name:                "Thalia and The Gitrog Monster",
		Completeness:        effects.CompletenessFull,
		PrintedKeywords:     []string{"first strike", "deathtouch"},
		AdditionalLandPlays: 1,
		Purpose:             game.Purpose{ExtraLandDrops: 1},
		Replacements: []game.ReplacementEffect{
			effects.OpponentsCreaturesAndNonbasicLandsEnterTapped("Thalia and The Gitrog Monster: enters tapped"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Thalia and The Gitrog Monster — sacrifice a creature or land, then draw a card",
				thaliaGitrogAttack),
		},
	})
}

func thaliaGitrogAttack(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return effects.SacrificeChoice{
		Player:     item.Controller,
		Candidates: effects.PermanentsControlledByMatching(g, item.Controller, effects.Or(effects.Creature(), effects.Land())),
		Question:   "Thalia and The Gitrog Monster — sacrifice a creature or land",
		ThenTail:   ctx.ItemTail(thaliaGitrogSacrificed),
	}.Apply(ctx)
}

var thaliaGitrogSacrificed = effects.CardTail("then/thalia-and-the-gitrog-monster-draw", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.DrawCards{N: 1}.Apply(ctx)
})
