package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ethereal Absolution — Enchantment {4}{W}{B} (Ravnica Allegiance):
//
//	"Creatures you control get +1/+1.
//	 Creatures your opponents control get -1/-1.
//	 {2}{W}{B}: Exile target card from an opponent's graveyard. If it was a
//	 creature card, you create a 1/1 white and black Spirit creature token with
//	 flying."
//
// Two layer 7c statics; the activation is Scavenging Ooze's exile-then-if-it-was
// shape, so the Spirit is made only when the card reached exile and was a
// creature card as it sat in the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6f95e3a-61a6-40a9-8070-ffd2767764ad",
		Name:         "Ethereal Absolution",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.CreaturesYouControl, 1, 1),
			effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller != source.Controller
			}, -1, -1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}{B}: Exile target card from an opponent's graveyard. If it was a creature card, you create a 1/1 white and black Spirit creature token with flying.",
			Cost:    effects.ManaCost("{2}{W}{B}"),
			Targets: effects.TargetCardInGraveyard("target card from an opponent's graveyard", effects.Not(effects.YouOwn())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.FirstLegalCardTarget(ctx)
				if id == uuid.Nil {
					return nil
				}
				return effects.ExileThenIfItWas{Target: id, Was: effects.WasCreatureCard, Keyed: etherealAbsolutionSpirit}.Apply(ctx)
			},
		}},
	})
}

// etherealAbsolutionSpirit is the Spirit, made only when the card reached exile
// and was a creature card. An on-disk key: never renamed, never reused.
var etherealAbsolutionSpirit = effects.IfItWasTail("if-it-was/ethereal-absolution-spirit", func(ctx *effects.Context) error {
	return effects.CreateToken{Template: effects.TokenCard("1/1 white and black Spirit with flying"), N: 1}.Apply(ctx)
}, nil)
