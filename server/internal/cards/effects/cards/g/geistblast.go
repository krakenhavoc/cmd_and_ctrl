package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Geistblast — Instant {2}{R} (Shadows over Innistrad, #5543):
//
//	"Geistblast deals 2 damage to any target.
//	 {2}{U}, Exile this card from your graveyard: Copy target instant or
//	 sorcery spell you control. You may choose new targets for the copy."
//
// The spell is a plain two-damage bolt. The graveyard ability
// (ActivatedAbilityShape.Zones) exiles the card as a cost and is Rootha's
// copy: it targets an instant or sorcery spell you control and lets you
// choose new targets for the copy. Both halves are on the one card, and the
// ability works only from the graveyard (CR 113.6k).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10891656-070f-49f4-80e0-78061b773f17",
		Name:         "Geistblast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve:    effects.DamageToFirstTargetStep(2),
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, Exile this card from your graveyard: Copy target instant or sorcery spell you control. You may choose new targets for the copy.",
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Targets: effects.TargetInstantOrSorcerySpell("target instant or sorcery spell you control", effects.YouControl()),
			Effect:  effects.CopyTargetedSpell,
		}},
	})
}
