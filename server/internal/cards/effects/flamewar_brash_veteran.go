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
			Effect: flamewarExileIntel,
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

// flamewarExileIntel is the back face's trigger: exile that many cards
// face down, an intel counter on each, then convert.
func flamewarExileIntel(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var exiled []uuid.UUID
	if err := (ExileTopFaceDown{Player: item.Controller, N: item.Params.Amount, Exiled: &exiled}).Apply(ctx); err != nil {
		return err
	}
	for _, id := range exiled {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if err := g.AddCounterForEffect(id, flamewarIntelCounter, 1); err != nil {
			return err
		}
	}
	return ConvertThis{}.Apply(ctx)
}
