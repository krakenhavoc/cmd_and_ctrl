package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archfiend of Sorrows — Creature — Demon {5}{B}{B}, 4/5 (Modern Horizons
// 2, #5501):
//
//	"Flying
//	 When this creature enters, creatures your opponents control get
//	 -2/-2 until end of turn.
//	 Unearth {3}{B}{B} ({3}{B}{B}: Return this card from your graveyard to
//	 the battlefield. It gains haste. Exile it at the beginning of the
//	 next end step or if it would leave the battlefield. Unearth only as
//	 a sorcery.)"
//
// Doomwake Giant's -N/-N over the opponents' creatures, chosen once as
// the trigger resolves (CR 611.2c), and Terror Ballista's unearth row
// (unearth.go, CR 702.84a), which enters it again and so fires the
// trigger again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fe84966e-0f0c-4df8-b61c-e9693492dfd8",
		Name:            "Archfiend of Sorrows",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2, OpponentsOnly: true}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Archfiend of Sorrows — creatures your opponents control get -2/-2 until end of turn",
				effects.Do(effects.BoostUntilEOT{
					Match:     effects.And(effects.Creature(), effects.OpponentControls()),
					Power:     -2,
					Toughness: -2,
					Label:     "Archfiend of Sorrows — -2/-2 until end of turn",
				})),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{3}{B}{B}")},
	})
}
