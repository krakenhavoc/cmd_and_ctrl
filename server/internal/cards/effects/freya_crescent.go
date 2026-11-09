package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Freya Crescent — Legendary Creature — Rat Knight {R}, 1/1 (Final
// Fantasy, #4462):
//
//	"Jump — During your turn, Freya Crescent has flying.
//	 {T}: Add {R}. Spend this mana only to cast an Equipment spell or
//	 activate an equip ability."
//
// Jump is Cid, Freeflier Pilot's turn-gated flying grant on the creature
// itself. The mana is Ronin, Shadow Stalker's one restriction with two
// halves (CR 106.6): an Equipment spell, or the cost of an equip ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e3085d8-9d46-41e0-ad9d-54c44d9ef267",
		Name:         "Freya Crescent",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.IsCreature() && isActivePlayer(g, source.Controller)
			}, "flying"),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "{T}: Add {R}. Spend this mana only to cast an Equipment spell or activate an equip ability.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Equipment")},
				[]string{game.ManaRestrictEquip},
			)},
		}},
	})
}
