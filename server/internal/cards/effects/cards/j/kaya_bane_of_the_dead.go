package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kaya, Bane of the Dead — Legendary Planeswalker — Kaya
// {3}{W/B}{W/B}{W/B}, loyalty 7:
//
//	"Your opponents and permanents your opponents control with hexproof
//	 can be the targets of spells and abilities you control as though
//	 they didn't have hexproof.
//	 −3: Exile target creature."
//
// A second proof card for #1560's hexproof waiver, and the one that
// needs its PLAYER half (CR 702.11d): an opponent's Leyline of
// Sanctity is waived for Kaya's controller exactly as an opponent's
// Swiftfoot Boots creature is. "You control" makes both waivers
// BySpellsAndAbilitiesYouControl, so another opponent of the hexproof
// player gets nothing from Kaya. "With hexproof" adds no condition:
// the waiver is only ever asked about something whose hexproof would
// refuse the target.
//
// The −3 is an ordinary loyalty ability and can use the waiver
// itself, which is the point of the card. If Kaya leaves with it on
// the stack, the CR 608.2b re-check makes a hexproof target illegal
// again (the 2024-09-20 ruling), because the re-check reads the same
// live waiver the announce did.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29d58aa4-9a7e-4052-9b0d-fc282f1be40e",
		Name:         "Kaya, Bane of the Dead",
		Completeness: effects.CompletenessFull,
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 7,
		HexproofBypasses: []game.HexproofBypass{
			effects.AsThoughNoHexproof(effects.BySpellsAndAbilitiesYouControl, effects.OpponentControls()),
			effects.PlayersAsThoughNoHexproof(effects.BySpellsAndAbilitiesYouControl, effects.Opponent()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "−3: Exile target creature.",
			Cost:    effects.LoyaltyCost(-3),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.SocB37ExileTheLegalTarget,
		}},
	})
}
