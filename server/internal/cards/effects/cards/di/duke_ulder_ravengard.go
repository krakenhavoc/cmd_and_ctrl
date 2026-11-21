package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Duke Ulder Ravengard — Legendary Creature — Human Noble Soldier
// {4}{R}{W}, 5/5 (Commander Legends: Battle for Baldur's Gate, #5948):
//
//	"At the beginning of combat on your turn, another target creature you
//	 control gains haste and myriad until end of turn. (Whenever it attacks,
//	 for each opponent other than defending player, you may create a token
//	 copy that's tapped and attacking that player or a planeswalker they
//	 control. Exile the tokens at end of combat.)"
//
// Haste is GrantKeywordUntilEOT; myriad is the granted trigger of Blade of
// Selves' bundle (GrantedMyriad), given to the creature until end of turn the
// way Banishing Knack gives its activated ability (GrantAbilitiesFor). The
// target is "another" creature, never Ulder himself, and is checked again as
// the ability resolves.
//
// No simplification.
func init() {
	const grant = "duke-ulder-ravengard/myriad"
	const label = "Duke Ulder Ravengard — another target creature you control gains haste and myriad until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "792b9752-2d46-48ca-abdd-408a787612da",
		Name:         "Duke Ulder Ravengard",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       grant,
			Triggered: []game.TriggeredAbility{effects.GrantedMyriad("Duke Ulder Ravengard")},
			Text:      "Myriad",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true), label,
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"}, Label: "Duke Ulder Ravengard — haste"}).Apply(ctx); err != nil {
								return err
							}
							return effects.GrantAbilitiesFor{
								Target: t.ID,
								Keys:   []string{grant},
								Label:  "Duke Ulder Ravengard — myriad until end of turn",
							}.Apply(ctx)
						}
						return nil
					}),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
