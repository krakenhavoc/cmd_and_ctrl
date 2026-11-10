package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Walking Bulwark — Artifact Creature — Golem {1}, 0/3:
//
//	"Defender
//	 {2}: Until end of turn, target creature with defender gains haste, can
//	 attack as though it didn't have defender, and assigns combat damage
//	 equal to its toughness rather than its power. Activate only as a
//	 sorcery."
//
// One scoped effect pinned to the target at resolution (CR 611.2c) with three
// layer 6 mods: the haste keyword, a defender waiver (the keyword stays, so
// the creature is still a creature with defender, CR 702.3b), and
// game.ModDamageByToughness, which changes how combat damage is assigned and
// not the creature's power (the ruling), and applies even when the toughness
// is the lesser. The target must have defender both when chosen and as the
// ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8ad0d89b-beef-40f7-8ce8-65589e3970be",
		Name:            "Walking Bulwark",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:        "{2}: Until end of turn, target creature with defender gains haste, can attack as though it didn't have defender, and assigns combat damage equal to its toughness rather than its power. Activate only as a sorcery.",
			Cost:         ManaCost("{2}"),
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature with defender", HasKeyword("defender")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return untilEndOfTurn(ctx, t.ID, nil, "Walking Bulwark — haste, no defender restriction, damage by toughness",
						game.AddKeywordsMod("haste"), game.WaiveDefenderMod(), game.DamageByToughnessMod())
				}
				return nil
			},
		}},
	})
}
