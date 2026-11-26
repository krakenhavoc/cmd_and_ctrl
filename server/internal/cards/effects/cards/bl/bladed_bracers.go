package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bladed Bracers — Artifact — Equipment {1}:
//
//	"Equipped creature gets +1/+1.
//	 As long as equipped creature is a Human or an Angel, it has vigilance.
//	 Equip {2}"
//
// The vigilance is conditional on the host's layered subtypes, so a host
// that changes type gains or loses it at once.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "24d373cd-245a-4051-b187-ba8cc83b0c48",
		Name:         "Bladed Bracers",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttachedWhile(func(host *game.Card, _ *game.Game, source *game.Card) bool {
				return host.HasSubtype(source.TextSubtype("Human")) || host.HasSubtype(source.TextSubtype("Angel"))
			}, "vigilance"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
