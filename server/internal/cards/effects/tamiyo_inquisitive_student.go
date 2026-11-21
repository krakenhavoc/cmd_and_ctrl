package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tamiyo, Inquisitive Student // Tamiyo, Seasoned Scholar — Modern Horizons 3
// (#5948), a transforming legendary creature whose back face is a planeswalker.
//
// Front face, Legendary Creature — Moonfolk Wizard {U}, 0/3:
//
//	"Flying
//	 Whenever Tamiyo attacks, investigate.
//	 When you draw your third card in a turn, exile Tamiyo, then return her
//	 to the battlefield transformed under her owner's control."
//
// Back face, Legendary Planeswalker — Tamiyo, loyalty 2:
//
//	"+2: Until your next turn, whenever a creature attacks you or a
//	     planeswalker you control, it gets -1/-0 until end of turn.
//	 −3: Return target instant or sorcery card from your graveyard to your
//	     hand. If it's a green card, add one mana of any color.
//	 −7: Draw cards equal to half the number of cards in your library,
//	     rounded up. You get an emblem with "You have no maximum hand size.""
//
// The back face registers under "<oracle_id>#1" (game.CatalogKey). The flip
// is Astrologian's Planisphere's tally (the controller's third draw of the
// turn, read off the per-turn draw count as the card is drawn) feeding
// ExileAndReturnTransformed, a new object entering with its printed loyalty
// (CR 306.5b). The +2 is Jace, Reality Sculptor's −3 delayed trigger on
// attack declarations against you or a planeswalker you control, with -1/-0.
// The −3's mana is added only if the card was green as it was returned
// (CR 608.2h on its colors); the emblem is Wrenn and Seven's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        tamiyoInquisitiveStudentOracle,
		Name:            "Tamiyo, Inquisitive Student",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Tamiyo, Inquisitive Student — investigate", Do(Investigate{N: 1})),
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && g.TurnTallyFor(source.Controller).CardsDrawn == 3
			}, "Tamiyo, Inquisitive Student — exile her, then return her transformed",
				func(g *game.Game, item *game.StackItem) error {
					return ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
	})

	Register(Spec{
		OracleID:        tamiyoInquisitiveStudentOracle + "#1",
		Name:            "Tamiyo, Seasoned Scholar",
		Completeness:    CompletenessFull,
		StartingLoyalty: 2,
		Emblem: &EmblemSpec{
			Label:    "Tamiyo, Seasoned Scholar emblem",
			Text:     "You have no maximum hand size.",
			HandSize: []game.HandSizeStatic{{Kind: game.HandSizeNoMaximum}},
		},
		Activated: []ActivatedAbility{
			{
				Label:  "+2: Until your next turn, whenever a creature attacks you or a planeswalker you control, it gets -1/-0 until end of turn.",
				Cost:   LoyaltyCost(2),
				Effect: tamiyoSeasonedScholarPlusTwo,
			},
			{
				Label:   "−3: Return target instant or sorcery card from your graveyard to your hand. If it's a green card, add one mana of any color.",
				Cost:    LoyaltyCost(-3),
				Targets: TargetCardInGraveyard("target instant or sorcery card from your graveyard", Or(Instant(), Sorcery()), YouOwn()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					card, found := g.LookupCardForEffect(t.ID)
					if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
						return err
					}
					if found && slices.Contains(card.EffectiveColors(), "G") {
						return AddMana{Produced: "{W|U|B|R|G}"}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label: "−7: Draw cards equal to half the number of cards in your library, rounded up. You get an emblem with \"You have no maximum hand size.\"",
				Cost:  LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					p := g.PlayerByIDForEffect(ctx.Controller())
					if p == nil {
						return nil
					}
					if err := (DrawCards{N: (p.Library.Size() + 1) / 2}).Apply(ctx); err != nil {
						return err
					}
					return CreateEmblem{}.Apply(ctx)
				},
			},
		},
	})
}

// tamiyoInquisitiveStudentOracle is shared by both faces.
const tamiyoInquisitiveStudentOracle = "1ed9e351-0000-4b75-8ecb-d15bd88df68a"

// tamiyoSeasonedScholarPlusTwo schedules the "until your next turn, whenever"
// trigger (CR 603.7b), Jace, Reality Sculptor's shape with a -1/-0 body. Its
// condition is the shared "attacks you or a planeswalker you control".
func tamiyoSeasonedScholarPlusTwo(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	controller := ctx.Controller()
	until := DurationUntilYourNextTurn(ctx, controller)
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   controller,
		SourceCardID: ctx.Source(),
		Label:        "Tamiyo, Seasoned Scholar — the attacking creature gets -1/-0 until end of turn",
		On:           []game.EventKind{game.EventAttack},
		Condition:    jaceRealitySculptorAttacksYouCondition,
		Body:         tamiyoSeasonedScholarShrinkBody,
		Repeats:      true,
		Duration:     &until,
	})
	return nil
}

// tamiyoSeasonedScholarShrinkBody gives the attacking creature -1/-0 until end
// of turn; one that has left the battlefield by then gets nothing.
var tamiyoSeasonedScholarShrinkBody = game.SimpleDelayedBody("tamiyo-seasoned-scholar/attacker-gets-minus-one", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range ctx.PayloadCards() {
		if _, ok := g.PermanentRefForEffect(id); !ok {
			continue
		}
		if err := (BoostUntilEOT{Target: id, Power: -1, Label: "Tamiyo, Seasoned Scholar — -1/-0"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
