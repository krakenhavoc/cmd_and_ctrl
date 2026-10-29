package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dáin Ironfoot — Legendary Creature — Dwarf Warrior {2}{R}, 1/4
// (The Hobbit, #4454):
//
//	"When Dáin enters, create a colorless Equipment artifact token named
//	 Axe with "Equipped creature gets +1/+0" and equip {2}. When you do,
//	 attach it to target creature you control.
//	 Whenever Dáin attacks, each equipped attacking creature gains double
//	 strike until end of turn."
//
// The Axe is Iron Hills Blacksmith's token (AxeEquipmentToken). "When you
// do" is the CR 603.12 reflexive trigger the token's creation raises: it
// goes on the stack with its own "target creature you control", chosen
// there, and the Axe it attaches rides the payload, so a table waiting on
// it is a restore point. A token that was not created (a replacement) raises
// nothing.
//
// The attack trigger's set is fixed as it resolves (CR 611.2c): every
// attacking creature that has an Equipment attached then, Dáin's own
// included, whoever controls it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9924c189-3c82-431a-b570-7805b2a406f0",
		Name:         "Dáin Ironfoot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dáin Ironfoot — create an Axe Equipment token, then attach it to target creature you control", dainIronfootCreateAxe),
			WheneverThisAttacks("Dáin Ironfoot — each equipped attacking creature gains double strike",
				Do(GrantKeywordUntilEOT{
					Match: func(g *game.Game, caster uuid.UUID, c game.Card) bool {
						return AttackingCreature()(g, caster, c) && isEquippedCreature(g, c)
					},
					Keywords: []string{"double strike"},
					Label:    "Dáin Ironfoot — double strike",
				})),
		},
	})
}

// dainIronfootCreateAxe is the enters trigger's first sentence: the Axe,
// created as a described token (the card's text names it), then its keyed
// tail.
func dainIronfootCreateAxe(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups: []game.TokenGroup{{
			Template: ctx.TextToken(AxeEquipmentToken()),
			Count:    1,
		}},
	}, ctx.ItemTail(dainIronfootAxeCreated))
}

// dainIronfootAxeCreated is "when you do": the reflexive trigger, raised
// only for an Axe that really arrived.
var dainIronfootAxeCreated = CardTail("dain-ironfoot/axe-created", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return ReflexiveTrigger{
		Label: "Dáin Ironfoot — attach the Axe to target creature you control",
		Cards: o.IDs[:1],
		Body:  dainIronfootAttachBody,
	}.Apply(ctx)
})

// dainIronfootAttachBody is the reflexive trigger's registration: the Axe is
// the payload, the creature the trigger's own target.
var dainIronfootAttachBody = game.ReflexiveBody("dain-ironfoot/attach-axe", simpleBody(dainIronfootAttachAxe),
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature you control", YouControl()) }))

// dainIronfootAttachAxe attaches the Axe to the target. Either having left
// is the ability doing as much as it can, which is nothing (CR 608.2b).
func dainIronfootAttachAxe(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	axes := ctx.PayloadCards()
	targets := ctx.LegalTargets()
	if len(axes) == 0 || len(targets) == 0 {
		return nil
	}
	return g.AttachForEffect(axes[0], targets[0])
}
