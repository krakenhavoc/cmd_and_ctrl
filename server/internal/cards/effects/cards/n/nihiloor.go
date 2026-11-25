package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nihiloor — Legendary Creature — Horror {2}{W}{U}{B}, 3/5 (Adventures in
// the Forgotten Realms):
//
//	"When Nihiloor enters, for each opponent, tap up to one untapped creature
//	 you control. When you do, gain control of target creature that player
//	 controls with power less than or equal to the tapped creature's power for
//	 as long as you control Nihiloor.
//	 Whenever you attack with a creature an opponent owns, you gain 2 life and
//	 that player loses 2 life."
//
// The enters trigger walks the opponents in turn order after its controller;
// for each, the controller picks up to one untapped creature they control
// (the picks chain, so a creature tapped for one opponent is not offered for
// the next). Each creature tapped creates its own CR 603.12 reflexive trigger,
// which goes on the stack once the ability has finished resolving; its target
// is a creature that opponent controls whose power is at most the tapped
// creature's, read live while the tapped creature is on the battlefield and
// as it last was once it has left (CR 608.2h), and chosen as the trigger goes
// on the stack. Control lasts for as long as the controller controls Nihiloor
// (CR 611.2b). The second line triggers once per attacking creature an
// opponent owns, whoever controls it now; "that player" is its owner. No
// simplification.
func init() {
	nihiloorPick = effects.AskedPick("choose-cards/nihiloor-tap-for-opponent", nihiloorPicked)
	effects.Register(effects.Spec{
		OracleID:     "148271de-3351-41a1-bcf5-3ecaadc51540",
		Name:         "Nihiloor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Nihiloor — for each opponent, tap up to one untapped creature you control; when you do, gain control of target creature that player controls with power less than or equal to the tapped creature's power for as long as you control Nihiloor",
				func(g *game.Game, item *game.StackItem) error { return nihiloorAsk(g, item, 1) }),
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return nihiloorOpponentOwned(g, ev, source.Controller) != uuid.Nil
				},
				Key: nihiloorDrainLabel,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, nihiloorDrainLabel)
					item.Params.Player = nihiloorOpponentOwned(g, ev, source.Controller)
					return item
				},
				Effect: nihiloorDrain,
			},
		},
	})
}

const nihiloorDrainLabel = "Nihiloor — you gain 2 life and that player loses 2 life"

// nihiloorOpponentOwned is the owner of the attacking creature when `you`
// attacked with a creature an opponent owns, else uuid.Nil.
func nihiloorOpponentOwned(g *game.Game, ev game.Event, you uuid.UUID) uuid.UUID {
	if !effects.AttackDeclaredByYou(ev, you) {
		return uuid.Nil
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.Owner == uuid.Nil || c.Owner == you || g.PlayerByIDForEffect(c.Owner) == nil {
		return uuid.Nil
	}
	return c.Owner
}

func nihiloorDrain(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.GainLife{Amount: 2}).Apply(ctx); err != nil {
		return err
	}
	victim := item.Params.Player
	if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
		return nil
	}
	return g.ChangePlayerLifeForEffect(ctx.Source(), victim, -2)
}

// nihiloorAsk asks the controller for the creature to tap for the nth
// opponent after them in turn order, skipping seats that are empty or out of
// the game. With no untapped creature left to offer, no later opponent can be
// asked either.
func nihiloorAsk(g *game.Game, item *game.StackItem, n int) error {
	you := item.Controller
	at := -1
	for i, p := range g.Seats {
		if p != nil && p.ID == you {
			at = i
			break
		}
	}
	if at < 0 {
		return nil
	}
	var untapped []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.IsCreature() && !c.Tapped {
			untapped = append(untapped, c.InstanceID)
		}
	}
	if len(untapped) == 0 {
		return nil
	}
	for ; n < len(g.Seats); n++ {
		opp := g.Seats[g.SeatInTurnOrderForEffect(at, n)]
		if opp == nil || opp.ID == you || opp.Eliminated {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:   you,
			Source:    item.SourceCardID,
			Question:  "Nihiloor — tap up to one untapped creature you control (for " + opp.Name + ")",
			Cards:     untapped,
			Min:       0,
			Max:       1,
			LiveZone:  game.ZoneBattlefield,
			ThenKey:   nihiloorPick,
			Carry:     []uuid.UUID{opp.ID},
			CarryNums: []int{n},
			Asked:     game.NewAskedItem(item),
		})
		return nil
	}
	return nil
}

// nihiloorPick is assigned in init: the pick asks the next opponent's pick,
// which would be an initialization cycle as a var initializer.
var nihiloorPick game.ChooseCardsThen

func nihiloorPicked(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Carry) != 1 || len(r.Nums) != 1 {
		return nil
	}
	opp, n := r.Carry[0], r.Nums[0]
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsCreature() || c.Controller != ctx.Controller() {
			continue
		}
		if err := (effects.TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
		if err := (effects.ReflexiveTrigger{
			Label: "Nihiloor — gain control of target creature that player controls with power less than or equal to the tapped creature's power",
			Body:  nihiloorStealBody,
			Params: game.EffectParams{
				Player: opp,
				Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch},
				Amount: c.CurrentPower(),
			},
		}).Apply(ctx); err != nil {
			return err
		}
		break
	}
	return nihiloorAsk(ctx.Game, ctx.Item, n+1)
}

// nihiloorPowerAtMostTapped is the predicate of the reflexive clause: a
// creature `opp` controls whose power is at most the tapped creature's.
func nihiloorPowerAtMostTapped(opp uuid.UUID, tapped game.ObjectRef, atTap int) effects.CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		if c.Controller != opp {
			return false
		}
		limit := atTap
		if t, ok := g.LookupCardForEffect(tapped.ID); ok && t.ObjectEpoch == tapped.Epoch && g.Battlefield.Contains(tapped.ID) {
			limit = t.CurrentPower()
		}
		return c.CurrentPower() <= limit
	}
}

var nihiloorStealBody = game.ReflexiveBody("nihiloor/steal", effects.SimpleBody(nihiloorSteal),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return effects.TargetCreature("target creature that player controls with power less than or equal to the tapped creature's power",
			nihiloorPowerAtMostTapped(p.Player, p.Object, p.Amount))
	})

func nihiloorSteal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		d, ok := effects.DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller())
		if !ok {
			return nil
		}
		return effects.GainControl{Target: t.ID, Duration: d, Label: "Nihiloor — gain control for as long as you control Nihiloor"}.Apply(ctx)
	}
	return nil
}
