package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Doom Reigns Supreme — Enchantment — Plan {1}{B} (Marvel Super Heroes,
// #4455):
//
//	"Whenever a Villain you control enters, each opponent loses 1 life and
//	 you gain 1 life. Put a plan counter on this enchantment.
//	 When the fifth plan counter is put on this enchantment, sacrifice it.
//	 When you do, target opponent exiles the top five cards of their
//	 library. You may cast up to two spells from among the exiled cards
//	 without paying their mana costs."
//
// Death to Our Enemies's plan counters and sacrifice-then-"when you do".
// The Villain trigger is Black Panther's entering-permanent test with the
// word read through the Enchantment's text; the drain and the counter are
// one footprinted do in printed order. The reflexive body targets an
// opponent as the trigger goes on the stack, exiles five cards from the top
// of that library as a run (a short library exiles what it has), then
// grants the controller ONE free, cast-only permission over the exiled
// cards with CastsLeft 2: two casts across all of them, spent a cast at a
// time, offered during the resolution (CR 608.2g). The offer stays open
// while a cast is left and is withdrawn, permission and all, when it is
// declined, so nothing can be cast after the ability finishes. Lands exiled
// are never cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f398a742-22c1-4da9-9d49-fdce2eb93d01",
		Name:         "Doom Reigns Supreme",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Villain"))
			}, "Doom Reigns Supreme — each opponent loses 1 life and you gain 1 life; put a plan counter on this enchantment",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1}, CounterOnThis{Kind: mshA3PlanCounter, N: 1})),
			On(game.EventCounterPlaced, mshA3PlanCounterReaches(5),
				"Doom Reigns Supreme — sacrifice it; when you do, target opponent exiles the top five cards of their library",
				b08OverlookSacrifice("Doom Reigns Supreme — target opponent exiles the top five cards of their library; you may cast up to two spells from among them", doomReignsSupremeBody)),
		},
	})
}

// doomReignsSupremeBody is the reflexive half, aimed at one opponent.
var doomReignsSupremeBody = game.ReflexiveBody("doom-reigns-supreme/exile-and-cast", simpleBody(mshA3DoomExile),
	constTargets(func() *game.TargetSpec { return TargetPlayer("target opponent", Opponent()) }))

// mshA3DoomExile exiles the top five cards of the target opponent's library
// and, once the run has settled, offers the casts.
func mshA3DoomExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := firstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	return MillToZone{
		Player:   target,
		N:        5,
		To:       game.ZoneExile,
		ThenTail: ctx.ItemTail(mshA3DoomExiled),
	}.Apply(ctx)
}

// mshA3DoomExiled grants the two free casts over the nonland cards exiled
// and offers them.
var mshA3DoomExiled = CardTail("msh-a3/doom-reigns-supreme-after-exile", func(ctx *Context, o game.TailOutcome) error {
	var cards []game.Card
	var ids []uuid.UUID
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		z := ctx.Game.FindCardZoneForEffect(id)
		if !ok || c.IsLand() || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		cards = append(cards, c)
		ids = append(ids, id)
	}
	if len(cards) == 0 {
		return nil
	}
	ctx.Game.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     ctx.Controller(),
		Zone:       game.ZoneExile,
		Duration:   ctx.Game.UntilEndOfTurnDuration(),
		Cost:       "{0}",
		Timing:     game.TimingFlash,
		CastOnly:   true,
		CastsLeft:  2,
		Source:     ctx.Source(),
		SourceName: "Doom Reigns Supreme",
		Label:      "Doom Reigns Supreme — cast it without paying its mana cost",
	}, cards)
	offerResolutionCastOver(ctx.Game, ctx.Controller(), ctx.Source(), "Doom Reigns Supreme",
		game.MayCastKeywordFreeCast, declineLeaveInExile, ids...)
	return nil
})
