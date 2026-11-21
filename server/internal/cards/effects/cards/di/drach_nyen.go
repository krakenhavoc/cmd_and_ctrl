package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drach'Nyen — Legendary Artifact — Equipment {4}{B}{R} (Warhammer 40,000
// Commander, S82 GAP batch 186, #5948):
//
//	"Echo of the First Murder — When Drach'Nyen enters, exile up to one
//	 target creature.
//	 Daemon Sword — Equipped creature has menace and gets +X/+0, where X
//	 is the exiled card's power.
//	 Equip {2}"
//
// The exile is linked to this Drach'Nyen object (CR 607.2a, ExileTarget
// Linked). The bonus is read live from the card still exiled with it, with
// the power it has in exile (CR 208.3: a card in exile has its printed
// power); a creature that left exile, or none exiled, makes X zero. A
// Drach'Nyen that left and came back is a new object and remembers nothing
// (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aaf4a72b-9606-4d8c-a286-fbce4aa08f97",
		Name:         "Drach'Nyen",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.GrantToAttached("menace"),
			effects.PumpAttachedPer(1, 0, func(g *game.Game, source *game.Card) int {
				n := 0
				for _, c := range g.ExiledWithForEffect(game.PermissionCardRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}) {
					n += max(c.Power, 0)
				}
				return n
			}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Drach'Nyen — exile up to one target creature", effects.ExileLinkedTargetsEffect),
				effects.UpToOneTargetCreature("up to one target creature")),
		},
	})
}
