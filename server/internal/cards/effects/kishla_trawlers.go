package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kishla Trawlers — Creature — Human Citizen {2}{U}, 3/2 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, you may exile a creature card from your
//	 graveyard. When you do, return target instant or sorcery card from
//	 your graveyard to your hand."
//
// A CR 603.12 reflexive trigger. The "you may exile" is a pick of zero or
// one creature card from your graveyard, asked as the enters trigger
// resolves; the exile is a real one, and the reflexive trigger exists only
// once a card reached exile. It chooses its target as it goes on the stack
// (CR 603.3d), so the card exiled is never the card returned and every
// player may respond to it. With no legal instant or sorcery card the
// reflexive trigger is removed (CR 603.3d), and the creature stays exiled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "09277185-fccd-4747-aa79-f569c581022a",
		Name:         "Kishla Trawlers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kishla Trawlers — you may exile a creature card from your graveyard; when you do, return target instant or sorcery card from your graveyard to your hand",
				kishlaTrawlersOffer),
		},
	})
}

// kishlaTrawlersOffer asks for up to one creature card of the controller's
// graveyard. With none there is nothing to ask.
func kishlaTrawlersOffer(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	me := ctx.Controller()
	var cards []uuid.UUID
	if p := g.PlayerByIDForEffect(me); p != nil && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			if c.HasCardType("creature") {
				cards = append(cards, c.InstanceID)
			}
		}
	}
	if len(cards) == 0 {
		return nil
	}
	_, err := g.ChooseCardsRunThenKeyedForEffect([]game.ChooseCardsPrompt{{
		Chooser:    me,
		FromPlayer: me,
		Source:     ctx.Source(),
		Question:   "Kishla Trawlers — you may exile a creature card from your graveyard to return an instant or sorcery card",
		Cards:      cards,
		Min:        0,
		Max:        1,
		Zone:       game.ZoneGraveyard,
	}}, kishlaTrawlersExiled, ctx.Source(), ctx.RunData([]uuid.UUID{me}))
	return err
}

// kishlaTrawlersExiled exiles the card picked, if any. An on-disk key.
var kishlaTrawlersExiled = RunThen("prompt-run/kishla-trawlers-exile", func(ctx *Context, r game.RunDone) error {
	var chosen []uuid.UUID
	for _, player := range r.Carry {
		chosen = append(chosen, r.Picks().By(player)...)
	}
	if len(chosen) == 0 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(chosen, ctx.ItemTail(kishlaTrawlersAfterExile))
})

// kishlaTrawlersAfterExile makes the reflexive trigger once a card really
// reached exile (a replacement may have sent it elsewhere). An on-disk key.
var kishlaTrawlersAfterExile = CardTail("kishla-trawlers/after-exile", func(ctx *Context, o game.TailOutcome) error {
	exiled := false
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			exiled = true
		}
	}
	if !exiled {
		return nil
	}
	return WhenYouDo("Kishla Trawlers — return target instant or sorcery card from your graveyard to your hand",
		kishlaTrawlersReturnBody).Apply(ctx)
})

// kishlaTrawlersReturnBody is the reflexive trigger: Awaken the Honored
// Dead's return, for an instant or sorcery card. An on-disk key.
var kishlaTrawlersReturnBody = game.ReflexiveBody("kishla-trawlers/return-to-hand",
	simpleBody(returnFirstLegalGraveyardTargetToHand),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("target instant or sorcery card from your graveyard", YouOwn(), Or(Instant(), Sorcery()))
	}))
