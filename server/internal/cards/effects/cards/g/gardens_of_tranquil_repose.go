package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gardens of Tranquil Repose — Plane — Necros (Doctor Who, #4870):
//
//	"Suspended Animation — Whenever a creature dies, exile it. Its
//	 controller scries 1.
//	 Whenever chaos ensues, create X 3/3 black Dalek artifact creature
//	 tokens with menace, where X is one plus the number of cards exiled
//	 with Gardens of Tranquil Repose."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The first ability triggers on every creature's death. "It" is the card
// that died, still in the graveyard as that object (CR 400.7); a token has
// ceased to exist and a card that moved again is a new object, so neither
// is exiled. The exile is linked to the plane OBJECT (CR 607.2a, Card.
// ExiledWith), and "its controller" is the controller the creature had as
// it last existed on the battlefield (CR 608.2h), who scries 1 whether or
// not the card was exiled. Turned face down by a planeswalk the plane is a
// new object (CR 901.7a), so the chaos ability (CR 311.7) counts only the
// cards exiled with the face-up plane it came from.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d45692d5-19e8-45ec-92bf-b79cb3919acd",
		Name:         "Gardens of Tranquil Repose",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, anyCreatureDied, "Suspended Animation — Gardens of Tranquil Repose exiles the creature; its controller scries 1",
				gardensOfTranquilReposeExile),
			effects.WheneverChaosEnsues("Gardens of Tranquil Repose — create X 3/3 Daleks with menace (X = one plus the cards exiled with it)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := 1
					if ref, ok := ctx.SourceRef(); ok {
						x += len(g.ExiledWithForEffect(game.PermissionCardRef(ref)))
					}
					return effects.CreateToken{Template: effects.TokenCard("3/3 black Dalek artifact with menace"), N: x}.Apply(ctx)
				}),
		},
	})
}

// anyCreatureDied is "whenever a creature dies": any creature, anyone's.
func anyCreatureDied(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	_, ok := effects.DiedCreature(ev, g)
	return ok
}

func gardensOfTranquilReposeExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	obj := ctx.Trigger().Object
	if obj == nil {
		return nil
	}
	controller := uuid.Nil
	if info, ok := g.PermanentForEffect(game.ObjectRef{ID: obj.ID, Epoch: obj.Epoch}); ok {
		controller = info.Controller
	}
	then := ctx.ItemTail(gardensOfTranquilReposeScry)
	then.Carry = []uuid.UUID{controller}
	if obj.Token || !effects.DiedCardStillInGraveyard(ctx, obj.ID) {
		return then.Run(g, nil)
	}
	return effects.ExileTarget{Target: obj.ID, Linked: true, ThenTail: then}.Apply(ctx)
}

// gardensOfTranquilReposeScry is "Its controller scries 1", once the exile
// has settled. An on-disk key: never renamed, never reused.
var gardensOfTranquilReposeScry = effects.CardTail("gardens-of-tranquil-repose/its-controller-scries", func(ctx *effects.Context, o game.TailOutcome) error {
	controller := effects.CarriedID(o.Tail.Carry, 0)
	if controller == uuid.Nil {
		return nil
	}
	return effects.Scry{Player: controller, N: 1}.Apply(ctx)
})
