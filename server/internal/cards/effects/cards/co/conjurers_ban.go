package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conjurer's Ban — Sorcery {W}{B} (Guildpact, S82 OLD batch 573, #5599):
//
//	"Choose a card name. Until your next turn, spells with the chosen name
//	 can't be cast and lands with the chosen name can't be played.
//	 Draw a card."
//
// The name is asked as the spell resolves and the bans wait on it as a
// registered tail told the name as its Word (Academic Probation's shape).
// "Spells with the chosen name can't be cast" is one named cast ban for each
// player (CastBanNamed); "lands with the chosen name can't be played" is a
// named land-play ban (ModCantPlayLandsNamed, game/land_play_named_ban.go).
// Both last until the caster's next turn begins (CR 611.2b). The name is free
// text (CR 201.2). The card is drawn as the spell resolves, after the name
// is asked for.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e6cbe03-6d6a-413e-9f3c-2770d408cc79",
		Name:         "Conjurer's Ban",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ctx.Game.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
				"Conjurer's Ban — choose a card name",
				ctx.ItemTail(conjurersBanNamed))
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}

var conjurersBanNamed = effects.CardTail("then/conjurers-ban-named", func(ctx *effects.Context, o game.TailOutcome) error {
	name := o.Tail.Word
	if name == "" {
		return nil
	}
	d := effects.DurationUntilYourNextTurn(ctx, ctx.Controller())
	for _, p := range effects.ApnapPlayers(ctx.Game) {
		if err := (effects.RestrictCasting{
			Player: p,
			Rule: game.CastBanRule{
				Kind:   game.CastBanNamed,
				Filter: game.PermissionFilter{Named: name},
			},
			Label:    "Conjurer's Ban — can't cast spells named " + name,
			Duration: d,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	ctx.Game.CantPlayLandsNamedForEffect(ctx.Source(), name, d, "Conjurer's Ban")
	return nil
})
