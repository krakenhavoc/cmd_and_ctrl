package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Roving Actuator — Artifact Creature — Robot {3}{R}, 3/4 (Edge of
// Eternities):
//
//	"Void — When this creature enters, if a nonland permanent left the
//	 battlefield this turn or a spell was warped this turn, exile up to
//	 one target instant or sorcery card with mana value 2 or less from
//	 your graveyard. Copy it. You may cast the copy without paying its
//	 mana cost."
//
// Void is an intervening if (void.go). "Up to one target" may be none, in
// which case the ability does nothing. The exile goes through the CR 614
// window and only a card that really reached exile is copied
// (CopyAndCast, ADR 0066's 2026-10-13 card-copy amendment); the copy is
// offered for free and is cast, or not, while the ability is still
// resolving (CR 707.12, CR 608.2g). A copy not cast ceases to exist and the
// original stays in exile.
//
// No simplification.
func init() {
	enters := effects.WhenThisEntersIfVoid("Roving Actuator — exile up to one instant or sorcery card from your graveyard, copy it, and you may cast the copy",
		func(g *game.Game, item *game.StackItem) error {
			ids := effects.LegalTargetIDs(effects.NewContext(g, item))
			if len(ids) == 0 {
				return nil
			}
			return g.ExileCardsKeyedForEffect(ids, effects.NewContext(g, item).ItemTail(rovingActuatorCopy))
		})
	enters.Targets = effects.TargetCardInGraveyard("up to one target instant or sorcery card with mana value 2 or less in your graveyard",
		effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery()), effects.ManaValueLE(2)).WithCount(0, 1)
	effects.Register(effects.Spec{
		OracleID:     "78ac1e6c-2704-4844-9e87-642324c23997",
		Name:         "Roving Actuator",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}

// rovingActuatorCopy copies what reached exile (#4253 slice 9).
var rovingActuatorCopy = effects.CardTail("roving-actuator/copy-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	g, item, exiled := ctx.Game, ctx.Item, o.IDs

	return effects.CopyAndCast{Cards: exiled, Free: true}.Apply(effects.NewContext(g, item))
})
