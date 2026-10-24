package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flamewar, Brash Veteran // Flamewar, Streetwise Operative (#3195, ADR
// 0152) — a More Than Meets the Eye card.
//
// Front face, Legendary Artifact Creature — Robot {1}{B}{R}, 3/2:
//
//	"More Than Meets the Eye {B}{R}
//	 Sacrifice another artifact: Put a +1/+1 counter on Flamewar and
//	 convert it. Activate only as a sorcery.
//	 {1}, Discard your hand: Put all exiled cards you own with intel
//	 counters on them into your hand."
//
// Back face, Legendary Artifact — Vehicle, 2/1:
//
//	"Living metal
//	 Menace, deathtouch
//	 Whenever Flamewar deals combat damage to a player, exile that many
//	 cards from the top of your library face down. Put an intel counter
//	 on each of them. Convert Flamewar."
//
// Intel is an ordinary counter on the exiled card (CR 122.1), the way
// Tinybones' stash is: the exile is face down, so only you may look at
// those cards, and the counter is cleared when a card leaves exile. The
// second ability takes every card in exile that you own and that holds
// an intel counter, whoever exiled it (a copy of Flamewar controlled by
// another player puts counters on THEIR own cards), into your hand.
// "Discard your hand" is a cost, so the hand is gone before the cards
// arrive. The counters go on whatever was exiled even if some exiling
// replacement redirected it.
//
// No simplification.
const flamewarOracleID = "cf812f1e-f48a-4375-8d2c-f6e3cb82e067"

const flamewarIntelCounter = "intel"

func init() {
	Register(Spec{
		OracleID:         flamewarOracleID,
		Name:             "Flamewar, Brash Veteran",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{MoreThanMeetsTheEye("{B}{R}")},
		Activated: []ActivatedAbility{
			{
				Label:        "Sacrifice another artifact: Put a +1/+1 counter on Flamewar and convert it. Activate only as a sorcery.",
				Cost:         SacrificeAnotherN(1, "another artifact", Artifact()),
				SorcerySpeed: true,
				// The conversion waits for a prompt the counter asks (a CR 616 order; CR 608.2c, #4155).
				Effect: Do(
					Instruction(func(ctx *Context) error {
						item := ctx.Item
						if onBattlefield(ctx.Game, item.SourceCardID) && !sourceIsNewObject(ctx.Game, item) {
							return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
						}
						return nil
					}),
					ConvertThis{},
				),
			},
			{
				Label:   "{1}, Discard your hand: Put all exiled cards you own with intel counters on them into your hand.",
				Cost:    Plus(ManaCost("{1}"), DiscardYourHand()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  flamewarTakeIntel,
			},
		},
	})
	Register(Spec{
		OracleID:        flamewarOracleID + "#1",
		Name:            "Flamewar, Streetwise Operative",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "deathtouch"},
		Static:          []game.StaticAbility{LivingMetal()},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return ThisDealtCombatDamageToAPlayer(ev, source, ch, g)
			},
			Key: "Flamewar, Streetwise Operative — exile that many cards face down with intel counters, then convert",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Flamewar, Streetwise Operative — exile that many cards face down with intel counters, then convert")
				item.Params.Amount = ev.Amount
				return item
			},
			// The counters and the conversion wait for a prompt the exile asks (a CR 614
			// replacement on a commander, CR 903.9; CR 608.2c, #4228).
			Effect: Do(
				Instruction(flamewarExileIntel),
				Instruction(flamewarMarkIntel),
				ConvertThis{},
			),
		}},
	})
}

// flamewarTakeIntel puts every exiled card the controller owns that
// holds an intel counter into their hand.
func flamewarTakeIntel(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, c := range g.Exile.Cards {
		if c.Owner == item.Controller && c.Counters[flamewarIntelCounter] > 0 {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := g.BounceToHandForEffect(id); err != nil {
			return err
		}
	}
	return nil
}

// flamewarIntelRecord files the event log's cursor from before the exile, for
// the step that marks the cards (#4228).
const flamewarIntelRecord = "flamewar/exile-cursor"

// flamewarExileIntel is the back face's trigger, step one: file the log's
// cursor, then exile that many cards face down.
func flamewarExileIntel(ctx *Context) error {
	ctx.RecordStep(flamewarIntelRecord, nil, int(b25LastEventSeq(ctx.Game)))
	return ExileTopFaceDown{Player: ctx.Controller(), N: ctx.Item.Params.Amount}.Apply(ctx)
}

// flamewarMarkIntel is step two: an intel counter on each card that left the
// controller's library for exile since the cursor and is still there. It reads
// the cards off the event log rather than from step one's result, so a
// commander the owner chose to leave in exile (CR 903.9, a prompt the exile
// asked) is marked too, once the answer is in.
func flamewarMarkIntel(ctx *Context) error {
	r := ctx.StepRecord(flamewarIntelRecord)
	if len(r.Ns) == 0 {
		return nil
	}
	var ids []uuid.UUID
	for i := len(ctx.Game.Events) - 1; i >= 0; i-- {
		ev := ctx.Game.Events[i]
		if ev.Seq <= uint64(r.Ns[0]) {
			break
		}
		if ev.Kind == game.EventZoneMove && ev.Actor == ctx.Controller() && ev.OldZone == game.ZoneLibrary && ev.NewZone == game.ZoneExile {
			ids = append(ids, ev.CardID)
		}
	}
	for _, id := range ids {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if err := ctx.Game.AddCounterForEffect(id, flamewarIntelCounter, 1); err != nil {
			return err
		}
	}
	return nil
}
