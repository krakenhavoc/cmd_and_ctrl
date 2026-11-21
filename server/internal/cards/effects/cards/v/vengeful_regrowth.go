package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vengeful Regrowth:
//
//	"Return up to three target land cards from your graveyard to the
//	 battlefield tapped. Create that many 4/2 green Plant Warrior creature
//	 tokens with reach.
//	 Flashback {6}{G}{G}"
//
// Each target still legal as it resolves returns (CR 608.2b); the tokens
// are counted from the lands that actually entered, so a land a
// replacement effect kept out makes no token. Flashback is the shared
// alternative cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "c4d05f64-d1bb-4725-81ba-baf737f3066c",
		Name:             "Vengeful Regrowth",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerValue},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{6}{G}{G}")},
		Targets:          effects.TargetCardInGraveyard("up to three target land cards from your graveyard", effects.YouOwn(), effects.Land()).WithCount(0, 3),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			returned := 0
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Tapped: true}).Apply(ctx); err != nil {
					return err
				}
				if z := ctx.Game.FindCardZoneForEffect(t.ID); z != nil && z.Kind == game.ZoneBattlefield {
					returned++
				}
			}
			if returned == 0 {
				return nil
			}
			return effects.CreateToken{Template: effects.TokenCard("4/2 green Plant Warrior with reach"), N: returned}.Apply(ctx)
		},
	})
}
