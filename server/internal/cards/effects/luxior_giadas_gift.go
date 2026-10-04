package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Luxior, Giada's Gift — Legendary Artifact — Equipment {1}:
//
//	"Equipped creature gets +1/+1 for each counter on it.
//	 Equipped permanent isn't a planeswalker and is a creature in
//	 addition to its other types. (Loyalty abilities can still be
//	 activated.)
//	 Equip planeswalker {1}
//	 Equip {3}"
//
// Two statics read the same attachment and sit in different layers
// (CR 613.1d and 613.4c): the type change in layer 4, which removes
// Planeswalker and adds Creature, and the +1/+1 for each counter in
// layer 7c. A planeswalker has no printed power or toughness, so as a
// creature it is 0/0 until the 7c bonus lands; every counter on it
// counts, loyalty first, so a four-loyalty walker is a 4/4. Because
// the +1/+1 is read off the counters in the same recompute, a loyalty
// change moves the size with it.
//
// "Equipped permanent" being a creature is what makes the attachment
// legal at all: the equipment state-based action (CR 704.5n, CR 301.5c)
// asks whether the host is a creature in the layered view, and layer 4
// has already said yes by then. A planeswalker that stops being a
// planeswalker is not subject to the zero-loyalty rule (CR 704.5i), and
// its loyalty abilities can still be activated, as the reminder text says
// (CR 606.3). "Equip planeswalker" is a second equip
// ability with its own cost and a planeswalker-you-control target;
// ordinary "Equip {3}" targets a creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f9263d7-916f-4535-95d2-888ab73cf339",
		Name:         "Luxior, Giada's Gift",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer4Type,
				AppliesTo: AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					kept := make([]string, 0, len(c.Types)+1)
					creature := false
					for _, t := range c.Types {
						switch t {
						case "Planeswalker":
							continue
						case "Creature":
							creature = true
						}
						kept = append(kept, t)
					}
					if !creature {
						kept = append(kept, "Creature")
					}
					c.Types = kept
				},
			},
			{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: AttachedToSource,
				Apply: func(c *game.Characteristic, host *game.Card, _ *game.Game, _ *game.Card) {
					n := 0
					for _, v := range host.Counters {
						n += v
					}
					c.Power += n
					c.Toughness += n
				},
			},
		},
		Activated: []ActivatedAbility{
			EquipOnlyAbility("Equip planeswalker {1}", "{1}",
				TargetPermanent("target planeswalker you control", Planeswalker(), YouControl())),
			EquipAbility("{3}"),
		},
	})
}
