package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kess, Dissident Mage — Legendary Creature — Human Wizard {1}{U}{B}{R},
// 3/4:
//
//	"Flying
//	 Once during each of your turns, you may cast an instant or sorcery
//	 spell from your graveyard. If a spell cast this way would be put
//	 into your graveyard, exile it instead."
//
// Serra Paragon's one-use budget (PermissionTypePermanent is a single
// use however it is spent, and the filter keeps it to instants and
// sorceries, so one cast of either kind uses it up) on a graveyard
// permission closed off your turn. The exile is Torrential Gearhulk's
// replacement, carried on the spell: it applies however the spell
// leaves the stack for the graveyard (resolves, is countered, fizzles).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5092c14-eec4-472c-999c-ba96c36b2fbb",
		Name:         "Kess, Dissident Mage",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:                    game.ZoneGraveyard,
			CastOnly:                true,
			Filter:                  game.PermissionFilter{InstantOrSorceryOnly: true},
			AltCostKey:              "kess_dissident_mage",
			PerType:                 []string{game.PermissionTypePermanent},
			Timing:                  game.TimingYourTurnOnly,
			ExileInsteadOfGraveyard: true,
			Label:                   "Once during each of your turns, cast an instant or sorcery spell from your graveyard; if it would be put into your graveyard, exile it instead (Kess, Dissident Mage)",
		}},
	})
}
