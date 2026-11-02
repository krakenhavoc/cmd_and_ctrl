package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doc Ock, Evil Inventor — {5}{U}{B} Legendary Creature — Human Scientist
// Villain, 8/8 (Marvel's Spider-Man Eternal, #4460):
//
//	"At the beginning of combat on your turn, target noncreature artifact
//	 you control becomes an 8/8 Robot Villain artifact creature in addition
//	 to its other types."
//
// No duration is stated, so the change lasts as long as that permanent is
// this object (CR 611.2a, 400.7), the way the hidden-enchantment creatures
// do (enchantment_becomes_creature.go). "In addition to its other types"
// keeps Artifact and adds Creature plus the subtypes Robot and Villain; the
// base P/T is set to 8/8 (CR 613.4b), counters and pumps still land on it.
//
// No simplification.
func init() {
	const label = "Doc Ock, Evil Inventor — target noncreature artifact you control becomes an 8/8 Robot Villain artifact creature"
	Register(Spec{
		OracleID:     "3e0738af-00ec-4707-a56d-85b5a5ce4529",
		Name:         "Doc Ock, Evil Inventor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat(label, docOckAnimate),
				TargetPermanent("target noncreature artifact you control", Artifact(), Noncreature(), YouControl())),
		},
	})
}

func docOckAnimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || !onBattlefield(g, t.ID) {
			continue
		}
		if err := (ScopedEffectFor{
			Target: t.ID,
			Mods: []game.Mod{
				game.AddTypesMod("Creature"),
				game.AddSubtypesMod(g.TextReaderForEffect(item).TextSubtypes("Robot", "Villain")...),
				game.SetBasePowerMod(8),
				game.SetBaseToughnessMod(8),
			},
			Duration: g.PinnedTo(game.IndefiniteDuration(), t.ID),
			Label:    "Doc Ock — becomes an 8/8 Robot Villain artifact creature",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
