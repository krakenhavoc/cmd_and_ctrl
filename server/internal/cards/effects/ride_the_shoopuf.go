package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ride the Shoopuf — Enchantment {1}{G} (Final Fantasy, #4462):
//
//	"Landfall — Whenever a land you control enters, put a +1/+1 counter on
//	 target creature you control.
//	 {5}{G}{G}: This enchantment becomes a 7/7 Beast creature in addition
//	 to its other types."
//
// The landfall is Grappling Kraken's shape (Landfall, targeting on the
// stack) over the shared +1/+1 counter body; a target that left in
// response gets nothing (CR 608.2b). The animation has no duration (CR
// 611.2a, the ruling), so it is one effect at one timestamp, pinned to
// this object (CR 611.2c, 400.7) and indefinite: the Creature type and
// the Beast subtype are added in layer 4 and the base power and toughness
// set to 7/7 in layer 7b. The enchantment keeps its Enchantment type, so
// it is an enchantment creature. If it was not under its controller's
// control since their turn began it can't attack that turn (CR 302.6,
// the engine's).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "81b67745-0b1f-4433-963d-9b9653fb4d95",
		Name:         "Ride the Shoopuf",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(Landfall("Ride the Shoopuf — put a +1/+1 counter on target creature you control", PlusOneCounterOnTheTarget),
				TargetCreature("target creature you control", YouControl())),
		},
		Activated: []ActivatedAbility{{
			Label:   "{5}{G}{G}: This enchantment becomes a 7/7 Beast creature in addition to its other types.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{5}{G}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				self := item.SourceCardID
				if !onBattlefield(g, self) || sourceIsNewObject(g, item) {
					return nil
				}
				return ScopedEffectFor{
					Target: self,
					Mods: []game.Mod{
						game.AddTypesMod("Creature"),
						// A word of the enchantment's text (CR 612.2, #3984).
						game.AddSubtypesMod(g.TextReaderForEffect(item).TextSubtypes("Beast")...),
						game.SetBasePowerMod(7),
						game.SetBaseToughnessMod(7),
					},
					Duration: g.PinnedTo(game.IndefiniteDuration(), self),
					Label:    "Ride the Shoopuf — becomes a 7/7 Beast creature",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
