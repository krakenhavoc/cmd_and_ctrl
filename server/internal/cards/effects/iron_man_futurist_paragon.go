package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iron Man, Futurist Paragon — Legendary Artifact Creature — Human Hero
// {4}{U}{U}, 5/5 (Marvel Super Heroes Commander):
//
//	"Flying
//	 At the beginning of combat on your turn, target artifact or creature
//	 becomes an artifact creature with base power and toughness 5/5 and gains
//	 flying."
//
// No duration is printed, so the change is permanent (CR 611.2a): one scoped
// effect pinned to the object, so a target that was flickered in response is
// left alone (CR 400.7). It adds the Artifact and Creature types (a creature
// target stays what it was, a noncreature artifact becomes a creature), sets
// base 5/5 in layer 7b, so counters and pumps still add on top, and grants
// flying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eabe9a02-ea47-4548-8933-e5d1269fab25",
		Name:            "Iron Man, Futurist Paragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat("Iron Man — target artifact or creature becomes a 5/5 artifact creature with flying",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					mods := []game.Mod{game.AddTypesMod("Artifact", "Creature")}
					mods = append(mods, game.SetBasePTMods(5, 5)...)
					mods = append(mods, game.AddKeywordsMod("flying"))
					return ScopedEffectFor{
						Target:   id,
						Mods:     mods,
						Duration: g.PinnedTo(game.IndefiniteDuration(), id),
						Label:    "Iron Man — becomes a 5/5 artifact creature with flying",
					}.Apply(ctx)
				}),
				TargetPermanent("target artifact or creature", Or(Artifact(), Creature()))),
		},
	})
}
