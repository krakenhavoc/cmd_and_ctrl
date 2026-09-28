package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cant_have.go — the card-facing half of #1651 (ADR 0038's amendment of
// 2026-09-28): "loses <keyword> and can't have <keyword>" and the
// turn-scoped hexproof waiver.
//
//	Static: []game.StaticAbility{
//		LoseAndCantHave(b27CreaturesOpponentsControl, "hexproof"), // Archetype of Endurance
//	},
//	LoseAndCantHaveUntilEOT{Match: …, Keywords: …}.Apply(ctx)    // Arcane Lighthouse
//	WaiveHexproofUntilEOT{Label: …}.Apply(ctx)                     // Detection Tower
//
// "Loses X" on its own (Shadowspear, Colossus Hammer) is still an
// ordinary layer-6 removal that a later grant can undo — CR 613.7 at
// work. Only a card that prints "can't have" (or "can't have or gain")
// uses these: the engine then strips the keyword after the whole
// layer-6 bucket, so no grant of any timestamp puts it back
// (game/cant_have.go).

// LoseAndCantHave is the static "<permanents> lose <keywords> and
// can't have or gain <keywords>" — the Archetype cycle's second line.
// One layer-6 static: it records the tokens, and the engine's strip
// after layer 6 removes them whatever granted them, the printed
// baseline included. `applies` is read live on every pass (a static
// has no locked set, CR 611.3a).
func LoseAndCantHave(applies func(target *game.Card, g *game.Game, source *game.Card) bool, keywords ...string) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: applies,
		Apply:     game.CantHaveKeywords(keywords...),
	}
}

// LoseAndCantHaveUntilEOT is "until end of turn, <permanents> lose
// <keywords> and can't have <keywords>" (Arcane Lighthouse). It
// changes characteristics, so CR 611.2c locks the set to what Match
// selects as it resolves, exactly as GrantKeywordUntilEOT does; a
// creature that comes under an opponent's control later is untouched.
type LoseAndCantHaveUntilEOT struct {
	// Target pins the effect to one permanent. Ignored when Match is
	// set.
	Target uuid.UUID

	// Match selects the affected permanents, evaluated ONCE at
	// resolution (CR 611.2c).
	Match CardPredicate

	// Keywords are canonical lowercase tokens.
	Keywords []string

	Label string
}

func (k LoseAndCantHaveUntilEOT) Apply(ctx *Context) error {
	if len(k.Keywords) == 0 {
		return nil
	}
	return untilEndOfTurn(ctx, k.Target, k.Match,
		eotLabel(k.Label, "loses and can't have until end of turn"),
		game.CantHaveKeywordsMod(k.Keywords...))
}

// WaiveHexproofUntilEOT is "until end of turn, your opponents and
// creatures your opponents control with hexproof can be the targets of
// spells and abilities you control as though they didn't have
// hexproof" (Detection Tower). A waiver changes no characteristic, so
// its set is a live rule read until cleanup
// (game.ScopeOpponentsAndTheirCreatures), and "you" is the controller
// of the resolving ability.
type WaiveHexproofUntilEOT struct {
	Label string
}

func (w WaiveHexproofUntilEOT) Apply(ctx *Context) error {
	ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeOpponentsAndTheirCreatures,
		ctx.Controller(), []game.Mod{game.WaiveHexproofMod()}, DurationUntilEndOfTurn(ctx),
		eotLabel(w.Label, "hexproof waived until end of turn"))
	return nil
}
