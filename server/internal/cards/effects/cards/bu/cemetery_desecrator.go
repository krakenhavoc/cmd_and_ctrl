package cardsbu

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cemetery Desecrator — Creature — Zombie {4}{B}{B}, 4/4 (Innistrad: Crimson
// Vow):
//
//	"Menace
//	 When this creature enters or dies, exile another card from a
//	 graveyard. When you do, choose one —
//	 • Remove X counters from target permanent, where X is the mana value
//	   of the exiled card.
//	 • Target creature an opponent controls gets -X/-X until end of turn,
//	   where X is the mana value of the exiled card."
//
// The exile is not targeted: the controller picks any card from any
// graveyard but the Desecrator itself (it is in one when it dies), as the
// trigger resolves. "When you do" is a reflexive trigger (CR 603.12) raised
// once the card has reached exile (nothing happens with every graveyard
// empty, or with the card kept out of exile by a replacement), carrying X
// as the mana value the exiled card has in exile. Its mode and its target
// are chosen as it goes on the stack, so the table can respond. Removing
// counters is exact, not "up to": the controller picks which kinds of
// counter and how many of each, X in all, or every counter the permanent has
// when it has fewer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "85e9a43e-5676-4d12-ac27-e8cf57307d5e",
		Name:            "Cemetery Desecrator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrDies(cemeteryDesecratorLabel, func(g *game.Game, item *game.StackItem) error {
				effects.VowB16QueueExile(g, item, cemeteryDesecratorLabel, item.SourceCardID, cemeteryDesecratorPicked)
				return nil
			}),
		},
	})
}

const cemeteryDesecratorLabel = "Cemetery Desecrator — exile another card from a graveyard"

var cemeteryDesecratorPicked = game.RegisterChooseCardsThen("choose-cards/cemetery-desecrator-exile", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || r.Asked == nil {
		return nil
	}
	ctx := effects.NewContext(g, r.Asked)
	return effects.ExileTarget{Target: r.Picked[0], ThenTail: ctx.ItemTail(cemeteryDesecratorExiled)}.Apply(ctx)
})

// cemeteryDesecratorExiled raises the reflexive trigger with X, the exiled
// card's mana value. An on-disk key: never renamed, never reused.
var cemeteryDesecratorExiled = effects.CardTail("cemetery-desecrator/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	return effects.ReflexiveTrigger{
		Label:  "Cemetery Desecrator — choose one",
		Body:   cemeteryDesecratorBody,
		Params: game.EffectParams{Amount: c.ManaValue()},
	}.Apply(ctx)
})

var cemeteryDesecratorBody = game.ReflexiveModalBody("cemetery-desecrator/modes", func(_ uuid.UUID, _ game.EffectParams) *game.ModeSpec {
	return effects.ChooseOne(
		effects.ModeDoing("Remove X counters from target permanent, where X is the mana value of the exiled card.",
			effects.TargetPermanent("target permanent"),
			func(item *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetCard || item.Params.Amount <= 0 {
					return nil
				}
				return cemeteryDesecratorRemoves.ChooseExactly(ctx, t.ID, nil, item.Params.Amount,
					"Cemetery Desecrator — remove counters")
			}),
		effects.ModeDoing("Target creature an opponent controls gets -X/-X until end of turn, where X is the mana value of the exiled card.",
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			func(item *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetCard || item.Params.Amount <= 0 {
					return nil
				}
				x := item.Params.Amount
				return effects.BoostUntilEOT{Target: t.ID, Power: -x, Toughness: -x,
					Label: "Cemetery Desecrator — -X/-X until end of turn"}.Apply(ctx)
			}),
	)
})

// cemeteryDesecratorRemoves takes the counters the controller chose off the
// permanent. A bot removes -1/-1 counters from its own permanent and
// everything else from another's.
var cemeteryDesecratorRemoves = effects.RegisterCounterChoice("number/cemetery-desecrator-remove-counters",
	func(ctx *effects.Context, target uuid.UUID, kind string, most int) int {
		mine := false
		if c, ok := ctx.Game.LookupCardForEffect(target); ok {
			mine = c.Controller == ctx.Controller()
		}
		if (kind == "-1/-1") == mine {
			return most
		}
		return 0
	},
	func(ctx *effects.Context, target uuid.UUID, _ []uuid.UUID, picked map[string]int) error {
		kinds := make([]string, 0, len(picked))
		for kind := range picked {
			kinds = append(kinds, kind)
		}
		sort.Strings(kinds)
		for _, kind := range kinds {
			if err := (effects.AddCounter{Target: target, Kind: kind, N: -picked[kind]}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
