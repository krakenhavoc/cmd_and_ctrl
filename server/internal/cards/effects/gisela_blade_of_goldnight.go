package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gisela, Blade of Goldnight — Legendary Creature — Angel {4}{R}{W}{W}, 5/5
// (Lost Caverns of Ixalan Commander):
//
//	"Flying, first strike
//	 If a source would deal damage to an opponent or a permanent an
//	 opponent controls, that source deals double that damage to that player
//	 or permanent instead.
//	 If a source would deal damage to you or a permanent you control,
//	 prevent half that damage, rounded up."
//
// Two replacement effects (CR 614, CR 615) with disjoint recipients, so no
// damage meets both: the doubling is Angrath's Marauders' (CR 614.1) over an
// opponent or a permanent an opponent controls, and the prevention leaves the
// damage rounded down (half rounded up is prevented, CR 615.1). Any source
// counts, combat damage and burn alike, yours and theirs.
func init() {
	Register(Spec{
		OracleID:        "66f9f325-5e8e-4ebf-b5b3-c6410d80f2c5",
		Name:            "Gisela, Blade of Goldnight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Replacements: []game.ReplacementEffect{
			{
				Watches:  []game.EventKind{game.EventDealDamage},
				RowLabel: "If a source would deal damage to an opponent or a permanent an opponent controls, that source deals double that damage instead.",
				Label:    "Gisela, Blade of Goldnight — double that damage",
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 {
						return false
					}
					return cmdr2B08RecipientIsOpponentOrTheirPermanent(ev.DamageTarget, g, src.Controller)
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.DamageAmount *= 2
					return nil
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
			},
			{
				Watches:    []game.EventKind{game.EventDealDamage},
				Prevention: true, // CR 615.1a — "prevent"
				RowLabel:   "If a source would deal damage to you or a permanent you control, prevent half that damage, rounded up.",
				Label:      "Gisela, Blade of Goldnight — prevent half that damage, rounded up",
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 {
						return false
					}
					return cmdr2B08RecipientIsYouOrYourPermanent(ev.DamageTarget, g, src.Controller)
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.DamageAmount /= 2 // half rounded up is prevented
					return nil
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
			},
		},
	})
}
