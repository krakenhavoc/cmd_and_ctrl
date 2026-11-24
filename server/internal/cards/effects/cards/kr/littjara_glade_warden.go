package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Littjara Glade-Warden — Creature — Shapeshifter {3}{G}, 3/3:
//
//	"Changeling (This card is every creature type.)
//	 {2}{G}, {T}, Exile a creature card from your graveyard: Put two +1/+1
//	 counters on target creature. Activate only as a sorcery."
//
// Moorland Haunt's exile-a-creature-card cost (a graveyard with none offers
// nothing to pay with) on a sorcery-speed ability with a target. The Warden
// taps, so it can't be activated the turn it arrives (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "438a0a7a-00bc-4d93-8e69-a8a6084523c5",
		Name:            "Littjara Glade-Warden",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"changeling"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{G}, {T}, Exile a creature card from your graveyard: Put two +1/+1 counters on target creature. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}{G}"), effects.TapCost(), effects.ExileFromGraveyard(1, "a creature card", effects.MatchCreature)),
			Targets:      effects.TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect:       effects.TwoPlusOneCountersOnTheTarget,
		}},
	})
}
