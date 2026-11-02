package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hela, Mistress of Darkness // Dark Tutor — Marvel's Spider-Man
// Legendary Creature — God Sorcerer Villain {4}{B}{B}, 6/6, with a
// Sorcery — Vengeance {1}{B}{B} half (#4460, layout adventure):
//
//	Hela: "Whenever a player loses life during your turn, you may create
//	 that many tapped 2/2 black Zombie creature tokens. Do this only once
//	 each turn."
//	Dark Tutor: "Vengeance (Cast only from your graveyard. Then exile it.)
//	 Search your library for a card, put it into your hand or graveyard,
//	 then shuffle. You lose 3 life."
//
// # Vengeance
//
// Vengeance is not in the pinned Comprehensive Rules (2026-09-25), so no
// rule number is cited for it: the engine implements exactly the reminder
// text. The half is cast from a graveyard and from nowhere else, and is
// exiled as it leaves the stack, however it leaves. That is aftermath's
// zone rule and exile (CR 702.127a) under a different name, so
// game.isVengeanceHalf rides aftermathZoneRule and the AltCostExiles
// stamp; the layout is "adventure" but the half is not an Adventure (CR
// 715), so the creature is never castable from the exile it lands in
// (castAsAdventure refuses it). Hela herself is an ordinary hand cast.
//
// # Hela
//
// "Loses life" is two event kinds, a negative EventChangeLife and the life
// an EventDealDamage cost (the b04OpponentLostLife pair), and unlike the
// opponents-only helpers it counts every player, Hela's controller
// included, but only on Hela's controller's turn. The engine emits one
// event per damage source, so the trigger waits for the batch to finish
// (AtBatchEnd, ThatMuchLifeLost) and "that many" is the TOTAL the player
// lost at the same time, as printed (CR 603.2c). Each player who lost life
// is their own occurrence.
//
// "You may" is the CR 603.5 prompt asked as the trigger resolves, and
// "do this only once each turn" limits the DOING (b28DoneThisTurn), as
// Baron Strucker's does: a declined offer leaves the turn's one use for a
// later loss, and the second of two triggers stacked before either
// resolved does nothing. The tally is per Hela object.
//
// No simplification.
const (
	helaOracleID = "203f6f4a-1cf5-4ddc-b60f-676832eeff6b"
	helaLabel    = "Hela, Mistress of Darkness — create that many tapped 2/2 black Zombies"
	helaDone     = "Hela, Mistress of Darkness — created Zombies this turn"
)

func init() {
	Register(Spec{
		OracleID:     helaOracleID,
		Name:         "Hela, Mistress of Darkness",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventChangeLife, game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return helaLifeLost(ev, g) > 0 && IsYourTurn(g, source.Controller) &&
						!b28DoneThisTurn(g, source.InstanceID, helaDone)
				},
				Key:        helaLabel,
				AtBatchEnd: ThatMuchLifeLost(),
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, helaLabel)
				},
				Effect: helaZombies,
			}, "Hela, Mistress of Darkness — create that many tapped 2/2 black Zombie creature tokens?"),
		},
	})
	Register(Spec{
		OracleID:     helaOracleID + "#1",
		Name:         "Dark Tutor",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:   ctx.Controller(),
				Dest:     game.ZoneLibrary,
				ToTop:    true,
				Limit:    1,
				Shuffle:  false,
				Reason:   "Dark Tutor — search your library for a card",
				ThenTail: ctx.Tail(darkTutorFound),
			}.Apply(ctx)
		},
	})
}

// helaLifeLost is the life one event cost a seated player: a negative
// EventChangeLife, or the life an EventDealDamage to a player cost
// (infect and a life total that can't change cost none). Zero otherwise.
func helaLifeLost(ev game.Event, g *game.Game) int {
	if ev.Target == uuid.Nil || g.PlayerByIDForEffect(ev.Target) == nil {
		return 0
	}
	switch ev.Kind {
	case game.EventChangeLife:
		return max(-ev.Amount, 0)
	case game.EventDealDamage:
		return ev.DamageLifeLoss()
	}
	return 0
}

// helaZombies is the resolved "you may create that many tapped 2/2
// Zombies": the TOTAL the player lost (the settled event's Amount), unless
// the turn's one use was spent while this trigger waited.
func helaZombies(g *game.Game, item *game.StackItem) error {
	if b28DoneThisTurn(g, item.SourceCardID, helaDone) {
		return nil
	}
	n := settledEventAmount(item)
	b28NoteDoneThisTurn(g, item.SourceCardID, helaDone)
	_, err := g.CreateTokensForEffect(item.Controller, NewContext(g, item).TextToken(TokenCard("2/2 black Zombie")), n,
		game.TokenEntryOptions{Tapped: true})
	return err
}

// darkTutorFound is the search's continuation. The found card is still in
// the library, on top, so the choice of where it goes (hand or graveyard)
// is made with it in view. Nothing found: the shuffle and the life loss
// still happen. An on-disk key.
var darkTutorFound = CardTail("dark-tutor/found", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return darkTutorFinish(ctx)
	}
	return PickOption{
		Question: "Dark Tutor — put the card into your hand or your graveyard?",
		Options:  []game.ChoiceOption{{Label: "Hand"}, {Label: "Graveyard"}},
		ThenKey:  darkTutorPlaced,
		Carry:    o.IDs,
	}.Apply(ctx)
})

// darkTutorPlaced puts the found card where the player chose, then
// finishes. A card that is no longer where the search left it is skipped.
// An on-disk key.
var darkTutorPlaced = OptionPickThen("option-pick/dark-tutor-placed", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || ctx.Game.FindCardZoneForEffect(r.Carry[0]) == nil {
		return darkTutorFinish(ctx)
	}
	if r.Index == 1 {
		if err := ctx.Game.PutIntoGraveyardForEffect(r.Carry[0]); err != nil {
			return err
		}
		return darkTutorFinish(ctx)
	}
	return ctx.Game.BounceCardsToHandKeyedForEffect(liveCards(ctx.Game, r.Carry[:1]), ctx.Tail(darkTutorAfter))
})

// darkTutorAfter is the tail of the move to hand.
var darkTutorAfter = CardTail("dark-tutor/after-hand", func(ctx *Context, _ game.TailOutcome) error {
	return darkTutorFinish(ctx)
})

// darkTutorFinish is "then shuffle. You lose 3 life."
func darkTutorFinish(ctx *Context) error {
	if err := ctx.Game.ShuffleLibraryForEffect(ctx.Controller()); err != nil {
		return err
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -3)
}
