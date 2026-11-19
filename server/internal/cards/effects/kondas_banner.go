package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Konda's Banner — Legendary Artifact — Equipment {2} (Champions of
// Kamigawa, #5846):
//
//	"Konda's Banner can be attached only to a legendary creature.
//	 Creatures that share a color with equipped creature get +1/+1.
//	 Creatures that share a creature type with equipped creature get
//	 +1/+1.
//	 Equip {2}"
//
// The first line is Spec.AttachOnlyTo, which the engine holds every
// attach to (game.attachmentLegalLocked): an equip or an effect that
// would attach the Banner to a nonlegendary creature leaves it where it
// is (CR 301.5b, CR 701.3b), and a creature that stops being legendary
// drops it as a state-based action (CR 301.5c, CR 704.5n). The equip
// ability is printed "Equip {2}", so it still targets any creature you
// control; aimed at a nonlegendary one it resolves and does nothing, as
// the rulings on Gate Smasher and O-Naginata (the same clause) say.
//
// The two anthems are layer 7c modifications (CR 613.4c), each its own
// static, so a creature that shares a color AND a creature type gets
// +2/+2 and no more however many it shares (the 2004-12-01 rulings). The
// equipped creature counts for both whenever it has a color or a
// creature type (it shares them with itself). "Share" reads current
// characteristics: colors through sharesAColorNow, creature types
// through game.SharesCreatureType (changelings included, "Artifact"
// never). Whose creatures doesn't matter: every creature on the
// battlefield is in. An unattached Banner pumps nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a355908b-4359-4bfe-9bc4-ccf721ab3a13",
		Name:         "Konda's Banner",
		Completeness: CompletenessFull,
		AttachOnlyTo: AttachedOnlyTo("a legendary creature", Legendary()),
		Static: []game.StaticAbility{
			kondasBannerAnthem("Creatures that share a color with equipped creature get +1/+1.",
				func(target, host *game.Card) bool { return sharesAColorNow(*target, *host) }),
			kondasBannerAnthem("Creatures that share a creature type with equipped creature get +1/+1.",
				game.SharesCreatureType),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

// kondasBannerAnthem is "Creatures that <relation> equipped creature get
// +1/+1": every creature `shares` relates to the creature the source is
// attached to, the equipped creature itself included.
func kondasBannerAnthem(label string, shares func(target, host *game.Card) bool) game.StaticAbility {
	return game.StaticAbility{
		Label:    label,
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			if g == nil || source == nil || !target.IsCreature() {
				return false
			}
			host := g.AttachedHostOf(source)
			return host != nil && host.IsCreature() && shares(target, host)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power++
			c.Toughness++
		},
	}
}
