package roadmap

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// declaresPowerOrToughnessPermission is the "cast creature spells with
// power or toughness N or less" probe (#4964): a cast permission whose
// filter carries the power-or-toughness clause
// (game.PermissionFilter.PowerOrToughnessCapped).
func declaresPowerOrToughnessPermission(s effects.Spec) bool {
	for _, p := range s.CastPermissions {
		if p.Filter.PowerOrToughnessCapped {
			return true
		}
	}
	return false
}
