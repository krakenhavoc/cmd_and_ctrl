package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Selfless Safewright — Creature — Elf Warrior {3}{G}{G}, 4/2 (Lorwyn
// Eclipsed, #4458):
//
//	"Flash
//	 Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 When this creature enters, choose a creature type. Other permanents
//	 you control of that type gain hexproof and indestructible until end
//	 of turn."
//
// The type is chosen as the trigger resolves (#2382, Distant Melody's
// prompt), not as the creature enters, so an opponent sees the trigger
// on the stack before the choice. The grant covers PERMANENTS of the
// type, not only creatures: a Kindred Elf enchantment you control gets
// it, and so does a changeling, which has every creature type. The set
// is fixed when the type is named (CR 611.2c), so a permanent that
// arrives later in the turn isn't covered. "Other" is the Safewright
// object itself (CR 109.1): it never protects itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ce19b337-9266-471d-bec3-49eb7f0e62ac",
		Name:            "Selfless Safewright",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		TapCost:         Convoke(),
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Selfless Safewright — choose a creature type; other permanents you control of that type gain hexproof and indestructible until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ChooseCreatureTypeTail(g, item.Controller, item.SourceCardID,
						"Selfless Safewright — choose a creature type", ctx.ItemTail(selflessSafewrightTypeChosen))
					return nil
				}),
		},
	})
}

// selflessSafewrightTypeChosen grants hexproof and indestructible until
// end of turn to the other permanents of the chosen type the trigger's
// controller controls. An on-disk key.
var selflessSafewrightTypeChosen = creatureTypeChosen("then/selfless-safewright-type-chosen", func(ctx *Context, t string, _ game.TailOutcome) error {
	source := ctx.Source()
	return GrantKeywordUntilEOT{
		Match: func(_ *game.Game, caster uuid.UUID, c game.Card) bool {
			return c.InstanceID != source && c.Controller == caster && c.HasSubtype(t)
		},
		Keywords: []string{"hexproof", "indestructible"},
		Label:    "Selfless Safewright — hexproof and indestructible",
	}.Apply(ctx)
})
