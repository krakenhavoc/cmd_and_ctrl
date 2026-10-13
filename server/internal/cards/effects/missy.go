package effects

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Missy — Legendary Creature — Time Lord Rogue {3}{U}{B}{R}, 4/5:
//
//	"Whenever another nonartifact creature dies, return it to the
//	 battlefield under your control face down and tapped. It's a 2/2
//	 Cyberman artifact creature.
//	 At the beginning of your end step, each opponent faces a villainous
//	 choice — Each artifact creature you control deals 1 damage to that
//	 opponent, or you draw a card and chaos ensues."
//
// The first ability is the graveyard door to a LISTED face-down body
// (game.ReturnFromGraveyardFaceDownForEffect, #1270): the reanimation
// pipeline with the face-down state riding the entry event, so no ETB
// trigger and no "as enters" hook runs (CR 708.2a) and the controller is
// the only player who may look (CR 708.5). The body is CybermanBody() —
// a nameless 2/2 Cyberman artifact creature, whatever the card is
// underneath. The kind is `turned`: an effect put it here, so only the
// card's own morph or disguise can turn it up (CR 708.7). It enters
// tapped (a clause of the entry, so nothing sees it untapped) under the
// controller of Missy, not its owner. "Another" excludes Missy, and
// "nonartifact" is the creature as it last existed (CR 603.10a); a token
// (CR 111.7) and a card that left the graveyard in response are the
// CR 400.7 "nothing happens". The trigger is mandatory.
//
// The second is a villainous choice (CR 701.55, #3143) asked of each
// opponent in APNAP order, one at a time. Option A is one simultaneous
// damage instance from each artifact creature the controller has as the
// answer comes in, each its own source (so deathtouch, lifelink and
// damage-source prevention see each), face-down Cybermen included.
// Option B draws a card.
//
// "Chaos ensues" is Planechase's planar die (CR 901): it does
// something only with a planar deck and a planar die in the game. Neither
// exists in a Commander game on this engine, and no card of the catalog
// cares, so the clause has nothing to do and nothing is left out.
//
// No simplifications.
var missyChoice = NewVillainousChoice("missy",
	"Missy — artifact creatures deal 1 damage each to you, or the controller draws a card?",
	VillainousOption{
		Label: "Each artifact creature the controller of Missy controls deals 1 damage to you",
		Harm:  1,
		Body: func(ctx *Context, f Facing) error {
			artifacts := MatchingBattlefield(ctx, And(Creature(), Artifact(), YouControl()))
			if len(artifacts) == 0 {
				return nil
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				for _, c := range artifacts {
					if err := (DealDamage{Source: c.InstanceID, Target: f.Player, Amount: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	},
	VillainousOption{
		Label: "The controller of Missy draws a card and chaos ensues",
		Harm:  1,
		Body: func(ctx *Context, _ Facing) error {
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})

const missyReturnLabel = "Missy — return it face down and tapped as a 2/2 Cyberman artifact creature"

func init() {
	Register(Spec{
		OracleID:     "905610cd-1461-4f5a-b9c8-19b193b9d513",
		Name:         "Missy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventLTB},
				Key:     missyReturnLabel,
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.CardID == source.InstanceID {
						return false
					}
					dead, ok := diedCreature(ev, g)
					return ok && !IsToken(dead) && !leftAsType(ev, dead, "artifact")
				},
				Build:  eventCardTriggerItem(missyReturnLabel),
				Effect: missyReturnIt,
			},
			AtYourEndStep("Missy — each opponent faces a villainous choice",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return FaceVillainousChoice{Players: ctx.Opponents(), Choice: missyChoice}.Apply(ctx)
				}),
		},
	})
}

// missyReturnIt is the first trigger's resolution: the dead card, if it
// is still that object in a graveyard, returns face down and tapped under
// the controller's control.
func missyReturnIt(g *game.Game, item *game.StackItem) error {
	id, ok := deadCardInGraveyard(g, item)
	if !ok {
		return nil
	}
	_, err := g.ReturnFromGraveyardFaceDownForEffect(id, item.Controller, true, CybermanBody())
	if errors.Is(err, game.ErrCardNotFound) {
		return nil
	}
	return err
}
