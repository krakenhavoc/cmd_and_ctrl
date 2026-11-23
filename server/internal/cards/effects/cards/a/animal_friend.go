package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Animal Friend — Enchantment — Aura {1}{G} (Fallout):
//
//	"Enchant creature
//	 Enchanted creature has "Whenever this creature attacks, create a 1/1
//	 green Squirrel creature token. Put a +1/+1 counter on that token for
//	 each Aura and Equipment attached to this creature other than Animal
//	 Friend.""
//
// Light of Promise's shape: the Aura's static grants the creature a trigger
// of its own, so "this creature" is the enchanted creature and "you" is its
// controller. The token is made through the keyed creation (Ajani's Chosen)
// so the counters are PUT on it afterwards, not entered with, and the count
// of Auras and Equipment attached to the creature, the granting Aura not
// among them, is taken as the tail runs, in the same resolution (CR 608.2c).
//
// No simplification.
const animalFriendGrant = "animal-friend/squirrel"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae4249de-1b8c-4b15-8960-180557a15065",
		Name:         "Animal Friend",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Tokens: 1},
		Grants: []effects.AbilityGrant{{
			Key: animalFriendGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisAttacks("Animal Friend — create a 1/1 green Squirrel creature token and put a +1/+1 counter on it for each other Aura and Equipment attached to this creature",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						tail := ctx.ItemTail(animalFriendCounters)
						tail.N = animalFriendOtherAttachments(ctx)
						return g.CreateTokensKeyedForEffect(game.TokenCreation{
							Controller: item.Controller,
							Source:     item.SourceCardID,
							Groups: []game.TokenGroup{{
								Template: ctx.TextToken(effects.TokenCard("1/1 green Squirrel")),
								Count:    1,
							}},
						}, tail)
					}),
			},
			Text: "Whenever this creature attacks, create a 1/1 green Squirrel creature token. Put a +1/+1 counter on that token for each Aura and Equipment attached to this creature other than Animal Friend.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(animalFriendGrant)},
	})
}

// animalFriendOtherAttachments is the number of Auras and Equipment attached
// to the enchanted creature other than the Aura that granted the ability,
// counted as the trigger resolves (CR 608.2c) and carried to the tail as N.
func animalFriendOtherAttachments(ctx *effects.Context) int {
	host, grantor := ctx.Source(), ctx.Grantor()
	n := 0
	for _, id := range ctx.Game.AttachmentsOf(host) {
		if id == grantor {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && (c.HasSubtype("Aura") || c.HasSubtype("Equipment")) {
			n++
		}
	}
	return n
}

// animalFriendCounters puts a +1/+1 counter on the Squirrel for each of
// those attachments (the tail's N). An on-disk key: never renamed, never
// reused.
var animalFriendCounters = effects.CardTail("animal-friend/counters-for-attachments", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || o.Tail.N <= 0 {
		return nil
	}
	return effects.AddCounter{Target: o.IDs[0], Kind: game.CounterPlusOne, N: o.Tail.N}.Apply(ctx)
})
