package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "696e7ddb-bdc7-40ee-bc5c-59e98f4a7401",
		Name:         "Restless Vents",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B|R}",
			Label:    "Add {B} or {R}",
		}},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(lootLabel, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, lootLabel, drawOneIfDiscarded)(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}{R}: Until end of turn, this land becomes a 2/3 black and red Insect creature with menace. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{B}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				return ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Insect")...),
						game.SetColorsMod(ctx.TextColors("B", "R")...),
						game.AddKeywordsMod("menace"),
					}, game.SetBasePTMods(2, 3)...),
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Restless Vents — a 2/3 black and red Insect creature with menace until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
