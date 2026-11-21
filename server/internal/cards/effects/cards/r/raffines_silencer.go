package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raffine's Silencer — Creature — Human Assassin {2}{B}, 1/1 (Streets of
// New Capenna, #5491):
//
//	"When this creature enters, it connives. (Draw a card, then discard a
//	 card. If you discarded a nonland card, put a +1/+1 counter on this
//	 creature.)
//	 When this creature dies, target creature an opponent controls gets
//	 -X/-X until end of turn, where X is this creature's power."
//
// Two printed triggers, two registered. X is the Silencer's power as it last
// existed on the battlefield (CR 608.2h), counters and layers included, and
// is read when the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2901948c-2a5a-4135-9135-27406e9d8652",
		Name:         "Raffine's Silencer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersItConnives("Raffine's Silencer — it connives"),
			effects.Targeting(
				effects.WhenThisDies("Raffine's Silencer — target creature an opponent controls gets -X/-X until end of turn, where X is its power",
					raffinesSilencerShrink),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}

func raffinesSilencerShrink(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	x := effects.B17LastKnownPowerOffBattlefield(g, item.SourceCardID)
	if info, ok := ctx.TriggeringPermanent(); ok {
		x = max(info.Power, 0)
	}
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok || x <= 0 {
		return nil
	}
	return effects.BoostUntilEOT{Target: id, Power: -x, Toughness: -x, Label: "Raffine's Silencer — -X/-X"}.Apply(ctx)
}
