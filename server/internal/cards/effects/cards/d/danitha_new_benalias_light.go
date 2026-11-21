package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Danitha, New Benalia's Light — Legendary Creature — Human Knight {1}{G}{W},
// 2/2:
//
//	"Vigilance, trample, lifelink
//	 Once during each of your turns, you may cast an Aura or Equipment
//	 spell from your graveyard."
//
// A STANDING graveyard permission (ADR 0066) derived off the battlefield.
// The one use per turn is Serra Paragon's budget (PermissionTypePermanent,
// one use however it is spent): an Aura is an enchantment and an Equipment
// an artifact, so a per-card-type budget would hand out two. The filter is
// AuraOrEquipmentOnly (Sigarda's Aid), TimingYourTurnOnly closes it off
// your turn and leaves the spell's own timing in force, and the spell's
// costs are paid as usual. A second Danitha opens one more use, as the
// budget is per granting object.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "971b30a8-4ab4-4e96-9d50-597da4d52994",
		Name:            "Danitha, New Benalia's Light",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample", "lifelink"},
		CastPermissions: []game.CastPermission{{
			Zone:     game.ZoneGraveyard,
			CastOnly: true,
			Filter:   game.PermissionFilter{AuraOrEquipmentOnly: true},
			PerType:  []string{game.PermissionTypePermanent},
			Timing:   game.TimingYourTurnOnly,
			Label:    "Once during each of your turns, cast an Aura or Equipment spell from your graveyard (Danitha, New Benalia's Light)",
		}},
	})
}
