package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soulstone Sanctuary — Land (Foundations, #4465):
//
//	"{T}: Add {C}.
//	 {4}: This land becomes a 3/3 creature with vigilance and all
//	 creature types. It's still a land."
//
// Mutavault's animation with no "until end of turn": the change has no
// duration, so it lasts as long as the land is this object (CR 611.2c,
// 400.7), pinned to it the way Emergent Haunting's is. One effect at one
// timestamp: the Creature type and every creature type in layer 4
// (AllCreatureTypesMod; it does not gain changeling), vigilance in layer
// 6 and a base 3/3 in layer 7b. The land stays colourless and keeps its
// mana ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c0f99b8-0222-4fac-932a-eb5d77826564",
		Name:         "Soulstone Sanctuary",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}: This land becomes a 3/3 creature with vigilance and all creature types. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{4}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AllCreatureTypesMod(),
						game.AddKeywordsMod("vigilance"),
					}, game.SetBasePTMods(3, 3)...),
					Duration: g.PinnedTo(game.IndefiniteDuration(), item.SourceCardID),
					Label:    "Soulstone Sanctuary — a 3/3 creature with vigilance and all creature types",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
