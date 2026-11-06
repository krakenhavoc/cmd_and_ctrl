package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// A Realm Reborn — Enchantment {4}{G}{G} (Final Fantasy, #4462):
//
//	"Other permanents you control have "{T}: Add one mana of any color.""
//
// Cryptolith Rite's layer-6 grant (ADR 0093) widened to every other
// permanent its controller controls: lands, artifacts, creatures,
// planeswalkers and the rest each have the mana ability as their own.
// A creature still can't use it the turn it arrives unless it has haste
// (CR 302.6). The enchantment itself is excluded ("other"), by object.
//
// No simplification.
const aRealmRebornGrant = "a-realm-reborn/any-color"

func init() {
	Register(Spec{
		OracleID:     "fd230ab5-874e-4827-9b5c-6d39d079cee6",
		Name:         "A Realm Reborn",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{AnyColorManaGrant(aRealmRebornGrant)},
		Static: []game.StaticAbility{Labelled(GrantAbilities(otherPermanentsYouControl, aRealmRebornGrant),
			"Other permanents you control have \"{T}: Add one mana of any color.\"")},
	})
}

// otherPermanentsYouControl is "other permanents you control".
func otherPermanentsYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.InstanceID != source.InstanceID && target.Controller == source.Controller
}
