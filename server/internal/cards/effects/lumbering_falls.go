package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lumbering Falls — Land:
//
//	"This land enters tapped.
//	 {T}: Add {G} or {U}.
//	 {2}{G}{U}: Until end of turn, this land becomes a 3/3 green and blue
//	 Elemental creature with hexproof. It's still a land."
//
// Creeping Tar Pit's manland: one continuous effect pinned to this object
// (CR 611.2c): the creature type and Elemental subtype (layer 4), green and
// blue (layer 5), hexproof (layer 6) and the base 3/3 (layer 7b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c2878f4e-3044-4d60-8ec1-b325edfca397",
		Name:         "Lumbering Falls",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G|U}",
			Label:    "Add {G} or {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}{U}: Until end of turn, this land becomes a 3/3 green and blue Elemental creature with hexproof. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{2}{G}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				return ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
						game.SetColorsMod(ctx.TextColors("G", "U")...),
						game.AddKeywordsMod("hexproof"),
					}, game.SetBasePTMods(3, 3)...),
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Lumbering Falls — a 3/3 green and blue Elemental creature with hexproof until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
