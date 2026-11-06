package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Treetop Village — Land (Avatar: The Last Airbender Eternal, #5075):
//
//	"This land enters tapped.
//	 {T}: Add {G}.
//	 {1}{G}: This land becomes a 3/3 green Ape creature with trample
//	 until end of turn. It's still a land."
//
// Creeping Tar Pit's manland shape: one continuous effect pinned to this
// object, the creature type and Ape (layer 4), green (layer 5), trample
// (layer 6) and the base 3/3 (layer 7b), so a +1/+1 counter or a pump
// still applies on top (the rulings). Summoning sickness is the
// engine's (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b53f216d-1592-4eee-b204-502a805fbc8c",
		Name:         "Treetop Village",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: This land becomes a 3/3 green Ape creature with trample until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{1}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				return ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Ape")...),
						game.SetColorsMod(ctx.TextColors("G")...),
						game.AddKeywordsMod("trample"),
					}, game.SetBasePTMods(3, 3)...),
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Treetop Village — a 3/3 green Ape creature with trample until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
