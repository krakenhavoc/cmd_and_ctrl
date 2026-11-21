package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Valiant Veteran — Creature — Kor Soldier {1}{W}, 2/2 (Dominaria United,
// S82 DMU batch 4):
//
//	"Other Soldiers you control get +1/+1.
//	 {3}{W}{W}, Exile this card from your graveyard: Put a +1/+1 counter on
//	 each Soldier you control."
//
// Imperious Perfect's layer 7c lord over OTHER Soldiers, and a graveyard
// activation (Bonebind Orator's shape, no target) that puts one counter on
// each Soldier as it resolves. The Veteran is exiled as a cost, so it is not
// there to be counted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ba21f18e-ee65-4c10-92c9-b9fb41510eba",
		Name:         "Valiant Veteran",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Other Soldiers you control get +1/+1.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID &&
					target.Controller == source.Controller &&
					target.IsCreature() && target.HasSubtype(source.TextSubtype("Soldier")) // a word of the source's text (CR 612.2, #3984)
			},
			Apply: effects.PumpBy(1, 1),
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}{W}, Exile this card from your graveyard: Put a +1/+1 counter on each Soldier you control.",
			Cost:    effects.Plus(effects.ManaCost("{3}{W}{W}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect:  effects.DmuB04PutCounterOnEachSoldier,
		}},
	})
}
