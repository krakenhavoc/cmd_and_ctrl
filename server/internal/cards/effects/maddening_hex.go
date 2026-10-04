package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Maddening Hex — Enchantment — Aura Curse {1}{R}{R}:
//
//	"Enchant player
//	 Whenever enchanted player casts a noncreature spell, roll a d6.
//	 This Aura deals damage to that player equal to the result. Then
//	 attach this Aura to another one of your opponents chosen at
//	 random."
//
// Curse of Opulence's attachment to a player, a Kambal-shaped cast
// trigger, and the keyed random stream (random_effects.go): the d6 is
// rolled through rollDice, so a "whenever you roll" card sees it, and
// the new host is drawn from the same stream (ChooseAtRandomForEffect). The
// damage is dealt by the Aura itself, as printed, to the player who
// cast the spell (read off the item's triggering event, so a curse that
// moved before this resolves still hits the right player).
//
// "Another one of YOUR opponents" is the Aura's controller's
// opponents, not the enchanted player's, and excludes the current host
// and anyone who has left the game. With nobody else to move to, the
// Aura stays where it is. A spell that is countered still triggered
// this on the cast, so it still burns.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5d96170f-14dd-496e-a485-bd3b9eedfe02",
		Name:         "Maddening Hex",
		Completeness: CompletenessFull,
		Targets:      EnchantPlayer(),
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if source.AttachedTo.Kind != game.TargetPlayer || source.AttachedTo.ID != ev.Actor {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !spell.IsCreature()
			},
			Key:    "Maddening Hex — roll a d6, deal that much damage to that player, then attach to another opponent at random",
			Effect: maddeningHexRollAndMove,
		}},
	})
}

// maddeningHexRollAndMove is the whole trigger: roll, burn the caster,
// then move the Aura to a random other opponent of its controller.
func maddeningHexRollAndMove(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	caster := item.Trigger.Event.Actor
	results, err := rollDice(ctx, 6, 1)
	if err != nil || len(results) == 0 {
		return err
	}
	if p := g.PlayerByIDForEffect(caster); p != nil && !p.Eliminated {
		if err := (DealDamage{Source: item.SourceCardID, Target: caster, Amount: results[0]}).Apply(ctx); err != nil {
			return err
		}
	}
	var others []uuid.UUID
	for _, opp := range ctx.Opponents() {
		if p := g.PlayerByIDForEffect(opp); p != nil && !p.Eliminated {
			others = append(others, opp)
		}
	}
	// "Another one of your opponents" is every opponent but the player
	// the Aura enchants as this resolves.
	if aura, ok := g.LookupCardForEffect(item.SourceCardID); ok && aura.AttachedTo.Kind == game.TargetPlayer {
		kept := others[:0]
		for _, id := range others {
			if id != aura.AttachedTo.ID {
				kept = append(kept, id)
			}
		}
		others = kept
	}
	pick := g.ChooseAtRandomForEffect(randomDraw(ctx), others, 1)
	if len(pick) == 0 {
		return nil
	}
	return g.AttachForEffect(item.SourceCardID, game.TargetRef{Kind: game.TargetPlayer, ID: pick[0]})
}
