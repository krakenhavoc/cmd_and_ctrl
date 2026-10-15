package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blasphemous Edict — Sorcery {3}{B}{B} (EDHREC top 1,000, #3654):
//
//	"You may pay {B} rather than pay this spell's mana cost if there
//	 are thirteen or more creatures on the battlefield.
//	 Each player sacrifices thirteen creatures of their choice."
//
// The alternative cost is Snuff Out's conditional offer, with the
// condition read off the whole battlefield (any controller, tokens
// included) as the spell is cast. The edict is Grave Pact's "each
// player", thirteen times over: every player, the caster included, is
// asked in turn for thirteen creatures, and a player with fewer
// sacrifices what they have.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6cf50e21-0f60-4e9a-b910-ebe1bad2a29e",
		Name:         "Blasphemous Edict",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:      "thirteen_creatures",
			Label:    "Pay {B} (thirteen or more creatures on the battlefield)",
			ManaCost: "{B}",
			Condition: func(g *game.Game, _ uuid.UUID) bool {
				n := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.IsCreature() {
						n++
					}
				}
				return n >= 13
			},
		}},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, p := range seatsFromActive(ctx.Game) {
				ctx.Game.PlayerSacrificesNForEffect(ctx.Source(), p,
					sacrificeSpec("a creature", Creature()),
					"Blasphemous Edict — sacrifice thirteen creatures", 13)
			}
			return nil
		},
	})
}
