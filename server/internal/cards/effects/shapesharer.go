package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shapesharer — Creature — Shapeshifter {1}{U}, 1/1:
//
//	"Changeling (This card is every creature type.)
//	 {2}{U}: Target Shapeshifter becomes a copy of target creature
//	 until your next turn."
//
// Changeling (CR 702.73a) needs no catalog entry of its own — see
// AGENTS.md §7's tribal section — so the whole card is the activated
// ability, and the whole ability is #1723's first seam: BecomeCopy
// had no way to say "until your next turn" (CR 611.2b) before this,
// only "until end of turn" or no stated duration at all.
//
// Two DIFFERENT target clauses (#764): slot 0 is the permanent that
// CHANGES — any Shapeshifter, which (changeling) is every creature on
// the battlefield with a creature type at all, this card's own
// changeling body included — and slot 1 is the creature it copies.
// Each is checked separately at resolution (CR 608.2b): a Shapeshifter
// that left in response takes the ability off the stack for nothing;
// a model that left leaves the ability with nothing to copy, per
// BecomeCopy's own "as much as possible" (CR 608.2b) rule.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "400c9fd8-c307-4b27-af6a-73e100717881",
		Name:         "Shapesharer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{2}{U}: Target Shapeshifter becomes a copy of target creature until your next turn.",
			Cost:  ManaCost("{2}{U}"),
			Targets: Clauses(
				TargetCreature("target Shapeshifter", OfCreatureType("Shapeshifter")),
				TargetCreature("target creature"),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				shifter, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				model, ok := ctx.ClauseTarget(1)
				if !ok {
					return nil
				}
				return BecomeCopy{
					Targets:  []uuid.UUID{shifter.ID},
					Of:       model.ID,
					Duration: CopyUntilYourNextTurn,
					Label:    "Shapesharer — becomes a copy until your next turn",
				}.Apply(ctx)
			},
		}},
	})
}
