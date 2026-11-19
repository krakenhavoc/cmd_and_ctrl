package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// CantBlockSubtypesThisTurn is "<creature type>s can't block this turn"
// (Pyrophobia's Cowards): one live rule record over every creature, so a
// creature of the type that arrives later is bound too and one that stops
// being it is freed (CR 611.2c: the effect changes no characteristic).
type CantBlockSubtypesThisTurn struct {
	Subtypes []string
	// Text is the clause as the refusal sentence reads it ("Cowards").
	Text  string
	Label string
}

// Apply registers the record. Caller is inside the resolution frame.
func (r CantBlockSubtypesThisTurn) Apply(ctx *Context) error {
	if len(r.Subtypes) == 0 {
		return nil
	}
	ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeCreatures, ctx.Controller(),
		[]game.Mod{game.CantBlockSubtypesMod(r.Subtypes, r.Text)}, DurationUntilEndOfTurn(ctx),
		eotLabel(r.Label, r.Text+" can't block this turn"))
	return nil
}
