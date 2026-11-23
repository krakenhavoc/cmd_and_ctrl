package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyfisher Spider — Creature — Spider {2}{B}{G}, 3/3 (Bloomburrow
// Commander, #5368):
//
//	"Reach
//	 When this creature enters, you may sacrifice another creature. When you
//	 do, destroy target nonland permanent.
//	 When this creature dies, you may gain 1 life for each creature card in
//	 your graveyard. If you do, exile this card from your graveyard."
//
// The enters trigger is Grist's −2 shape: a "you may sacrifice" over your
// other creatures (the source excluded by identity, so a second Spider is on
// offer), and only a creature that really went raises the reflexive trigger,
// which targets as it is put on the stack (CR 603.12) — no legal target and it
// is removed (CR 603.3d), though the creature is already gone.
//
// The dies trigger asks "you may" as it resolves, counts the creature cards in
// your graveyard then (this Spider is one of them while it is there), gains
// that much life, and — "if you do" — exiles the Spider's card from the
// graveyard if it is still there; a Spider reanimated in response is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "67158085-4807-4f0e-8599-c55ec7ec252e",
		Name:            "Skyfisher Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Skyfisher Spider — you may sacrifice another creature; when you do, destroy target nonland permanent",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Skyfisher Spider — sacrifice another creature to destroy target nonland permanent?",
						effects.Creature(), true, skyfisherSpiderSacrificed)
				}),
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.CardDied(ev, source)
				},
				OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Skyfisher Spider — gain 1 life for each creature card in your graveyard and exile it?"},
				Key:            "Skyfisher Spider — you may gain 1 life for each creature card in your graveyard; if you do, exile this card",
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.GainLife{Amount: effects.B11CreatureCardsInGraveyard(g, item.Controller)}).Apply(ctx); err != nil {
						return err
					}
					if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					return effects.ExileTarget{Target: item.SourceCardID}.Apply(ctx)
				},
			},
		},
	})
}

// skyfisherSpiderSacrificed raises the reflexive "destroy target nonland
// permanent" once a creature has really gone. An on-disk key: never renamed,
// never reused.
var skyfisherSpiderSacrificed = effects.RegisterSacrificedOne("prompt-run/skyfisher-spider-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.ReflexiveTrigger{
		Label: "Skyfisher Spider — destroy target nonland permanent",
		Body:  skyfisherSpiderDestroyBody,
	}.Apply(ctx)
})

// skyfisherSpiderDestroyBody is the reflexive trigger's registration, with its
// own target clause.
var skyfisherSpiderDestroyBody = game.ReflexiveBody("skyfisher-spider/destroy",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		return effects.DestroyTheTargetPermanent(item, effects.NewContext(g, item))
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetPermanent("target nonland permanent", effects.Nonland())
	}))
