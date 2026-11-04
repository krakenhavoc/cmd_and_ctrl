package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Channel — Sorcery {G}{G} (Alpha; #4885):
//
//	"Until end of turn, any time you could activate a mana ability, you
//	 may pay 1 life. If you do, add {C}."
//
// The permission is a mana ability the PLAYER has until end of turn,
// bound to no card (game.GrantPlayerManaAbilityForEffect,
// player_mana_ability.go): the caster may use it whenever a mana
// ability may be activated, as many times as they can pay, and never
// with less than 1 life (CR 119.4). It ends in the cleanup step. The
// auto-tapper never spends the life for the player; they press it.
//
// Yavimaya Bloomsage's prepare spell is this card word for word and
// shares channelOnResolve.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d1b815d1-2848-40d4-a555-66822d1becbc",
		Name:         "Channel",
		Completeness: CompletenessFull,
		OnResolve:    channelOnResolve,
	})
}

// channelOnResolve gives the caster "pay 1 life: add {C}" until end of
// turn, attributed to the resolving spell.
func channelOnResolve(item *game.StackItem, ctx *Context) error {
	ctx.Game.GrantPlayerManaAbilityForEffect(ctx.Controller(),
		game.PlayerManaAbility{LifeCost: 1, Produced: "{C}"},
		"Channel", ctx.Source(), ctx.Game.UntilEndOfTurnDuration())
	return nil
}
