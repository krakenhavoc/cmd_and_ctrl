package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fanged Flames — Sorcery {1}{R}:
//
//	Devoid (This card has no color.)
//	Fanged Flames deals 4 damage to target creature or planeswalker. If that creature or planeswalker would die this turn, exile it instead.
//
// The replacement is the spell's own effect, so it is registered on a legal target whether or not the damage is dealt (ADR 0108 §1).
//
// Devoid is declared in PrintedKeywords, and the engine reads it as CR
// 702.114a's colour-defining ability (#2152), so the card is colourless
// in every zone.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "46515251-2172-4b0a-81ac-4c0120b73360",
		Name:            "Fanged Flames",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Targets:         TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve:       damageFirstTargetExileIfItDies(fixedAmount(4)),
	})
}
