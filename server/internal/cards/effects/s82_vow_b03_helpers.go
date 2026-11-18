package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Helpers shared by the Blood-token cards of S82's Crimson Vow batch 3.

// s82Vow03CreateBlood is "create N Blood token(s)" for the ability's controller.
func s82Vow03CreateBlood(n int) Applier {
	return CreateToken{Template: BloodToken(), N: n}
}

// s82Vow03ThisDealtCombatDamage is "whenever this creature deals combat
// damage": any recipient (CR 510.2), the source being the creature the
// ability is on.
func s82Vow03ThisDealtCombatDamage(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventDealDamage && ev.Combat && ev.Amount > 0 && ev.Source == source.InstanceID
}
