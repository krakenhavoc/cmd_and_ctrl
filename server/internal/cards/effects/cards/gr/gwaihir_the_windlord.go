package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gwaihir the Windlord — Legendary Creature — Bird Noble {4}{W}{U}, 4/4 (The
// Lord of the Rings: Tales of Middle-earth Commander):
//
//	"This spell costs {2} less to cast as long as you've drawn two or more
//	 cards this turn.
//	 Flying, vigilance
//	 Other Birds you control have vigilance."
//
// The reduction is a self cost modifier (Avenge's shape) reading the
// caster's draw tally as the spell is priced. Flying and vigilance are
// printed; the last line is a tribal keyword grant (TribalKeywordGrant) to
// the other Birds its controller controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b967fbb8-ddc8-485b-bd18-cbf3ef6b1bf8",
		Name:            "Gwaihir the Windlord",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast as long as you've drawn two or more cards this turn.", effects.GapB127DrewTwoOrMoreCards),
		},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Bird"}, Others: true, YoursOnly: true}, "vigilance"),
		},
	})
}
