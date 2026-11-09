package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Squall, Gunblade Duelist — Legendary Creature — Human Mercenary
// {R}{W}{B}, 3/2 (Final Fantasy, #4462):
//
//	"First strike
//	 As Squall enters, choose a number.
//	 Whenever one or more creatures attack one of your opponents, if any
//	 of those creatures have power or toughness equal to the chosen number,
//	 Squall deals damage equal to its power to defending player."
//
// The number is a CR 614.12 "as enters" choice, Nyota Uhura's, with no
// printed ceiling: any whole number from 0 up (CR 107.1b, 107.1c). The
// trigger is one occurrence per player attacked in a declaration
// (OncePerBatchPerPlayer, CR 603.2c), and an attack at a planeswalker or
// battle is not an attack at a player (CR 508.3e). "Those creatures" are
// the creatures attacking that player, whoever controls them. The
// intervening "if" (CR 603.4) is asked as the attack is declared, and again
// as the ability resolves, from the same creatures: each read as it is now,
// or as it last existed on the battlefield once it has left (CR 608.2h).
// The damage is Squall's power as it is now, or its last-known power, dealt
// by Squall (lifelink and deathtouch as it had them) to the defending
// player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "880c2b0a-b708-4f11-a88a-ec990c904602",
		Name:            "Squall, Gunblade Duelist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		AsEnters:        squallChooseANumber,
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					defender := squallDefender(g, ev, source)
					return defender != uuid.Nil && squallAnyMatchesNow(squallAttackers(g, defender), source.ChosenNumber)
				},
				Key: squallLabel,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					defender := squallDefender(g, ev, source)
					item := game.NewTriggeredItem(source, squallLabel)
					item.Params.Player = defender
					item.Params.Amount = source.ChosenNumber
					for _, c := range squallAttackers(g, defender) {
						item.Params.Objects = append(item.Params.Objects, game.ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch})
					}
					return item
				},
				Effect: squallDamage,
			}),
		},
	})
}

const squallLabel = "Squall, Gunblade Duelist — deals damage equal to its power to defending player"

// squallChooseANumber is "as Squall enters, choose a number", stored on the
// permanent as it enters.
func squallChooseANumber(card *game.Card, ctx *Context) error {
	return ctx.Game.QueueChooseNumberForEffect(game.ChooseNumber{
		Chooser:  card.Controller,
		Source:   card.InstanceID,
		Resource: game.PayResourceNone,
		Min:      0,
		NoMax:    true,
		Goal:     3,
		Question: "Squall, Gunblade Duelist — choose a number",
		Branch:   game.NumberBranch{Then: storeChosenNumber, Carry: []uuid.UUID{card.Controller}},
	})
}

// squallDefender is the player the attack names when it is one of the
// source controller's opponents, else uuid.Nil.
func squallDefender(g *game.Game, ev game.Event, source *game.Card) uuid.UUID {
	p := attackedPlayer(g, ev)
	if p == uuid.Nil || p == source.Controller {
		return uuid.Nil
	}
	return p
}

// squallAttackers lists the creatures attacking `defender` now.
func squallAttackers(g *game.Game, defender uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.AttackingTarget == defender {
			out = append(out, c)
		}
	}
	return out
}

// squallAnyMatchesNow is "any of those creatures have power or toughness
// equal to the chosen number" as the attack is declared.
func squallAnyMatchesNow(attackers []game.Card, n int) bool {
	for _, c := range attackers {
		if c.CurrentPower() == n || c.CurrentToughness() == n {
			return true
		}
	}
	return false
}

// squallDamage rechecks the intervening "if", each creature as it is now or
// as it last existed once it has left (CR 608.2h), and deals the damage.
func squallDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	ref, refOK := ctx.SourceRef()
	if !ok || !refOK || info.Power <= 0 {
		return nil
	}
	matches := false
	for _, obj := range item.Params.Objects {
		if p, found := g.PermanentForEffect(obj); found && (p.Power == item.Params.Amount || p.Toughness == item.Params.Amount) {
			matches = true
			break
		}
	}
	if !matches {
		return nil
	}
	return DealDamage{SourceObject: &ref, Target: item.Params.Player, Amount: info.Power}.Apply(ctx)
}
