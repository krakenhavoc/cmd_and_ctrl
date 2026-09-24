package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The One Ring — Legendary Artifact {4}:
//
//	"Indestructible
//	 When The One Ring enters, if you cast it, you gain protection
//	 from everything until your next turn.
//	 At the beginning of your upkeep, you lose 1 life for each burden
//	 counter on The One Ring.
//	 {T}: Put a burden counter on The One Ring, then draw a card for
//	 each burden counter on The One Ring."
//
// The engine of the card is the burden counter: every activation
// costs one more life a turn forever and draws one more card. That
// half is ordinary and complete here — a tap ability whose ORDER is
// the whole rules content (the counter goes on FIRST, so the first
// activation draws one card, not zero), and a mandatory upkeep
// trigger that reads the counters when it RESOLVES, so a burden
// added in response to it is paid for.
//
// Indestructible is a printed keyword the engine enforces (S25); the
// deck importer stamps it from Scryfall and the declaration here is
// the fallback for tokens, fixtures and the dev spawner.
//
// THE FAMOUS CLAUSE SHIPPED WITH #1197. "You gain protection from
// everything until your next turn" used to be omitted entirely,
// because protection tests its quality against the SOURCE of a spell
// or ability and a PLAYER had no ability slice to hold the answer.
// ADR 0072's 2026-09-22 amendment gave a player one, so the enters
// trigger is registered and the shield is real: nothing an opponent
// controls can target you, and every source of damage — combat and
// noncombat alike — is prevented until your next turn begins
// (CR 702.16e / 702.16i).
//
// "IF YOU CAST IT" is the intervening-if (CR 603.4), read off the
// event log the way Zacama and Tiamat read theirs
// (b16EnteredFromStack): a Ring reanimated, blinked or cheated onto
// the battlefield gives no shield, exactly as printed. Checked as the
// trigger would fire and not again at resolution, which for "if you
// cast it" can never change between the two.
//
// The duration is stamped when the TRIGGER resolves, not when the
// Ring entered — CR 611.2's window starts when the effect is created.
// A Ring whose trigger is countered or removed gives no shield at
// all, which falls out of the model rather than needing a rule.
//
// Two narrow consequences of the Ring leaving the battlefield mid-
// resolution, both weaker than printed and neither worth a caveat of
// its own: a Ring bounced in response to its tap ability puts no
// counter on and draws nothing, and one bounced in response to its
// upkeep trigger costs no life. CR 608.2 would use the counters it
// last had; the engine reads the card where it is.
func init() {
	Register(Spec{
		OracleID:        "3aa83ed2-f48b-4ce6-a614-2c54ddf50538",
		Name:            "The One Ring",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && b16EnteredFromStack(g, source.InstanceID)
			}, "The One Ring — protection from everything until your next turn", theOneRingShield),
			AtYourUpkeep("The One Ring — lose 1 life for each burden counter", theOneRingUpkeepBurden),
		},
		Activated: []ActivatedAbility{{
			Label:  "{T}: Put a burden counter on The One Ring, then draw a card for each burden counter on The One Ring.",
			Cost:   TapCost(),
			Effect: theOneRingTapDraw,
		}},
	})
}

// theOneRingBurden is the counter the card names.
const theOneRingBurden = "burden"

// theOneRingShield is "you gain protection from everything until your
// next turn" (#1197). It reads the trigger's CONTROLLER rather than
// the Ring's current controller: the trigger is controlled by whoever
// controlled the Ring as it entered (CR 603.3a), and a Ring stolen in
// response shields the player whose trigger it is.
//
// No battlefield check. CR 608.2 resolves an ability whether or not
// its source is still there, and the shield is about the player, not
// about the artifact — a Ring bounced with the trigger on the stack
// still protects you, as printed.
func theOneRingShield(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return GainPlayerKeyword{
		Player:   item.Controller,
		Keyword:  ProtectionFromEverything,
		Label:    "The One Ring — protection from everything",
		Duration: DurationUntilYourNextTurn(ctx, item.Controller),
	}.Apply(ctx)
}

// theOneRingUpkeepBurden is the upkeep drain, counted at resolution.
func theOneRingUpkeepBurden(g *game.Game, item *game.StackItem) error {
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	n := c.Counters[theOneRingBurden]
	if n <= 0 {
		return nil
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -n)
}

// theOneRingTapDraw is "put a burden counter on it, THEN draw a card
// for each burden counter on it" — the counter first, so the draw is
// for the new total. Reversing the two would cost a card on every
// activation.
//
// #1290: the draw reads the count AFTER the counter LANDS, via
// AddCounterThenForEffect's continuation, not on the next line — a
// Doubling Season / Hardened Scales board pauses the placement on a
// CR 616 prompt, and reading before it resumes would draw for the
// pre-placement count.
func theOneRingTapDraw(g *game.Game, item *game.StackItem) error {
	if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) { // #1432
		return nil
	}
	return g.AddCounterThenForEffect(item.SourceCardID, theOneRingBurden, 1, func(g *game.Game, _ int) error {
		c, ok := g.LookupCardForEffect(item.SourceCardID)
		if !ok {
			return nil
		}
		return DrawCards{Player: item.Controller, N: c.Counters[theOneRingBurden]}.Apply(NewContext(g, item))
	})
}
