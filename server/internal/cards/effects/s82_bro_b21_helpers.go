package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82BroB21XGolem creates one X/X colorless Golem artifact creature token
// under controller's control (the printed token card is */*, so the size is
// set as it is created, Vulshok Factory's way). When haste is set the token
// gains haste until end of turn (Rootwire Amalgam). A size below zero is
// zero, and the token then dies as a state-based action (CR 704.5f).
func s82BroB21XGolem(ctx *Context, x int, haste bool) error {
	tmpl := s82OneB12Token("Golem", "Token Artifact Creature — Golem", x, x, nil, nil)
	ids, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(tmpl), 1, game.TokenEntryOptions{})
	if err != nil || !haste {
		return err
	}
	for _, id := range ids {
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"},
			Label: "Rootwire Amalgam — the Golem gains haste until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// s82BroB21WasCast is "if you cast it" on an enters trigger: the entering
// permanent came from a spell (CR 400.7d, game/cast_provenance.go). It is
// the intervening if (CR 603.4); each effect rechecks it as it resolves.
func s82BroB21WasCast(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && source.Provenance.FromZone != ""
}

// s82BroB21EntersIfCast is a "When this creature enters, if you cast it, ..."
// trigger: ETB of this permanent, with the intervening if checked both as it
// triggers and as it resolves.
func s82BroB21EntersIfCast(label string, effect func(ctx *Context) error) game.TriggeredAbility {
	return On(game.EventETB, AllOf(Self, s82BroB21WasCast), label,
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			if ctx.CastProvenance().FromZone == "" {
				return nil
			}
			return effect(ctx)
		})
}
