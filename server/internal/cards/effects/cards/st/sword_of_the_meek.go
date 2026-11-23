package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sword of the Meek — Artifact — Equipment {2}:
//
//	"Equipped creature gets +1/+2.
//	 Whenever a 1/1 creature you control enters, you may return this
//	 card from your graveyard to the battlefield, then attach it to that
//	 creature.
//	 Equip {2}"
//
// The trigger works from the graveyard (Gate Colossus's shape) and reads
// the entering creature's current power and toughness (CR 613). The
// return carries the creature's ID on its keyed tail; the attach is the
// tail's, so a replacement that pauses the entry is answered first. The
// attach is of the returned Equipment itself — a new object — to a
// creature that must still be on the battlefield (CR 701.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "215c287d-56a5-46da-b49e-8524b6d320a4",
		Name:         "Sword of the Meek",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 2)},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.InGraveyard(effects.On(game.EventETB, swordOfTheMeekSeesAOneOne,
				"Sword of the Meek — return it from your graveyard to the battlefield, then attach it to that creature",
				swordOfTheMeekReturn)),
				"Sword of the Meek — return it to the battlefield and attach it to that creature?"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

func swordOfTheMeekSeesAOneOne(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
	if !ok || !c.IsCreature() {
		return false
	}
	_, power, toughness := c.PowerToughness()
	return power == 1 && toughness == 1
}

func swordOfTheMeekReturn(g *game.Game, item *game.StackItem) error {
	if effects.SourceIsNewObject(g, item) {
		return nil
	}
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	tail := effects.NewContext(g, item).ItemTail(swordOfTheMeekAttach)
	tail.Carry = []uuid.UUID{item.Trigger.Event.CardID}
	return g.ReturnFromGraveyardToBattlefieldKeyedForEffect(item.SourceCardID, item.Controller, false, tail)
}

// swordOfTheMeekAttach is the return's tail: attach the Sword to the
// creature that entered. An on-disk key.
var swordOfTheMeekAttach = effects.EnteredTail("sword-of-the-meek/attach", func(ctx *effects.Context, entered uuid.UUID, o game.TailOutcome) error {
	g := ctx.Game
	if entered == uuid.Nil || len(o.Tail.Carry) == 0 || !effects.OnBattlefield(g, entered) || !effects.OnBattlefield(g, o.Tail.Carry[0]) {
		return nil
	}
	return g.AttachForEffect(entered, game.TargetRef{Kind: game.TargetCard, ID: o.Tail.Carry[0]})
})
