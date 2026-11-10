package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Collective Inferno — Enchantment {3}{R}{R} (Lorwyn Eclipsed, #4458):
//
//	"Convoke
//	 As this enchantment enters, choose a creature type.
//	 Double all damage that sources you control of the chosen type would
//	 deal."
//
// Convoke is the shared tap cost (CR 702.51). The type is chosen as it
// enters (ChooseCreatureTypeAsEnters, Door of Destinies). The doubling is
// Angrath's Marauders' damage replacement (CR 614.1), for a damage source
// its controller controls that has the chosen creature type as it is, or as
// it last was (a changeling is every type), whatever it damages. It is the
// same source that deals the doubled damage, not the enchantment, so
// lifelink, deathtouch and "dealt by" triggers see the original source
// (the ruling). With no type chosen yet it doubles nothing. Several
// replacements are ordered by the damaged player or permanent's controller
// (the ruling, CR 616.1).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "649e4969-71d4-434f-85b0-205ec2c6b24f",
		Name:         "Collective Inferno",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		AsEnters:     ChooseCreatureTypeAsEnters("Collective Inferno"),
		Replacements: []game.ReplacementEffect{{
			Watches:  []game.EventKind{game.EventDealDamage},
			RowLabel: "Damage that sources you control of the chosen type would deal is doubled.",
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || src.NamedTribe == "" {
					return false
				}
				ch, ok := damageSourceCharacteristics(ev, g)
				return ok && ch.Controller == src.Controller && stdB66DamageSourceHasType(ch, src.NamedTribe)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
		}},
	})
}
