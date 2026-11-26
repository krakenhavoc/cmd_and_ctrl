package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stain the Mind — Sorcery {4}{B} (Magic 2015, #5556):
//
//	"Convoke
//	 Choose a nonland card name. Search target player's graveyard, hand, and
//	 library for any number of cards with that name and exile them. Then that
//	 player shuffles."
//
// Lost Legacy's resolution (the keyed card-name prompt, free text, CR 201.2)
// with convoke and no draw. "Nonland" is enforced where the engine can: only
// cards that are not lands are exiled, so naming a land exiles nothing. "Any
// number of cards" is all of them: exiling fewer is never better for the
// caster. The target player shuffles after the search whether or not anything
// was found. With the target player illegal as the spell resolves it does
// nothing (CR 608.2b) and no name is chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1994d563-8726-49c9-a18a-6ba2604a8cb7",
		Name:         "Stain the Mind",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetPlayer {
				return nil
			}
			ctx.Game.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
				"Stain the Mind — choose a nonland card name",
				effects.WithCarry(ctx.ItemTail(stainTheMindNamed), legal[0].ID))
			return nil
		},
	})
}

// stainTheMindNamed is the rest of the spell once the name (the tail's Word)
// is known: every nonland card with the name in the target player's
// graveyard, hand and library is exiled, then that player shuffles. An
// on-disk key.
var stainTheMindNamed = effects.CardTail("then/stain-the-mind-named", func(ctx *effects.Context, o game.TailOutcome) error {
	target, name := effects.TailCarry0(o), o.Tail.Word
	p := ctx.PlayerByID(target)
	if p == nil || name == "" {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(effects.S82OldB358NonlandCardsNamedInAllZones(p, name), game.KeyedTail{
		Ref: stainTheMindExiled, Player: target, Source: ctx.Source(),
	})
})

// stainTheMindExiled is "then that player shuffles". An on-disk key.
var stainTheMindExiled = effects.CardTail("tail/stain-the-mind-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	return ctx.Game.ShuffleLibraryForEffect(o.Tail.Player)
})
