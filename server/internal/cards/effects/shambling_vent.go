package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shambling Vent — Land:
//
//	"This land enters tapped.
//	 {T}: Add {W} or {B}.
//	 {1}{W}{B}: Until end of turn, this land becomes a 2/3 white and black
//	 Elemental creature with lifelink. It's still a land."
//
// Restless Vents' manland shape: one continuous animation at one timestamp
// (creature type and Elemental subtype in layer 4, colours in layer 5, the
// keyword in layer 6, base 2/3 in layer 7b), pinned to this object until
// end of turn (CR 611.2c). The keyword is part of the animation, not a
// printed keyword of the land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4725abd4-06bc-464a-bc5b-c9e0f71ec079",
		Name:         "Shambling Vent",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|B}",
			Label:    "Add {W} or {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}{B}: Until end of turn, this land becomes a 2/3 white and black Elemental creature with lifelink. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{W}{B}"),
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
						game.SetColorsMod(ctx.TextColors("W", "B")...),
						game.AddKeywordsMod("lifelink"),
					}, game.SetBasePTMods(2, 3)...),
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Shambling Vent — a 2/3 Elemental creature with lifelink until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
