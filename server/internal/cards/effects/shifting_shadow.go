package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shifting Shadow — Enchantment — Aura {2}{R} (Reprint, #5908):
//
//	"Enchant creature
//	 Enchanted creature has haste and "At the beginning of your upkeep,
//	 destroy this creature. Reveal cards from the top of your library
//	 until you reveal a creature card. Put that card onto the battlefield
//	 and attach Shifting Shadow to it, then put all other cards revealed
//	 this way on the bottom of your library in a random order.""
//
// Haste is a layer-6 grant on the host (GrantToAttached). The upkeep ability
// is a bundle the Aura grants (Bear Umbra's shape, ADR 0093): its source is
// the enchanted creature, so "your" is that creature's controller (not the
// Aura's), the destroyed creature is the source and the revealed library is
// that controller's. The Aura is named by the item's grantor
// (StackItem.GrantedBy), stamped when the trigger went on the stack.
//
// Three steps, each waiting on the one before (CR 608.2c): the destruction
// (a regeneration shield or indestructible keeps the creature, and the rest
// still happens), the reveal and the put, then, as the put's tail, the Aura
// moves onto the creature that arrived and the rest go to the bottom in a
// random order. The move is an attach by an effect (CR 303.4f, CR 701.3b): an
// Aura that is no longer on the battlefield, or that could not enchant the new
// creature (protection, say), stays where it is. A
// library with no creature card reveals itself entirely and puts it all on the
// bottom.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "564b1253-2e1f-46c5-b5ae-3932127ca614",
		Name:         "Shifting Shadow",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []AbilityGrant{{
			Key: shiftingShadowUpkeep,
			Triggered: []game.TriggeredAbility{
				AtYourUpkeep("Shifting Shadow — destroy this creature; reveal until a creature card, put it onto the battlefield and attach Shifting Shadow to it",
					Do(
						Instruction(func(ctx *Context) error {
							return DestroyTarget{Target: ctx.Source()}.Apply(ctx)
						}),
						Instruction(shiftingShadowReveal),
					)),
			},
			Text: "At the beginning of your upkeep, destroy this creature. Reveal cards from the top of your library until you reveal a creature card. Put that card onto the battlefield and attach Shifting Shadow to it, then put all other cards revealed this way on the bottom of your library in a random order.",
		}},
		Static: []game.StaticAbility{
			GrantToAttached("haste"),
			GrantAbilitiesToAttached(shiftingShadowUpkeep),
		},
	})
}

const shiftingShadowUpkeep = "shifting-shadow/upkeep"

// shiftingShadowReveal is the second and third sentences: reveal until a
// creature card, put it onto the battlefield, then (its tail) attach the
// grantor to it and bottom the rest.
func shiftingShadowReveal(ctx *Context) error {
	player := ctx.Controller()
	run, hit := revealUntil(ctx, player, game.Card.IsCreature, "Shifting Shadow — reveal until a creature card")
	var carry []uuid.UUID
	if g := ctx.Item.GrantedBy; g != uuid.Nil {
		carry = []uuid.UUID{g}
	}
	return PutFromLibraryOntoBattlefield{
		Player:    player,
		Cards:     run,
		Match:     func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
		All:       true,
		ThenTail:  shiftingShadowAttachThenBottom,
		ThenCarry: carry,
	}.Apply(ctx)
}

// shiftingShadowAttachThenBottom is "attach Shifting Shadow to it, then put
// all other cards revealed this way on the bottom of your library in a random
// order", keyed. The carried ID is the Aura. An on-disk key.
var shiftingShadowAttachThenBottom = PutFromLibraryTailCarrying("shifting-shadow/attach-then-bottom", 1,
	func(ctx *Context, res PutFromLibraryResult) error {
		if len(res.Carry) > 0 && len(res.Entered) > 0 {
			aura, creature := res.Carry[0], res.Entered[0]
			if ctx.Game.AttachLegalForEffect(aura, creature) {
				if err := ctx.Game.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: creature}); err != nil {
					return err
				}
			}
		}
		return PutRestOnBottomInRandomOrder(ctx.Game, res)
	})
