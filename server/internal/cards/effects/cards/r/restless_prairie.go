package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Restless Prairie — Land (The Lost Caverns of Ixalan, #4470):
//
//	"This land enters tapped.
//	 {T}: Add {G} or {W}.
//	 {2}{G}{W}: This land becomes a 3/3 green and white Llama creature
//	 until end of turn. It's still a land.
//	 Whenever this land attacks, other creatures you control get +1/+1
//	 until end of turn."
//
// Restless Reef's shape: the animation is one effect at one timestamp
// (creature type and Llama subtype in layer 4, green and white in layer
// 5, base 3/3 in layer 7b), pinned to this object (CR 611.2c). The attack
// trigger is the land's own printed ability, which can only fire while it
// is a creature. "Other creatures you control" is the set as the trigger
// resolves (CR 611.2c): the Prairie is never in it, and a creature that
// arrives afterwards gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c071257a-63e7-48d0-a677-0b396a09b624",
		Name:         "Restless Prairie",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G|W}",
			Label:    "Add {G} or {W}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Restless Prairie — other creatures you control get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID)), Power: 1, Toughness: 1,
						Label: "Restless Prairie — other creatures you control get +1/+1",
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}{W}: This land becomes a 3/3 green and white Llama creature until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{2}{G}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Llama")...),
						game.SetColorsMod(ctx.TextColors("G", "W")...),
					}, game.SetBasePTMods(3, 3)...),
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Restless Prairie — a 3/3 green and white Llama creature until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
