package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cant_be_countered_by_color.go — #6263, Autumn's Veil:
//
//	"Spells you control can't be countered by blue or black spells this
//	 turn, and creatures you control can't be the targets of blue or
//	 black spells this turn."
//
// The first clause is a counter-shield turn grant narrowed to counters
// made by spells of those colours (game.CounterShieldGrant.ByColors),
// given with GrantCounterShield:
//
//	GrantCounterShield{From: "Autumn's Veil", Grant: SpellsCantBeCounteredByColorsThisTurn(
//	    "Spells you control can't be countered by blue or black spells this turn.",
//	    game.CounterShieldYouControl, game.PermissionFilter{}, "UB")}
//
// The second is CantBeTargetedBySpellsOfColorsThisTurn.

// SpellsCantBeCounteredByColorsThisTurn is SpellsCantBeCounteredThisTurn
// narrowed to counters by spells of `colors` (uppercase WUBRG letters run
// together, "UB").
func SpellsCantBeCounteredByColorsThisTurn(text string, whose game.CounterShieldWhose, filter game.PermissionFilter, colors string) game.CounterShieldGrant {
	gr := SpellsCantBeCounteredThisTurn(text, whose, filter)
	gr.ByColors = colors
	return gr
}

// CantBeTargetedBySpellsOfColorsThisTurn is "<Scope> can't be the targets
// of <colour> spells this turn": a live rule record (CR 611.2c) over
// Scope (zero: game.ScopeYourCreatures, "creatures you control"), the
// "you" being the controller of the resolving spell or ability. Colors
// are uppercase WUBRG letters. It refuses spells of those colours,
// whoever controls them, and no ability.
type CantBeTargetedBySpellsOfColorsThisTurn struct {
	Scope  game.AffectedScope
	Colors []string
	Label  string
}

func (a CantBeTargetedBySpellsOfColorsThisTurn) Apply(ctx *Context) error {
	scope := a.Scope
	if scope == game.ScopeNone {
		scope = game.ScopeYourCreatures
	}
	ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), scope, ctx.Controller(),
		[]game.Mod{game.CantBeTargetedBySpellsOfColorsMod(a.Colors...)}, ctx.Game.UntilEndOfTurnDuration(), a.Label)
	return nil
}
