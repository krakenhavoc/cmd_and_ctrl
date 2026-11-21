package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// In the Pale Moonlight — Enchantment — Saga {3}{B}{B} (Star Trek,
// #4973 batch 12):
//
//	"I — Each opponent sacrifices a nontoken creature of their choice.
//	 II — Investigate.
//	 III — You may sacrifice an artifact or creature. If you do, create
//	       two 2/2 black Romulan creature tokens."
//
// Chapter I is Sheoldred's Edict's first mode: each opponent picks, an
// opponent with no nontoken creature sacrifices nothing.
//
// Chapter III asks as it resolves, over the artifacts and creatures you
// control then (a Clue from chapter II is a fine answer), and makes the
// Romulans only once something has really been sacrificed. With nothing
// to sacrifice nobody is asked and nothing is made. The Saga is
// sacrificed by the state-based check after chapter III has resolved
// (CR 714.4), not before.
//
// No simplification.
func init() {
	const name = "In the Pale Moonlight"
	effects.Register(effects.Spec{
		OracleID:     "90c78f79-982f-4acd-a8ba-0c6ce115a248",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "each opponent sacrifices a nontoken creature"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.EachPlayerSacrifices{ExceptController: true, Match: effects.B04NontokenCreature, Label: "a nontoken creature"}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "investigate"),
				effects.Do(effects.Investigate{N: 1})),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "you may sacrifice an artifact or creature to create two 2/2 black Romulans"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "In the Pale Moonlight — sacrifice an artifact or creature to create two 2/2 black Romulan creature tokens?",
						effects.Or(effects.Artifact(), effects.Creature()), false, inThePaleMoonlightRomulans)
				}),
		},
	})
}

// inThePaleMoonlightRomulans is the two Romulans, once a permanent has
// really been sacrificed. Keyed: an on-disk identity, never renamed or
// reused.
var inThePaleMoonlightRomulans = effects.RegisterSacrificedOne("prompt-run/in-the-pale-moonlight-romulans", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.CreateToken{Template: effects.TokenCard("2/2 black Romulan"), N: 2}.Apply(ctx)
})
