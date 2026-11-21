package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Restless Vents — Land (The Lost Caverns of Ixalan, #4470):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {R}.
//	 {1}{B}{R}: Until end of turn, this land becomes a 2/3 black and red
//	 Insect creature with menace. It's still a land.
//	 Whenever this land attacks, you may discard a card. If you do, draw a
//	 card."
//
// Restless Reef's shape: the animation is one effect at one timestamp
// (creature type and Insect subtype in layer 4, black and red in layer 5,
// menace in layer 6, base 2/3 in layer 7b), pinned to this object (CR
// 611.2c). Menace is part of the animation's text, not a printed keyword
// of the land, so it ends with the effect. The attack trigger is the
// land's own printed ability; its rummage is Rubble Rouser's, the discard
// first and a draw only for a card actually discarded.
//
// No simplification.
func init() {
	const lootLabel = "Restless Vents — you may discard a card; if you do, draw a card"
	effects.Register(effects.Spec{
		OracleID:     "696e7ddb-bdc7-40ee-bc5c-59e98f4a7401",
		Name:         "Restless Vents",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B|R}",
			Label:    "Add {B} or {R}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks(lootLabel, func(g *game.Game, item *game.StackItem) error {
				return effects.B39MayDiscardThenDraw(1, false, lootLabel, effects.DrawOneIfDiscarded)(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}{R}: Until end of turn, this land becomes a 2/3 black and red Insect creature with menace. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}{B}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Insect")...),
						game.SetColorsMod(ctx.TextColors("B", "R")...),
						game.AddKeywordsMod("menace"),
					}, game.SetBasePTMods(2, 3)...),
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Restless Vents — a 2/3 black and red Insect creature with menace until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
