package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unable to Scream — Enchantment — Aura {U} (Duskmourn: House of Horror,
// #4466):
//
//	"Enchant creature
//	 Enchanted creature loses all abilities and is a Toy artifact
//	 creature with base power and toughness 0/2 in addition to its other
//	 types.
//	 As long as enchanted creature is face down, it can't be turned face
//	 up."
//
// The first static is one effect across three layers (CR 613.6), each in
// the layer the rules put it, as Kenrith's Transformation's is:
//
//   - layer 4: artifact and creature card types and the Toy creature
//     type, ADDED to what it already is (Super-Soldier Serum's additive
//     shape, not SetAttachedTypes's replacement); "Toy" is a word of the
//     Aura's text (CR 612.2), read through it;
//   - layer 6: loses all abilities;
//   - layer 7b: base power and toughness 0/2, so counters and pumps still
//     apply on top.
//
// The second is a TurnFaceUpRestriction scoped to the creature the Aura
// is attached to (Karlov Watchdog's seam, game/turn_face_up_gate.go):
// read live, it shuts the turn-face-up special action, whatever cost it
// would pay, and an effect that would turn it face up does nothing to it
// (CR 101.2). The restriction is only asked of a face-down permanent, so
// it is in force exactly "as long as" the creature is face down, and the
// Aura leaving lifts it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d240726-5214-4493-8537-29bf65db1e4b",
		Name:         "Unable to Scream",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			{
				Label:                 "Enchanted creature is a Toy artifact creature in addition to its other types.",
				Layer:                 game.Layer4Type,
				ContinuesAfterRemoval: true,
				AppliesTo:             AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					for _, t := range []string{"Artifact", "Creature"} {
						if !hasFold(c.Types, t) {
							c.Types = append(c.Types, t)
						}
					}
					if toy := source.TextSubtype("Toy"); !hasFold(c.Subtypes, toy) {
						c.Subtypes = append(c.Subtypes, toy)
					}
				},
			},
			LoseAllAbilities(),
			SetAttachedBasePT(0, 2),
		},
		TurnFaceUpRestrictions: []game.TurnFaceUpRestriction{{
			Label: "As long as enchanted creature is face down, it can't be turned face up.",
			Forbids: func(q game.TurnFaceUpQuery) bool {
				return q.Source.IsAttachedTo(q.Card.InstanceID)
			},
		}},
	})
}
