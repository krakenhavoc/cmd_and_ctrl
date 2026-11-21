package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wicker Picker — Artifact Creature — Scarecrow Guest {3}, 2/3:
//
//	"Creature spells you cast have sticker kicker {1}. (You may pay an
//	 additional {1} as you cast a creature spell. If you do, you get {TK},
//	 then you may put a sticker on it.)"
//
// A permanent's "<spells> you cast have <optional cost>" is
// Spec.GrantedOptionalCosts (#3245, Silverquill, the Disputant): the offer
// is derived from the battlefield on every cast, so it lasts as long as
// this is under the caster's control. Sticker kicker (CR 702.33h) is a
// kicker whose second half is an additional cost: the caster gets a ticket
// counter and may put a sticker on the spell, as it becomes cast. The
// sticker is on the spell card on the stack and stays on it as the creature
// enters (CR 123.5), so it is on the permanent. It is a kicker of its own
// (game.StickerKickerKey), so a creature spell that prints a kicker offers
// both, and "if it was kicked" on that creature still means its own kicker
// (CR 702.33e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15a71d11-da00-4c1a-bf64-ec3f627ed6a3",
		Name:         "Wicker Picker",
		Completeness: effects.CompletenessFull,
		GrantedOptionalCosts: []game.GrantedOptionalCost{
			effects.SpellsYouCastHave("Creature spells you cast have sticker kicker {1}.", effects.Creature(), effects.StickerKicker("{1}")),
		},
	})
}
