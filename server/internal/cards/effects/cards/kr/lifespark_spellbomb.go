package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lifespark Spellbomb — Artifact {1} (Mirrodin, S82 OLD, #5606):
//
//	"{G}, Sacrifice this artifact: Until end of turn, target land becomes a
//	 3/3 creature that's still a land.
//	 {1}, Sacrifice this artifact: Draw a card."
//
// The first ability is one continuous effect pinned to the target land
// (CR 611.2a, CR 611.2c): it adds the creature type (layer 4) and sets the
// base power and toughness to 3/3 (layer 7b), so the land keeps its land
// type, its subtypes and its abilities, and a land that has been under its
// controller's control since the turn began can attack (CR 302.6). The
// sacrifice is a cost, so the Spellbomb is gone whatever happens to the
// target, and a land that has left does nothing (CR 608.2b). The second is
// the Spellbomb cycle's {1} cantrip.
//
// No simplification.
func init() {
	const label = "Lifespark Spellbomb — target land becomes a 3/3 creature until end of turn; it's still a land"
	effects.Register(effects.Spec{
		OracleID:     "3f351696-1402-4b69-b924-f0ba14320287",
		Name:         "Lifespark Spellbomb",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{G}, Sacrifice this artifact: Until end of turn, target land becomes a 3/3 creature that's still a land.",
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Cost:    effects.Plus(effects.ManaCost("{G}"), effects.SacrificeThis()),
				Targets: effects.TargetPermanent("target land", effects.Land()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.ScopedEffectFor{
						Target:   id,
						Mods:     append([]game.Mod{game.AddTypesMod("Creature")}, game.SetBasePTMods(3, 3)...),
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    label,
					}.Apply(ctx)
				},
			},
			{
				Label:   "{1}, Sacrifice this artifact: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
		},
	})
}
