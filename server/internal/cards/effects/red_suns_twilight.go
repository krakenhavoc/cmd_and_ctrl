package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Red Sun's Twilight — Sorcery {X}{R}{R}:
//
//	"Destroy up to X target artifacts. If X is 5 or more, for each artifact
//	 destroyed this way, create a token that's a copy of it. Those tokens
//	 gain haste. Exile them at the beginning of the next end step."
//
// "Up to X" is the announced X as the ceiling (CountFromX with UpToX). An
// artifact counts as destroyed this way only if it landed in a graveyard
// (CR 701.8a), so an indestructible or regenerated one is not copied. The
// copies are made together (CR 608.2f) from each artifact as it last existed
// on the battlefield (CR 707.2, CR 608.2h); a token that was destroyed has no
// card to copy but its copiable values still make a token. The tail then
// gives each copy haste and schedules ONE delayed exile for the next end step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "21c77cdb-9ba1-4621-b096-7eb94c076cef",
		Name:         "Red Sun's Twilight",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      redSunsTwilightTargets(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var destroyed []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if z := ctx.Game.FindCardZoneForEffect(t.ID); z != nil && (z.Kind == game.ZoneGraveyard || z.Kind == game.ZoneCommand) {
					destroyed = append(destroyed, t.ID)
				}
			}
			if ctx.X() < 5 || len(destroyed) == 0 {
				return nil
			}
			return CreateTokenCopiesTogether{
				Controller: item.Controller,
				Copies:     destroyed,
				Then:       ctx.ItemTail(redSunsTwilightRest),
			}.Apply(ctx)
		},
	})
}

// redSunsTwilightTargets is "up to X target artifacts".
func redSunsTwilightTargets() *game.TargetSpec {
	spec := TargetPermanent("up to X target artifacts", Artifact())
	spec.CountFromX = true
	spec.UpToX = true
	return spec
}

// redSunsTwilightRest is "Those tokens gain haste. Exile them at the
// beginning of the next end step." An on-disk key: never renamed.
var redSunsTwilightRest = CardTail("red-suns-twilight/haste-and-exile-copies", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		if err := (ScopedEffectFor{Target: id, Mods: []game.Mod{game.AddKeywordsMod("haste")},
			Duration: game.IndefiniteDuration(), Label: "Red Sun's Twilight — haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	return ScheduleDelayedTrigger{
		Label: "Red Sun's Twilight — exile the tokens",
		Cards: o.IDs,
		Body:  exileListedCardsBody,
	}.Apply(ctx)
})
