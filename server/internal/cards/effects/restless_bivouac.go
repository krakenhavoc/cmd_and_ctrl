package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b3c7b46f-c9ab-40ca-b50b-a4e0d0bd9be8",
		Name:         "Restless Bivouac",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R|W}",
			Label:    "Add {R} or {W}",
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Restless Bivouac — put a +1/+1 counter on target creature you control",
				PlusOneCounterOnChosenTarget),
				TargetCreature("target creature you control", YouControl())),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}{W}: This land becomes a 2/2 red and white Ox creature until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{1}{R}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				return ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Ox")...),
						game.SetColorsMod(ctx.TextColors("R", "W")...),
					}, game.SetBasePTMods(2, 2)...),
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Restless Bivouac — a 2/2 red and white Ox creature until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
