package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Puppet Conjurer — Artifact Creature — Human Wizard {1}{B}, 1/2 (Shards of
// Alara, S82 OLD batch 510, #5588):
//
//	"{U}, {T}: Create a 0/1 blue Homunculus artifact creature token.
//	 At the beginning of your upkeep, sacrifice a Homunculus."
//
// The sacrifice is mandatory and any Homunculus you control may be chosen
// (the token, or another creature of that type); with none there is nothing
// to do.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06494c26-3822-43ee-bedf-611c9ec7cc6b",
		Name:         "Puppet Conjurer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Create a 0/1 blue Homunculus artifact creature token.",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("0/1 blue Homunculus artifact creature"), N: 1}),
		}},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Puppet Conjurer — sacrifice a Homunculus", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				homunculus := ctx.TextSubtype("Homunculus")
				var candidates []uuid.UUID
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && c.HasSubtype(homunculus) && !g.CantBeSacrificedForEffect(c.InstanceID) {
						candidates = append(candidates, c.InstanceID)
					}
				}
				if len(candidates) == 0 {
					return nil
				}
				return effects.SacrificeChoice{
					Player:     item.Controller,
					Candidates: candidates,
					Question:   "Puppet Conjurer — sacrifice a Homunculus",
				}.Apply(ctx)
			}),
		},
	})
}
