package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ob Nixilis, the Hate-Twisted — Legendary Planeswalker — Nixilis
// {3}{B}{B}, starting loyalty 5 (EDHREC rank 4099):
//
//	"Whenever an opponent draws a card, Ob Nixilis deals 1 damage to
//	 that player.
//	 −2: Destroy target creature. Its controller draws two cards."
//
// The group-slug walker. Five loyalty is a lot of body, the passive
// punishes exactly the card-advantage engines that win Commander
// games, and the minus is unconditional creature removal that only
// looks like it helps the victim — with the passive on the
// battlefield, the two cards they draw are two damage back.
//
// THE PASSIVE IS A TRIGGERED ABILITY ON A PLANESWALKER, which is
// worth saying out loud because it is the shape most planeswalkers do
// not have: it is not a loyalty ability, it uses the stack, and it
// fires once per card drawn — a draw-three is three triggers and
// three damage, not one. It watches OPPONENTS only, so Ob Nixilis's
// controller draws freely.
//
// "THAT PLAYER" is the drawer, read off the draw event, not "target
// opponent": the ability does not target, so hexproof-style
// protections and "can't be the target" effects do nothing about it.
//
// THE −2 HELPS THE VICTIM, AND THAT IS PRINTED. "Its controller draws
// two cards" is not optional and is not a drawback the engine may
// quietly drop — a version that destroyed without the draw would be
// STRONGER than printed, which the #259 rule forbids. It fires the
// controller's own draw replacements and, with Ob Nixilis still
// there, his own passive twice.
//
// The two cards go to the creature's CONTROLLER as read at
// resolution. A target that became illegal in response is skipped
// entirely (CR 608.2b) and nobody draws.
//
// No simplification: Ob Nixilis has no ultimate, so nothing is
// omitted.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88cd12e5-ca89-4176-96fd-320fe2ccd087",
		Name:         "Ob Nixilis, the Hate-Twisted",
		Completeness: effects.CompletenessFull,
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 5,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDrawCard},
			AppliesTo: effects.ByAnOpponent,
			Key:       "Ob Nixilis — 1 damage to that player",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DealDamage{
					Source: item.SourceCardID,
					Target: item.Trigger.Event.Actor,
					Amount: 1,
				}.Apply(effects.NewContext(g, item))
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "−2: Destroy target creature. Its controller draws two cards.",
			Cost:    effects.LoyaltyCost(-2),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				// The controller is read BEFORE the destruction —
				// once the creature is in the graveyard,
				// Card.Controller tracks its owner instead.
				victim, ok := g.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				owner := victim.Controller
				if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
				return effects.DrawCards{Player: owner, N: 2}.Apply(ctx)
			},
		}},
	})
}
