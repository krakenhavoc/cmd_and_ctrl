package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Valor's Reach Tag Team — Sorcery, the BACK FACE of Invasion of Kylem
// (face 1):
//
//	"Create two 3/2 red and white Warrior creature tokens with
//	 'Whenever this token and at least one other creature token attack,
//	 put a +1/+1 counter on this token.'"
//
// Registered under "<oracle_id>#1". The tokens' printed trigger rides
// the token catalog (ADR 0083). Like Sokka's, it matches only the
// token's own attack event, which the engine emits after every attacker
// of the declaration is stamped, so one declaration is one trigger per
// Warrior; the "other" attacker must be a creature TOKEN its controller
// controls, a nontoken attacker not counting.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     invasionOfKylemOracleID + "#1",
		Name:         "Valor's Reach Tag Team",
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{
				Controller: ctx.Controller(),
				Template:   tokenFromCatalog(printedValorsReachWarriorToken),
				N:          2,
			}.Apply(ctx)
		},
	})
}

// printedValorsReachWarriorToken is the 3/2 red and white Warrior as
// printed, trigger included.
func printedValorsReachWarriorToken() tokenTemplate {
	return tokenTemplate{
		Slug: "warrior-valors-reach",
		Card: game.Card{
			Name:      "Warrior",
			TypeLine:  "Token Creature — Warrior",
			Power:     3,
			Toughness: 2,
			Colors:    []string{"R", "W"},
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclared(ev, source) && otherAttackingCreatureTokens(g, source) >= 1
			}, "Warrior — put a +1/+1 counter on this token", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Text: "Whenever this token and at least one other creature token attack, put a +1/+1 counter on this token.",
	}
}

// otherAttackingCreatureTokens counts the attacking creature tokens
// source's controller controls other than source.
func otherAttackingCreatureTokens(g *game.Game, source *game.Card) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == source.Controller && c.InstanceID != source.InstanceID && c.IsCreature() &&
			c.IsToken() && c.AttackingTarget != uuid.Nil {
			n++
		}
	}
	return n
}
