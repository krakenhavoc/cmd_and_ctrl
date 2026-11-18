package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flawless Forgery — Sorcery {3}{U}{U} (Streets of New Capenna Commander,
// #5491):
//
//	"Casualty 3 (As you cast this spell, you may sacrifice a creature with
//	 power 3 or greater. When you do, copy this spell and you may choose a
//	 new target for the copy.)
//	 Exile target instant or sorcery card from an opponent's graveyard. Copy
//	 that card. You may cast the copy without paying its mana cost."
//
// Casualty is the engine's keyword (Casualty(3), game/casualty.go). The
// body is Arcane Proxy's exile-and-copy with an opponent's graveyard as the
// target: only a card that really reached exile is copied, and the copy is
// cast during the resolution or ceases to exist (CopyAndCast, CR 707.12,
// CR 608.2g). The target is re-judged as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "0e3cb267-5704-4708-94c8-e2fbce4978af",
		Name:          "Flawless Forgery",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Casualty(3)},
		Targets: TargetCardInGraveyard("target instant or sorcery card from an opponent's graveyard",
			Or(Instant(), Sorcery()), Not(YouOwn())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return arcaneProxyExileAndCopy(ctx)
		},
	})
}
