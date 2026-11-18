package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Moonsnare Prototype — Artifact {U} (Neon Dynasty, #5493):
//
//	"{T}, Tap an untapped artifact or creature you control: Add {C}.
//	 Channel — {4}{U}, Discard this card: The owner of target nonland
//	 permanent puts it on their choice of the top or bottom of their
//	 library."
//
// The mana ability is Citanul Stalwart's: the {T} taps the Prototype, so the
// other permanent is another one (TapAnotherUntapped), which is not the {T}
// symbol and does not target. The channel's placement is Endless Detour's:
// the permanent's OWNER chooses top or bottom, and picks the spot, not the
// caster.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7346c266-19d8-4432-acf5-f00aff074ad5",
		Name:         "Moonsnare Prototype",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, TapOthers: TapAnotherUntapped("an untapped artifact or creature you control", Or(Artifact(), Creature())).TapOthers},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "Channel — {4}{U}, Discard this card: The owner of target nonland permanent puts it on their choice of the top or bottom of their library.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{4}{U}"), DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetPermanent("target nonland permanent", Not(Land())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				c, ok := g.LookupCardForEffect(id)
				if !ok || c.Owner == uuid.Nil {
					return nil
				}
				return PutInLibraryInAnyOrder{
					Chooser:   c.Owner,
					Cards:     []uuid.UUID{id},
					Placement: game.LibraryPlaceTopOrBottom,
					Label:     "Moonsnare Prototype — put it on the top or bottom of your library",
				}.Apply(ctx)
			},
		}},
	})
}
