package roadmap

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// declaresUnattachEquipment is the "Unattach an Equipment from <this>"
// cost probe (#4799): an activated ability whose cost unattaches an
// Equipment from its source (game.AbilityCost.UnattachEquipment, built by
// effects.UnattachAnEquipment).
func declaresUnattachEquipment(s effects.Spec) bool {
	for _, a := range s.Activated {
		if a.Cost.UnattachEquipment > 0 {
			return true
		}
	}
	return false
}
