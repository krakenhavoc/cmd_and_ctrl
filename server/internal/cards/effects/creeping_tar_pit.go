package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Creeping Tar Pit — Land:
//
//	"This land enters tapped.
//	 {T}: Add {U} or {B}.
//	 {1}{U}{B}: Until end of turn, this land becomes a 3/2 blue and black
//	 Elemental creature. It's still a land. It can't be blocked this
//	 turn."
//
// A manland built like Raging Ravine: ONE continuous effect from the
// resolving ability (CR 611.2a), pinned to this object (CR 611.2c,
// CR 400.7), applied in each layer it touches at one timestamp
// (CR 613.7): the creature type and Elemental subtype (layer 4), blue and
// black (layer 5), the base 3/2 (layer 7b). "It can't be blocked this
// turn" is a restriction on the same object in the same effect, so it
// ends with the turn like the rest.
//
// The engine supplies the rest: a land that came under its controller's
// control this turn can't attack, and its {T} mana ability can't be
// activated while it is a creature (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "250cb58b-2924-4dff-92fe-ac0ebbbeb218",
		Name:         "Creeping Tar Pit",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U|B}",
			Label:    "Add {U} or {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}{B}: Until end of turn, this land becomes a 3/2 blue and black Elemental creature. It's still a land. It can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{U}{B}"),
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
						game.SetColorsMod(ctx.TextColors("U", "B")...),
						game.AddRestrictionsMod(game.CantBeBlocked),
					}, game.SetBasePTMods(3, 2)...),
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Creeping Tar Pit — a 3/2 blue and black Elemental creature that can't be blocked until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
