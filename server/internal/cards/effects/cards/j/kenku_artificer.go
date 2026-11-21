package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kenku Artificer — Creature — Bird Artificer {2}{U}, 1/1 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Homunculus Servant — When this creature enters, put three +1/+1 counters
//	 on up to one target noncreature artifact. That artifact becomes a 0/0
//	 Homunculus artifact creature with flying."
//
// Rootwise Survivor's shape with no end: the counters go on first, then an
// indefinite effect (CR 611.2a, no duration) makes the artifact a 0/0 creature
// with the Homunculus subtype and flying, in addition to its other types. The
// counters are what keep it alive (CR 704.5f), 3/3 in total. The target is
// checked as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e262b226-9463-40a9-95b2-fe2c3eafc043",
		Name:         "Kenku Artificer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Kenku Artificer — put three +1/+1 counters on up to one target noncreature artifact; it becomes a 0/0 Homunculus artifact creature with flying",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 3}).Apply(ctx); err != nil {
							return err
						}
						mods := append([]game.Mod{
							game.AddTypesMod("Creature"),
							game.AddSubtypesMod(ctx.TextSubtype("Homunculus")),
							game.AddKeywordsMod("flying"),
						}, game.SetBasePTMods(0, 0)...)
						return effects.ScopedEffectFor{
							Target:   id,
							Mods:     mods,
							Duration: game.IndefiniteDuration(),
							Label:    "Kenku Artificer — becomes a 0/0 Homunculus artifact creature with flying",
						}.Apply(ctx)
					}),
				effects.TargetPermanent("up to one target noncreature artifact",
					effects.And(effects.Artifact(), effects.Noncreature())).WithCount(0, 1)),
		},
	})
}
