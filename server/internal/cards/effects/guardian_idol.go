package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guardian Idol — Artifact {2} (Foundations Jumpstart):
//
//	"This artifact enters tapped.
//	 {T}: Add {C}.
//	 {2}: This artifact becomes a 2/2 Golem artifact creature until end of
//	 turn."
//
// Mishra's Factory's animation (CR 611.2a, layers 4 and 7b) on a rock that
// enters tapped (Abraded Bluffs' SelfEntersTapped).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "829d2e11-192f-4093-99f8-e13767b321a2",
		Name:         "Guardian Idol",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}: This artifact becomes a 2/2 Golem artifact creature until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				return ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Golem")...),
					}, game.SetBasePTMods(2, 2)...),
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Guardian Idol — a 2/2 Golem artifact creature until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
