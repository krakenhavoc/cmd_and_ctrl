package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flamescroll Celebrant // Revel in Silence — modal double-faced card
// (Strixhaven: School of Mages, #5502):
//
//	Flamescroll Celebrant — Creature — Human Shaman {1}{R}, 2/1
//	 "Whenever an opponent activates an ability that isn't a mana ability,
//	  this creature deals 1 damage to that player.
//	  {1}{R}: This creature gets +2/+0 until end of turn."
//	Revel in Silence — Instant {W}{W}
//	 "Your opponents can't cast spells or activate planeswalkers' loyalty
//	  abilities this turn. Exile Revel in Silence."
//
// # Two keys, one card
//
// The creature keeps the bare oracle ID and the instant takes
// "<oracle_id>#1" (Wandering Archaic's keyspace).
//
// # The front face
//
// Harsh Mentor's trigger with no source filter and no mana abilities: the
// activator is read off the triggering event when it resolves, and the damage
// comes from the Celebrant. The pump is Firebreathing.
//
// # The back face
//
// Silence's outright cast ban on each opponent until end of turn, plus
// game.GrantCantActivateLoyaltyAbilitiesForEffect (#5502): a player static
// that refuses a loyalty ability of a planeswalker (CR 606.1) and nothing else,
// so other abilities and mana abilities are untouched. Then the spell exiles
// itself, as Temporal Mastery does, so it is not left in the graveyard.
//
// No simplification on either face.
func init() {
	const id = "5f1d3d22-18f7-42b7-8039-ef7aa5d29856"
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Flamescroll Celebrant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverAnOpponentActivates("Flamescroll Celebrant — 1 damage to that player", nil, false,
				func(activator uuid.UUID) effects.Effect {
					return func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						return effects.DealDamage{Source: ctx.Source(), Target: activator, Amount: 1}.Apply(ctx)
					}
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 0, "Flamescroll Celebrant — +2/+0"),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Revel in Silence",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			d := effects.DurationUntilEndOfTurn(ctx)
			for _, opp := range ctx.Opponents() {
				if err := (effects.RestrictCasting{
					Player:   opp,
					Rule:     game.CastBanRule{Kind: game.CastBanOutright},
					Label:    "Revel in Silence — can't cast spells this turn",
					Duration: d,
				}).Apply(ctx); err != nil {
					return err
				}
				ctx.Game.GrantCantActivateLoyaltyAbilitiesForEffect(opp,
					"Revel in Silence — can't activate planeswalkers' loyalty abilities this turn", ctx.Source(), d)
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
