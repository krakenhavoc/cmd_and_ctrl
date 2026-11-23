package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trystan, Callous Cultivator // Trystan, Penitent Culler — {2}{G} Legendary
// Creature — Elf Druid 3/4 // Legendary Creature — Elf Warlock 3/4 (Lorwyn
// Eclipsed, #4458):
//
//	Front: "Deathtouch
//	        Whenever this creature enters or transforms into Trystan,
//	        Callous Cultivator, mill three cards. Then if there is an Elf
//	        card in your graveyard, you gain 2 life.
//	        At the beginning of your first main phase, you may pay {B}. If
//	        you do, transform Trystan."
//	Back:  "Deathtouch
//	        Whenever this creature transforms into Trystan, Penitent Culler,
//	        mill three cards, then you may exile an Elf card from your
//	        graveyard. If you do, each opponent loses 2 life.
//	        At the beginning of your first main phase, you may pay {G}. If
//	        you do, transform Trystan."
//
// The same shape as Ashling, Rekindled // Ashling, Rimebound: the back face
// is the "<oracle id>#1" row, the transform payment is a MayPay whose paid
// branch is TransformThis (ashlingTransformPaid, keyed, shared). The Elf
// card is looked for after the mill, as the graveyard stands then (the
// front face counts Trystan's own Elf card only if it is there).
func init() {
	const oracle = "961d1128-a022-480c-918e-16ba0abf844b"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Trystan, Callous Cultivator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventTransform}, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return (ev.Kind == game.EventETB && ev.CardID == source.InstanceID) || effects.ThisTransformedInto(0)(ev, source, lki, g)
			}, trystanCultivatorLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return g.MillToZoneKeyedForEffect(item.Controller, 3, game.ZoneGraveyard, game.MillUntil{}, ctx.ItemTail(trystanCultivatorMilled))
			}),
			effects.AtYourPrecombatMain("Trystan, Callous Cultivator — you may pay {B} to transform Trystan", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{B}",
					Question: "Trystan, Callous Cultivator — pay {B} to transform Trystan?",
					Then:     effects.AshlingTransformPaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Trystan, Penitent Culler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventTransform, effects.ThisTransformedInto(1), trystanCullerLabel, trystanCullerMill),
			effects.AtYourPrecombatMain("Trystan, Penitent Culler — you may pay {G} to transform Trystan", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{G}",
					Question: "Trystan, Penitent Culler — pay {G} to transform Trystan?",
					Then:     effects.AshlingTransformPaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

const (
	trystanCultivatorLabel = "Trystan, Callous Cultivator — mill three cards, then if there is an Elf card in your graveyard, you gain 2 life"
	trystanCullerLabel     = "Trystan, Penitent Culler — mill three cards, then you may exile an Elf card from your graveyard; if you do, each opponent loses 2 life"
)

// trystanCultivatorMilled is the front face's "then if there is an Elf card
// in your graveyard, you gain 2 life", run once the mill has settled. An
// on-disk key: never renamed, never reused.
var trystanCultivatorMilled = effects.CardTail("mill-then/trystan-callous-cultivator", func(ctx *effects.Context, _ game.TailOutcome) error {
	if !effects.DawnhandEulogistElfInGraveyard(ctx) {
		return nil
	}
	return effects.GainLife{Amount: 2}.Apply(ctx)
})

func trystanCullerMill(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return g.MillToZoneKeyedForEffect(item.Controller, 3, game.ZoneGraveyard, game.MillUntil{}, ctx.ItemTail(trystanCullerMilled))
}

// trystanCullerMilled is the back face's pick after the mill: you may exile an
// Elf card from your graveyard. An on-disk key: never renamed, never reused.
var trystanCullerMilled = effects.CardTail("mill-then/trystan-penitent-culler", func(ctx *effects.Context, _ game.TailOutcome) error {
	g, item := ctx.Game, ctx.Item
	elf := ctx.TextSubtype("Elf")
	ids := effects.B42CardsInGraveyardOf(g, item.Controller, func(c game.Card) bool { return c.HasSubtype(elf) })
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Trystan, Penitent Culler — you may exile an Elf card from your graveyard (each opponent loses 2 life)",
		Cards:    ids,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  trystanCullerExile,
		Asked:    game.NewAskedItem(item),
	})
	return nil
})

var trystanCullerExile = effects.AskedPick("choose-cards/trystan-penitent-culler-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.ExileTarget{Target: r.Picked[0], ThenTail: ctx.ItemTail(trystanCullerDrain)}.Apply(ctx)
})

// trystanCullerDrain is "if you do, each opponent loses 2 life": told the
// card only when it reached exile. An on-disk key: never renamed, never reused.
var trystanCullerDrain = effects.CardTail("exile-then/trystan-penitent-culler", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.EachOpponentLosesLife{N: 2}.Apply(ctx)
})
