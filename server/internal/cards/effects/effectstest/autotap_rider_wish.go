package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// autotap_rider_wish_test.go — #3853, ADR 0068 amendment 2026-10-15:
// auto-pay picks the mana that does something, through the real cards.
// The engine half (every shape, in both orders, and the payability bound)
// is game/autotap_wish_test.go; the bot's use of the same planner is
// aiseat/autopay_rider_bot_test.go.

const (
	WishSolRingOracle      = "6ad8011d-3471-4369-9d68-b264cc027487"
	WishGruulScrapper      = "878e8980-a2cc-4335-aafe-f5166ef48f79"
	WishMoonholdOracle     = "9b515bb8-6e37-4b2d-8d56-bb517c3b267c"
	WishRibbonsOracle      = "f074c0d0-8455-484d-ae75-820fdfbc4740"
	WishRiversGraspOracle  = "f671583f-bd22-4d3e-bf19-fb740c607f8f"
	WishBorosFuryShield    = "20eff2ce-f26d-48a2-8a1f-435a7e2968c8"
	wishBoseijuShelters    = "36937483-30cb-449a-8028-75017a124922"
	WishGeneratorServantID = "68ac061a-e4a3-46df-9169-62f6a8481584"
)

func PermanentTapped(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Tapped
		}
	}
	return false
}
