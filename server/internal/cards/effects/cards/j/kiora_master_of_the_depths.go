package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kiora, Master of the Depths — Legendary Planeswalker — Kiora {2}{G}{U}
// (Battle for Zendikar, #5547), loyalty 4:
//
//	"+1: Untap up to one target creature and up to one target land.
//	 −2: Reveal the top four cards of your library. You may put a creature
//	 card and/or a land card from among them into your hand. Put the rest
//	 into your graveyard.
//	 −8: You get an emblem with "Whenever a creature you control enters, you
//	 may have it fight target creature." Then create three 8/8 blue Octopus
//	 creature tokens."
//
// The emblem trigger is one trigger per entering creature, with its target
// chosen as it goes on the stack and the fight optional on resolution; the
// emblem is made before the tokens, so each Octopus entering triggers it too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0b829189-ff9d-4f6c-9a2a-fdb4a9bfeeb6",
		Name:            "Kiora, Master of the Depths",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &effects.EmblemSpec{
			Label: "Kiora, Master of the Depths emblem",
			Text:  "Whenever a creature you control enters, you may have it fight target creature.",
			Triggered: []game.TriggeredAbility{
				effects.Optional(effects.Targeting(
					effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl,
						"Kiora, Master of the Depths emblem — have it fight target creature", effects.Cmdr2S37TurnedUpCreatureFights),
					effects.TargetCreature("target creature")),
					"Kiora, Master of the Depths emblem — have the creature fight target creature?"),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Untap up to one target creature and up to one target land.",
				Cost:  effects.LoyaltyCost(1),
				Targets: effects.Clauses(
					effects.TargetCreature("up to one target creature").WithCount(0, 1),
					effects.TargetPermanent("up to one target land", effects.Land()).WithCount(0, 1),
				),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for slot := 0; slot < 2; slot++ {
						t, ok := ctx.ClauseTarget(slot)
						if !ok || t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label:  "−2: Reveal the top four cards of your library. You may put a creature card and/or a land card from among them into your hand. Put the rest into your graveyard.",
				Cost:   effects.LoyaltyCost(-2),
				Effect: effects.S82OldB314KioraMinusTwo,
			},
			{
				Label: "−8: You get an emblem with \"Whenever a creature you control enters, you may have it fight target creature.\" Then create three 8/8 blue Octopus creature tokens.",
				Cost:  effects.LoyaltyCost(-8),
				Effect: effects.Do(effects.CreateEmblem{},
					effects.CreateToken{Template: effects.TokenCard("8/8 blue Octopus"), N: 3}),
			},
		},
	})
}
