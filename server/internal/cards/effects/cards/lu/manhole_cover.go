package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Manhole Cover — Artifact {2} (Teenage Mutant Ninja Turtles Commander,
// #5011):
//
//	"Flash
//	 When this artifact enters, target creature gains indestructible until
//	 end of turn.
//	 {2}, Sacrifice this artifact: Target player draws a card."
//
// Flash rides PrintedKeywords, so the cast from hand is open at instant
// speed. The entry trigger is Plumecreed Escort's targeted grant with any
// creature and indestructible (a canonical keyword the destruction path
// reads), pinned to the creature as the trigger resolves (CR 611.2c); a
// target gone by then gets nothing (CR 608.2b). The activation is Oona's
// Grace's body behind a {2} and sacrifice cost (CR 602.2b), with no timing
// restriction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "38f7ff90-4cdf-4403-ad10-1400e05c32e9",
		Name:            "Manhole Cover",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Manhole Cover — target creature gains indestructible until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							return effects.GrantKeywordUntilEOT{
								Target:   t.ID,
								Keywords: []string{"indestructible"},
								Label:    "Manhole Cover — indestructible until end of turn",
							}.Apply(ctx)
						}
						return nil
					}),
				effects.TargetCreature("target creature")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice this artifact: Target player draws a card.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target player"),
			Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 1}),
			Effect:  effects.TheFirstLegalTargetPlayerDrawsACard,
		}},
	})
}
