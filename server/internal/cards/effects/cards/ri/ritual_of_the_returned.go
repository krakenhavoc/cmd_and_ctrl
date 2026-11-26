package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ritual of the Returned — Instant {3}{B} (Journey into Nyx, S82 OLD, #5558):
//
//	"Exile target creature card from your graveyard. Create a black Zombie
//	 creature token. Its power is equal to that card's power and its toughness is
//	 equal to that card's toughness."
//
// Soul Separator's exile and sized Zombie: the token is made from the card that
// reached exile, with its power and toughness there (a characteristic-defining
// ability included). If the card left the graveyard in response, the spell is
// countered on resolution for want of its one target and nothing is made
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c74e7441-7563-4954-a1d3-3c0463294455",
		Name:         "Ritual of the Returned",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Targets:      effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return effects.ExileTarget{
				Target: t.ID,
				ThenTail: game.KeyedTail{
					Ref: ritualOfTheReturnedZombie, Player: item.Controller, Source: item.SourceCardID,
					Carry: []uuid.UUID{t.ID},
				},
			}.Apply(ctx)
		},
	})
}

// ritualOfTheReturnedZombie makes the sized Zombie from the card that reached
// exile. An on-disk key.
var ritualOfTheReturnedZombie = effects.CardTail("then/s82-old-b371-ritual-of-the-returned-zombie", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	card, ok := ctx.Game.LookupCardForEffect(o.Tail.Carry[0])
	if !ok {
		return nil
	}
	zombie := ctx.TextToken(effects.TokenCard("0/0 black Zombie"))
	zombie.Power, zombie.Toughness = max(card.PowerForComparison(), 0), max(card.CurrentToughness(), 0)
	return ctx.Game.CreateTokenForEffect(o.Tail.Player, zombie, 1)
})
