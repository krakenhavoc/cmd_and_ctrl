package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b208_helpers.go — shared bodies for Dominaria kicker cards, batch 208 (#5530).

// OldB208DamageClauseTarget deals `amount` damage from the resolving spell to the
// pick of target clause `clause`, when it is still legal (CR 608.2b).
func OldB208DamageClauseTarget(ctx *Context, clause, amount int) error {
	t, ok := ctx.ClauseTarget(clause)
	if !ok {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: amount}.Apply(ctx)
}

// OldB208ThisWasntKicked is "if this creature wasn't kicked": the permanent's
// cast record names no kick. A permanent that was not cast was not kicked.
func OldB208ThisWasntKicked(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return game.CardKickedTimes(*source) == 0
}

// OldB208SacrificeThisIfNotKicked is Skizzik's end-step body: sacrifice the
// source, when it is still the same permanent, still unkicked and still under
// the trigger controller's control (an opponent's trigger cannot make you
// sacrifice what is yours now).
func OldB208SacrificeThisIfNotKicked(g *game.Game, item *game.StackItem) error {
	if !B09SourceStillOnBattlefield(g, item) {
		return nil
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || c.Controller != item.Controller || game.CardKickedTimes(c) != 0 {
		return nil
	}
	return g.SacrificePermanentForEffect(item.SourceCardID)
}

// OldB208ThreatenTarget is Act of Treason's body as a trigger's or spell's
// Effect: gain control of the first legal target creature until end of turn,
// untap it, and it gains haste until end of turn.
func OldB208ThreatenTarget(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (GainControl{Target: id, Duration: DurationUntilEndOfTurn(ctx), Label: name + " — gain control until end of turn"}).Apply(ctx); err != nil {
			return err
		}
		if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: name + " — haste until end of turn"}.Apply(ctx)
	}
}

// OldB208AttackedAlone is "attacks alone": the declaration this event belongs to
// named exactly one attacker (CR 506.5, game.AttackedAlone), read as the
// trigger is harvested.
func OldB208AttackedAlone(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	return game.AttackedAlone(g)
}
