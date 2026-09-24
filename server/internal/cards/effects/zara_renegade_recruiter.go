package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zara, Renegade Recruiter — Legendary Creature — Human Pirate
// {3}{U}{R}, 4/3 (Edea steal-and-sac deck, #1565):
//
//	"Flying
//	 Whenever Zara attacks, look at defending player's hand. You may
//	 put a creature card from it onto the battlefield under your
//	 control tapped and attacking that player or a planeswalker they
//	 control. Return that creature to its owner's hand at the
//	 beginning of the next end step."
//
// Three existing pieces, in printed order:
//
//   - **Look at the hand.** Only Zara's controller becomes a knower of
//     the defending player's hand, as Gitaxian Probe's look does. It is
//     a look, not a reveal: the rest of the table learns nothing.
//   - **The pick** is a resolution-time choose-cards prompt over the
//     creature cards in that hand, floor zero ("you MAY put"). It does
//     not target, so hexproof and shroud do not matter to a card in a
//     hand anyway. When the defending player controls a planeswalker,
//     Zara's controller then chooses what the creature attacks: that
//     player, or one of their planeswalkers.
//   - **The entry** is ninjutsu's CR 506.3c door (#1227):
//     PutFromHandOntoBattlefield with Controller, Tapped and Attacking.
//     The creature was never declared as an attacker, so it fires no
//     "whenever this attacks" trigger and needs no haste. It enters
//     under Zara's controller's control and its owner still owns it,
//     which is what Edea and Don Andres read.
//
// The return is a CR 603.7 delayed trigger at the next end step.
// "That creature" names the object that entered (CR 400.7): the
// effect records its object epoch as it lands and returns it only if
// it is still that object on the battlefield. One that died, was
// flickered or was bounced and recast stays where it is.
//
// The defending player is read as Zara's attack is declared: the
// player she attacks, or the controller of the planeswalker or the
// protector of the battle she attacks (b17DefendingPlayer).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cd2720c2-522c-4fdc-9cef-9c5ce250fe7b",
		Name:            "Zara, Renegade Recruiter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventAttack},
			AppliesTo: ThisAttacked,
			Key:       zaraLabel,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				defender := b17DefendingPlayer(g, ev)
				return game.NewTriggeredItem(source, zaraLabel,
					func(g *game.Game, item *game.StackItem) error {
						return zaraLookAndPut(g, item, defender)
					})
			},
		}},
	})
}

const zaraLabel = "Zara, Renegade Recruiter — look at defending player's hand"

// zaraLookAndPut is the attack trigger's resolution.
func zaraLookAndPut(g *game.Game, item *game.StackItem, defender uuid.UUID) error {
	p := g.PlayerByIDForEffect(defender)
	if p == nil || p.Eliminated || p.Hand == nil {
		return nil
	}
	me := item.Controller
	var creatures []uuid.UUID
	for i := range p.Hand.Cards {
		p.Hand.Cards[i].AddKnower(me)
		if p.Hand.Cards[i].IsCreature() {
			creatures = append(creatures, p.Hand.Cards[i].InstanceID)
		}
	}
	if len(creatures) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    me,
		FromPlayer: defender,
		Source:     item.SourceCardID,
		Question:   "Zara, Renegade Recruiter — you may put a creature card from their hand onto the battlefield tapped and attacking",
		Cards:      creatures,
		Min:        0,
		Max:        1,
		Zone:       game.ZoneHand,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			return zaraChooseAttackTarget(g, item, defender, picked[0])
		},
	})
	return nil
}

// zaraChooseAttackTarget asks what the creature attacks when the
// defending player controls a planeswalker, and puts it straight in
// attacking that player otherwise.
func zaraChooseAttackTarget(g *game.Game, item *game.StackItem, defender, card uuid.UUID) error {
	targets := []uuid.UUID{defender}
	options := []game.ChoiceOption{{Label: "Attack that player"}}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == defender && c.IsPlaneswalker() {
			targets = append(targets, c.InstanceID)
			options = append(options, game.ChoiceOption{Label: "Attack " + c.Name, Cards: []uuid.UUID{c.InstanceID}})
		}
	}
	if len(targets) == 1 {
		return zaraPutAttacking(g, item, card, defender)
	}
	return PickOption{
		Question: "Zara, Renegade Recruiter — what does it attack?",
		Options:  options,
		Then: func(ctx *Context, index int) error {
			if index < 0 || index >= len(targets) {
				index = 0
			}
			return zaraPutAttacking(ctx.Game, item, card, targets[index])
		},
	}.Apply(NewContext(g, item))
}

// zaraPutAttacking puts the card onto the battlefield under Zara's
// controller's control, tapped and attacking `attacking`, and
// schedules its return to its owner's hand at the next end step.
func zaraPutAttacking(g *game.Game, item *game.StackItem, card, attacking uuid.UUID) error {
	if z := g.FindCardZoneForEffect(card); z == nil || z.Kind != game.ZoneHand {
		return nil
	}
	return g.PutFromHandOntoBattlefieldThenForEffect(card, game.HandEntryOptions{
		Controller: item.Controller,
		Tapped:     true,
		Attacking:  attacking,
	}, func(g *game.Game, entered uuid.UUID) error {
		if entered == uuid.Nil {
			return nil
		}
		c, ok := g.LookupCardForEffect(entered)
		if !ok {
			return nil
		}
		return ScheduleDelayedTrigger{
			Label:  "Zara, Renegade Recruiter — return that creature to its owner's hand",
			Cards:  []uuid.UUID{entered},
			Body:   zaraReturnToHandBody,
			Params: game.EffectParams{Object: game.ObjectRef{ID: entered, Epoch: c.ObjectEpoch}},
		}.Apply(NewContext(g, item))
	})
}

// zaraReturnToHand is the delayed trigger's body (zaraReturnToHandBody,
// delayed_bodies.go): bounce the creature if it is still the object
// that entered, which p.Object names.
func zaraReturnToHand(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return BounceToHand{Target: id}.Apply(NewContext(g, item))
}
