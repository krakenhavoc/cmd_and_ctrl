package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fireflux Squad — {3}{R} Creature — Human Soldier 4/3 (Commander 2020):
//
//	"Haste
//	 Whenever this creature attacks, you may exile another target attacking
//	 creature you control. If you do, reveal cards from the top of your
//	 library until you reveal a creature card. Put that card onto the
//	 battlefield tapped and attacking and the rest on the bottom of your
//	 library in a random order."
//
// A targeted, optional attack trigger. As it resolves the exiled creature's
// defender is read first, then it is exiled as a keyed move whose tail runs
// only if it went ("if you do", CR 608.2c). The tail is Shifting Shadow's
// reveal-until (RevealUntil); the creature enters tapped and attacking the
// player or permanent the exiled creature was attacking, the way every other
// catalogued "tapped and attacking" entry picks its defender (Hanweir
// Garrison), and is never declared as an attacker, so nothing triggers on its
// attack (CR 508.4). The rest of the revealed cards go to the bottom in a
// random order, or the whole run when there is no creature card.
func init() {
	fluxTail = effects.CardTail("tail/fireflux-squad-reveal", func(ctx *effects.Context, o game.TailOutcome) error {
		if !effects.OneWentThisWay(o) || len(o.Tail.Carry) == 0 {
			return nil
		}
		defender := o.Tail.Carry[0]
		player := ctx.Controller()
		run, hit := effects.RevealUntil(ctx, player, game.Card.IsCreature, "Fireflux Squad — reveal until a creature card")
		tail := game.KeyedTail{Ref: effects.PutRestOnBottomInRandomOrderTail, Player: player, Source: ctx.Source(), Carry: run}
		if hit == uuid.Nil {
			return tail.Run(ctx.Game, nil)
		}
		return ctx.Game.PutCardsFromLibraryOntoBattlefieldKeyedForEffect([]uuid.UUID{hit}, game.LibraryEntryOptions{
			Controller: player,
			Tapped:     true,
			Attacking:  defender,
		}, tail)
	})
	t := effects.Optional(effects.On(game.EventAttack, effects.ThisAttacked, "Fireflux Squad — you may exile another target attacking creature you control; if you do, reveal until a creature card and put it onto the battlefield tapped and attacking",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			for _, tgt := range ctx.LegalTargets() {
				if tgt.Kind != game.TargetCard {
					continue
				}
				c, ok := g.LookupCardForEffect(tgt.ID)
				if !ok || c.AttackingTarget == uuid.Nil {
					return nil
				}
				t := ctx.ItemTail(fluxTail)
				t.Carry = []uuid.UUID{c.AttackingTarget}
				return effects.ExileTarget{Target: tgt.ID, ThenTail: t}.Apply(ctx)
			}
			return nil
		}), "Fireflux Squad — exile another attacking creature you control?")
	t.Targets = effects.Another(effects.TargetCreature("another target attacking creature you control", effects.AttackingCreature(), effects.YouControl()))
	effects.Register(effects.Spec{
		OracleID:        "d53fecd7-aa12-482e-8f45-5efe6348070d",
		Name:            "Fireflux Squad",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{t},
	})
}

var fluxTail game.TailRef
