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
// ONE CLAUSE IS OMITTED AND IT IS THE FAMOUS ONE. "You gain
// protection from everything until your next turn" has no shape:
// protection tests its quality against the SOURCE of a spell or
// ability, and the targeting choke point never receives one — ADR
// 0038 §7 sets out why protection is deliberately outside the
// keyword table and what it would cost to add. So the enters trigger
// is not registered at all rather than registered as a trigger that
// announces a shield nobody honours: a badge promising a rule
// nothing enforces is worse than a card that is plainly missing the
// clause. The Ring here is the draw engine and nothing else, which
// is weaker than printed in the only direction a simplification may
// go (#259).
//
// Two narrow consequences of the Ring leaving the battlefield mid-
// resolution, both weaker than printed and neither worth a caveat of
// its own: a Ring bounced in response to its tap ability puts no
// counter on and draws nothing, and one bounced in response to its
// upkeep trigger costs no life. CR 608.2 would use the counters it
// last had; the engine reads the card where it is.
func init() {
	Register(Spec{
		OracleID:     "3aa83ed2-f48b-4ce6-a614-2c54ddf50538",
		Name:         "The One Ring",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Casting The One Ring doesn't shield you — the \"protection from everything until your next turn\" clause isn't implemented, so you can still be attacked, targeted and burned the turn it lands.",
		},
		PrintedKeywords: []string{"indestructible"},
		Triggered: []game.TriggeredAbility{
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
func theOneRingTapDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	if err := (AddCounter{Target: item.SourceCardID, Kind: theOneRingBurden, N: 1}).Apply(ctx); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	return DrawCards{Player: item.Controller, N: c.Counters[theOneRingBurden]}.Apply(ctx)
}
