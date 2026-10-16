package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cyberdrive Awakener — Artifact Creature — Construct {5}{U}, 4/4
// (EDHREC rank 970):
//
//	"Flying
//	 Other artifact creatures you control have flying.
//	 When this creature enters, each noncreature artifact you control
//	 becomes a 4/4 artifact creature until end of turn."
//
// The static is a layer-6 keyword grant (KeywordGrant) over the other
// artifact creatures you control, read in the layer pass, so an
// artifact the trigger animated in layer 4 is a creature by the time
// layer 6 asks and flies with the rest.
//
// The trigger is The Blackstaff of Waterdeep's animation (a layer-4
// type addition and a layer-7b base power and toughness) over a set
// locked as it resolves (CR 611.2c): each artifact you control that is
// not a creature then. An artifact that arrives later stays as it is,
// and the animated ones stay 4/4 creatures until the cleanup step even
// if the Awakener leaves. They keep every type they had.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "126a5d0b-e7d6-458e-87af-47d845678014",
		Name:            "Cyberdrive Awakener",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			Labelled(KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
					target.IsArtifact() && target.IsCreature()
			}, "flying"), "Other artifact creatures you control have flying."),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cyberdrive Awakener — each noncreature artifact you control becomes a 4/4 artifact creature until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return ScopedEffectFor{
						Match: And(Artifact(), Not(Creature()), YouControl()),
						Mods: []game.Mod{
							game.AddTypesMod("Artifact", "Creature"),
							game.SetBasePowerMod(4),
							game.SetBaseToughnessMod(4),
						},
						Duration: DurationUntilEndOfTurn(ctx),
						Label:    "Cyberdrive Awakener — a 4/4 artifact creature until end of turn",
					}.Apply(ctx)
				}),
		},
	})
}
