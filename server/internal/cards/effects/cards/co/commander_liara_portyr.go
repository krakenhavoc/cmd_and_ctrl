package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander Liara Portyr — Legendary Creature — Human Soldier {3}{R}{W}, 5/3:
//
//	"Whenever you attack, spells you cast from exile this turn cost {X}
//	 less to cast, where X is the number of players being attacked. Exile
//	 the top X cards of your library. Until end of turn, you may cast
//	 spells from among those exiled cards."
//
// "Whenever you attack" is one trigger per attack declaration however
// many creatures attack (OncePerBatch over AttackDeclaredByYou). X is the
// number of distinct players being attacked as the trigger resolves (a
// player whose planeswalker or battle is attacked counts, once), counted
// once and fixed (CR 608.2h).
//
// The reduction is the #6093 timed record
// (game.ModExiledSpellsCostLessForPlayer): every spell you cast from exile
// for the rest of the turn costs {X} less, from these cards or any other
// exile permission, generic mana only (CR 601.2f). The exiled cards may be
// cast (not played: a land among them stays in exile) until end of turn.
// With no player being attacked X is 0 and nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fcc7ed01-ff64-4ffe-9b83-daa4132fd92d",
		Name:         "Commander Liara Portyr",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Commander Liara Portyr — spells you cast from exile this turn cost {X} less; exile the top X cards, cast them this turn",
				commanderLiaraPortyrAttack)),
		},
	})
}

// commanderLiaraPortyrAttack is the trigger's body.
func commanderLiaraPortyrAttack(g *game.Game, item *game.StackItem) error {
	x := effects.GapB218PlayersBeingAttacked(g)
	if x <= 0 {
		return nil
	}
	g.RegisterScopedRuleEffectForEffect(item.SourceCardID, game.ScopeGame, item.Controller,
		[]game.Mod{game.ExiledSpellsCostLessForPlayerMod(item.Controller, x)}, g.UntilEndOfTurnDuration(),
		"Commander Liara Portyr — spells you cast from exile this turn cost less")
	return effects.ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: x, CastOnly: true}.Apply(effects.NewContext(g, item))
}
