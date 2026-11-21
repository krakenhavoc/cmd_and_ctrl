package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Restless Bivouac — Land (Wilds of Eldraine, #4471):
//
//	"This land enters tapped.
//	 {T}: Add {R} or {W}.
//	 {1}{R}{W}: This land becomes a 2/2 red and white Ox creature until end
//	 of turn. It's still a land.
//	 Whenever this land attacks, put a +1/+1 counter on target creature you
//	 control."
//
// Restless Ridgeline's shape: the animation is one effect at one
// timestamp (creature type and Ox subtype in layer 4, red and white in
// layer 5, base 2/2 in layer 7b), pinned to this object (CR 611.2c). The
// attack trigger is the land's own printed ability. Its target is any
// creature you control, the Bivouac included, chosen as the trigger goes
// on the stack (CR 603.3d); a target gone by resolution gets nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3c7b46f-c9ab-40ca-b50b-a4e0d0bd9be8",
		Name:         "Restless Bivouac",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R|W}",
			Label:    "Add {R} or {W}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Restless Bivouac — put a +1/+1 counter on target creature you control",
				effects.PlusOneCounterOnChosenTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}{W}: This land becomes a 2/2 red and white Ox creature until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{1}{R}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Ox")...),
						game.SetColorsMod(ctx.TextColors("R", "W")...),
					}, game.SetBasePTMods(2, 2)...),
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Restless Bivouac — a 2/2 red and white Ox creature until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
