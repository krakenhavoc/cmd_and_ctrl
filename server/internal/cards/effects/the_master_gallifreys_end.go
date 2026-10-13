package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Master, Gallifrey's End — Legendary Creature — Time Lord Rogue
// {2}{B}{R}, 4/3:
//
//	"Make Them Pay — Whenever a nontoken artifact creature you control
//	 dies, you may exile it. If you do, choose an opponent with the most
//	 life among your opponents. That player faces a villainous choice —
//	 They lose 4 life, or you create a token that's a copy of that
//	 card."
//
// A villainous choice (CR 701.55, #3143) faced by the opponent with the
// most life. One such opponent is the answer; a tie is the controller's
// to break, asked as a "choose a player" over exactly the tied seats
// (OpponentsWithTheMostLife). The life totals are read when the exile
// has happened, as the sentence orders it.
//
// "You may exile it" is asked when the trigger fires, the engine's
// standard optional-trigger shape, and the exile itself is a clause of
// the resolution: "it" is the object that died (CR 400.7), so a card
// already moved out of the graveyard in response is not exiled and
// nothing else happens ("if you do"). "Nontoken" and "artifact" are the
// creature as it last existed (CR 603.10a), "you control" the player who
// controlled it as it died.
//
// The second option copies the exiled card: the token has the card's
// copiable values (CR 707.2), not the counters or damage the creature
// had. The Master herself is not an artifact creature, so her own death
// is never the trigger.
//
// No simplifications.
var masterChoice = NewVillainousChoice("the-master-gallifreys-end",
	"The Master — lose 4 life, or the controller creates a token copy of the exiled card?",
	VillainousOption{
		Label:    "You lose 4 life",
		LifeCost: 4,
		Body: func(ctx *Context, f Facing) error {
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), f.Player, -4)
		},
	},
	VillainousOption{
		Label: "The controller of The Master creates a token that's a copy of the exiled card",
		Harm:  2,
		Body: func(ctx *Context, f Facing) error {
			if len(f.Carry) == 0 {
				return nil
			}
			return CreateTokenCopy{Controller: ctx.Controller(), Copy: f.Carry[0], N: 1}.Apply(ctx)
		},
	})

const masterLabel = "The Master — exile it, then the opponent with the most life faces a villainous choice"

func init() {
	Register(Spec{
		OracleID:     "e5e2e4ef-1d78-41eb-b634-1a90d4cb7125",
		Name:         "The Master, Gallifrey's End",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventLTB},
				Key:     masterLabel,
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					dead, ok := diedCreature(ev, g)
					if !ok || IsToken(dead) || !leftAsType(ev, dead, "artifact") {
						return false
					}
					return leftUnderControlOf(ev, dead) == source.Controller
				},
				Build:  eventCardTriggerItem(masterLabel),
				Effect: masterMakeThemPay,
			}, "The Master — exile the artifact creature that died?"),
		},
	})
}

// masterMakeThemPay exiles the dead card if it is still the same object
// in a graveyard and, only if that happened, picks the opponent and
// asks them.
func masterMakeThemPay(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	deadID, ok := deadCardInGraveyard(g, item)
	if !ok {
		return nil
	}
	if err := (ExileTarget{Target: deadID}).Apply(ctx); err != nil {
		return err
	}
	if z := g.FindCardZoneForEffect(deadID); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	face := func(ctx *Context, opponent uuid.UUID) error {
		return FaceVillainousChoice{
			Players: []uuid.UUID{opponent},
			Choice:  masterChoice,
			Carry:   []uuid.UUID{deadID},
		}.Apply(ctx)
	}
	most := OpponentsWithTheMostLife(ctx)
	switch len(most) {
	case 0:
		return nil
	case 1:
		return face(ctx, most[0])
	}
	return ChoosePlayer{
		Only:     most,
		Question: "The Master — choose an opponent with the most life",
		Then: func(ctx *Context) error {
			if opp := ctx.ChosenPlayer(); opp != uuid.Nil {
				return face(ctx, opp)
			}
			return nil
		},
	}.Apply(ctx)
}
