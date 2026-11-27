package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Torii Watchward — Creature — Spirit {4}{W}, 3/3 (Saviors of Kamigawa,
// #5601):
//
//	"Vigilance
//	 Soulshift 4 (When this creature dies, you may return target Spirit card
//	 with mana value 4 or less from your graveyard to your hand.)"
//
// Vigilance and Soulshift are the engine's (CR 702.20a, 702.46a).
//
// No simplification.
func init() {
	effects.ChkB1Soulshifter("8b52c2ed-c1a1-4abc-a33b-223c3d0e229d", "Torii Watchward", 4, "vigilance")
}
