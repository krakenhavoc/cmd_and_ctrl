package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dusk Mangler — Creature — Horror {5}{B}{B}, 5/4 (Streets of New Capenna,
// S82):
//
//	"As an additional cost to cast this spell, sacrifice a creature, discard
//	 a card, or pay 4 life.
//	 When this creature enters, each opponent sacrifices a creature of their
//	 choice, discards a card, and loses 4 life."
//
// The cost is one mandatory either/or cost (ADR 0100, EitherCost): the
// caster announces the branch as the spell is cast and only that branch is
// paid.
//
// The entry trigger is Arbiter of Woe's, with Scarring Memories' printed
// order (CR 608.2c): every opponent sacrifices a creature of their choice
// (the edict, nothing targets), and only once each has chosen, every
// opponent discards a card of their choice, then each loses 4 life. An
// opponent with no creature or no card skips that part and still does the
// rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74de3356-9779-40f9-9881-fe795a1d5bbd",
		Name:         "Dusk Mangler",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.EitherCost(
			effects.SacrificeCost("a creature", effects.Creature()).Keyed("sacrifice"),
			effects.DiscardCost(1).Keyed("discard"),
			effects.PayLifeCost(4).Keyed("life"),
		),
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dusk Mangler — each opponent sacrifices a creature of their choice, discards a card, and loses 4 life",
				effects.Do(effects.EachPlayerSacrifices{ExceptController: true, Match: effects.Creature(), Label: "a creature", Keyed: duskManglerEdicted})),
		},
	})
}

// duskManglerEdicted is the rest of the trigger once every opponent has
// answered the edict and the creatures have gone (EachPlayerSacrifices.Keyed,
// #4290): each opponent discards a card, then each loses 4 life. An on-disk
// key: never renamed, never reused.
var duskManglerEdicted = effects.RegisterEdictThen("prompt-run/dusk-mangler-discard-and-life", func(ctx *effects.Context, _ game.PromptedSacrifices) error {
	if err := effects.EachOpponentDiscardsOne(ctx.Game, ctx.Item); err != nil {
		return err
	}
	return effects.EachOpponentLosesLife{N: 4}.Apply(ctx)
})
