package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// CantBeTargetedByOpponentSpellsOfColorsThisTurn is "<Scope> can't be the
// targets of <colour> spells your opponents control this turn" (Display of
// Dominance): a live rule record (CR 611.2c) over Scope (zero:
// game.ScopeYourPermanents), the "you" being the controller of the resolving
// spell. Colors are uppercase WUBRG letters. It refuses only spells the
// controller's opponents control, and no ability.
type CantBeTargetedByOpponentSpellsOfColorsThisTurn struct {
	Scope  game.AffectedScope
	Colors []string
	Label  string
}

func (a CantBeTargetedByOpponentSpellsOfColorsThisTurn) Apply(ctx *Context) error {
	scope := a.Scope
	if scope == game.ScopeNone {
		scope = game.ScopeYourPermanents
	}
	ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), scope, ctx.Controller(),
		[]game.Mod{game.CantBeTargetedByOpponentSpellsOfColorsMod(a.Colors...)}, ctx.Game.UntilEndOfTurnDuration(), a.Label)
	return nil
}

// ExileOneOrMoreFromGraveyard is "Exile one or more <kind> cards from your
// graveyard" as a CR 602 activated ability's cost (Corpseweft): any number
// of matching cards, at least one. The effect reads how many were exiled
// through ctx.Exiled().
func ExileOneOrMoreFromGraveyard(label string, match func(game.Card) bool) game.AbilityCost {
	return game.AbilityCost{ExileCards: &game.ExileCost{
		AnyNumber: true, MinCards: 1, Label: label, From: game.ZoneGraveyard, Match: match,
	}}
}

// TuckTargetCreatureCardsOnTopThenDraw is "Put any number of target creature
// cards from your graveyard on top of your library. Draw a card." (Forever
// Young, Gravepurge): the legal targets go on top, then the controller draws.
func TuckTargetCreatureCardsOnTopThenDraw() func(*game.StackItem, *Context) error {
	return ResolveSteps(
		Instruction(TuckLegalCardTargetsOnTop),
		DrawCards{N: 1},
	)
}

// TuckLegalCardTargetsOnTop puts each still-legal card target on top of its
// OWNER's library, in target order (Forever Young, Plow Under).
func TuckLegalCardTargetsOnTop(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := ctx.Game.TuckToLibraryForEffect(t.ID, false); err != nil {
			return err
		}
	}
	return nil
}
