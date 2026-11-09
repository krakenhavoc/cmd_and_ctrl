package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Emet-Selch, Unsundered // Hades, Sorcerer of Eld (Final Fantasy, #4462) —
// a transforming legendary creature.
//
// Front face, Legendary Creature — Elder Wizard {1}{U}{B}, 2/4:
//
//	"Vigilance
//	 Whenever Emet-Selch enters or attacks, draw a card, then discard a
//	 card.
//	 At the beginning of your upkeep, if there are fourteen or more cards in
//	 your graveyard, you may transform Emet-Selch."
//
// Back face, Hades, Sorcerer of Eld — Legendary Creature — Avatar, 6/6
// (hades_sorcerer_of_eld.go).
//
// The loot is one ability with two trigger conditions (WhenThisEntersOrAttacks)
// whose discard waits on the draw. The upkeep trigger's condition is an
// intervening "if" (CR 603.4): checked as the phase begins and again as the
// ability resolves, before "you may" is asked. The transform is the
// permanent's own, not an exile-and-return: it stays the same object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        emetSelchUnsunderedOracleID,
		Name:            "Emet-Selch, Unsundered",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Emet-Selch, Unsundered — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return lootOne(g, item, 1) }),
			InterveningIf(CondBoard{Name: "there are fourteen or more cards in your graveyard", Test: func(ctx *Context) bool {
				p := ctx.PlayerByID(ctx.Controller())
				return p != nil && p.Graveyard != nil && p.Graveyard.Size() >= 14
			}}, AtYourUpkeep("Emet-Selch, Unsundered — you may transform Emet-Selch",
				func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) || !g.CanTransformForEffect(item.SourceCardID) {
						return nil
					}
					return MayChoice{
						Question: "Emet-Selch, Unsundered — transform Emet-Selch?",
						Then:     emetSelchUnsunderedTransform,
						Carry:    []uuid.UUID{item.SourceCardID},
					}.Apply(NewContext(g, item))
				})),
		},
	})
}

// emetSelchUnsunderedOracleID is shared with the back face, which registers
// under it plus "#1" (game.CatalogKey).
const emetSelchUnsunderedOracleID = "3b102ccc-7629-457c-aba6-e9b00fd50c85"

// emetSelchUnsunderedTransform is the "yes": transform Emet-Selch (the
// carried ID). Keyed (#4291).
var emetSelchUnsunderedTransform = MayYes("confirm/emet-selch-unsundered-transform", func(ctx *Context, r game.ConfirmAnswer) error {
	return Transform{Target: carryAt(r, 0)}.Apply(ctx)
})
