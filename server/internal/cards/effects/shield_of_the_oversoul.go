package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shield of the Oversoul — Enchantment — Aura {2}{G/W} (EDHREC rank
// 4503):
//
//	"Enchant creature
//	 As long as enchanted creature is green, it gets +1/+1 and has
//	 indestructible.
//	 As long as enchanted creature is white, it gets +1/+1 and has
//	 flying."
//
// A three-mana Aura that gives a Selesnya creature +2/+2, flying AND
// indestructible — because a green-white creature is both colours and
// both clauses apply at once. That is the whole card, and it is why
// it is a Voltron staple: an indestructible flier carrying a
// commander's damage is very hard to answer with anything but exile
// or a bounce.
//
// The conditions are read LIVE off the enchanted creature's
// post-layer colours on every recompute, not fixed when the Aura
// resolves. Two consequences, both printed:
//
//   - Put it on a mono-green creature and it is +1/+1 and
//     indestructible. Turn that creature white later — with a
//     Celestial Dawn, or by copying a white creature — and the
//     flying half switches on by itself.
//   - Make the creature colourless (Song of the Dryads, a
//     Mycosynth Lattice) and BOTH halves switch off. The Aura stays
//     attached and does nothing, exactly as the card reads.
//
// Colour is read in layer 5, which runs before layer 6 (abilities)
// and layer 7 (P/T), so an effect that changes the creature's colour
// is already applied when these clauses ask. That ordering is why the
// two halves can be written as ordinary conditional statics and do
// not need a dependency declaration (CR 613.8).
//
// Four statics rather than two, because a P/T change and an ability
// grant live in different layers and one entry cannot sort into
// both.
//
// No simplification.
func init() {
	green := b43AttachedHasColor("G")
	white := b43AttachedHasColor("W")
	Register(Spec{
		OracleID:     "9b598025-80a3-4144-be9a-863165161594",
		Name:         "Shield of the Oversoul",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			b43PumpAttachedWhen(green, 1, 1),
			b43GrantToAttachedWhen(green, "indestructible"),
			b43PumpAttachedWhen(white, 1, 1),
			b43GrantToAttachedWhen(white, "flying"),
		},
	})
}
