package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        oracle,
		Name:            "Trystan, Callous Cultivator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventTransform}, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return (ev.Kind == game.EventETB && ev.CardID == source.InstanceID) || ThisTransformedInto(0)(ev, source, lki, g)
			}, trystanCultivatorLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (MillCards{N: 3}).Apply(ctx); err != nil {
					return err
				}
				if !dawnhandEulogistElfInGraveyard(ctx) {
					return nil
				}
				return GainLife{Amount: 2}.Apply(ctx)
			}),
			AtYourPrecombatMain("Trystan, Callous Cultivator — you may pay {B} to transform Trystan", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{B}",
					Question: "Trystan, Callous Cultivator — pay {B} to transform Trystan?",
					Then:     ashlingTransformPaid,
				}.Apply(NewContext(g, item))
			}),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Trystan, Penitent Culler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			On(game.EventTransform, ThisTransformedInto(1), trystanCullerLabel, trystanCullerMill),
			AtYourPrecombatMain("Trystan, Penitent Culler — you may pay {G} to transform Trystan", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{G}",
					Question: "Trystan, Penitent Culler — pay {G} to transform Trystan?",
					Then:     ashlingTransformPaid,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

const (
	trystanCultivatorLabel = "Trystan, Callous Cultivator — mill three cards, then if there is an Elf card in your graveyard, you gain 2 life"
	trystanCullerLabel     = "Trystan, Penitent Culler — mill three cards, then you may exile an Elf card from your graveyard; if you do, each opponent loses 2 life"
)

func trystanCullerMill(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (MillCards{N: 3}).Apply(ctx); err != nil {
		return err
	}
	elf := ctx.TextSubtype("Elf")
	ids := b42CardsInGraveyardOf(g, item.Controller, func(c game.Card) bool { return c.HasSubtype(elf) })
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
}

var trystanCullerExile = askedPick("choose-cards/trystan-penitent-culler-exile", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ExileTarget{Target: r.Picked[0], ThenTail: ctx.ItemTail(trystanCullerDrain)}.Apply(ctx)
})

// trystanCullerDrain is "if you do, each opponent loses 2 life": told the
// card only when it reached exile. An on-disk key: never renamed, never reused.
var trystanCullerDrain = CardTail("exile-then/trystan-penitent-culler", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return EachOpponentLosesLife{N: 2}.Apply(ctx)
})
