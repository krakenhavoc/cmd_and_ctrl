package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Celes, Rune Knight — Legendary Creature — Human Wizard Knight
// {1}{R}{W}{B}, 4/4 (deck #3655, seam #3770):
//
//	"When Celes enters, discard any number of cards, then draw that
//	 many cards plus one.
//	 Whenever one or more other creatures you control enter, if one or
//	 more of them entered from a graveyard or was cast from a
//	 graveyard, put a +1/+1 counter on each creature you control."
//
// The enters trigger is Brass's Tunnel-Grinder's loot: a prompted
// discard run with the draw as its continuation, so "that many" is the
// number really discarded.
//
// The second ability is a batch trigger with an intervening if that is
// about the batch's members, not the first one: "one or more of them"
// came from a graveyard. The condition lives in AppliesTo and the
// OncePerBatch slot is spent only by an event that passes it
// (harvestMatchLocked judges AppliesTo first), so a batch whose first
// creature was cast from hand and whose second was reanimated still
// triggers, once — and a batch of three reanimated creatures triggers
// once, not three times (CR 603.2c).
//
// Two readings make up "came from a graveyard or was cast from a
// graveyard", and they are different facts:
//
//   - "Entered from a graveyard" is EventETB.EnteredFrom, the zone the
//     object moved out of (undying, persist, reanimation, a blink that
//     returned it from the graveyard).
//   - "Was cast from a graveyard" is Card.Provenance.FromZone, the zone
//     the spell that became THIS object was cast from (CR 400.7d). A
//     flashback or escape creature enters from the STACK, so the first
//     reading misses it. Provenance is stamped as the spell becomes the
//     permanent, before EventETB fires, and is cleared when the
//     permanent leaves, so a flickered or bounced-and-replayed creature
//     is a new object that was not cast from anywhere (CR 400.7).
//
// Celes itself is excluded ("other"), and a token comes from no zone
// and was not cast, so it never qualifies. The counters go on every
// creature its controller has when the trigger resolves, Celes
// included, as one placement per creature (Archangel of Thune's body).
//
// No simplification.
func init() {
	const (
		etbLabel     = "Celes, Rune Knight — discard any number of cards, then draw that many plus one"
		counterLabel = "Celes, Rune Knight — a +1/+1 counter on each creature you control"
	)
	Register(Spec{
		OracleID:     "aea8345a-191d-4c62-8d15-a716a9934b66",
		Name:         "Celes, Rune Knight",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(etbLabel, celesRuneKnightLoot),
			OncePerBatch(On(game.EventETB, celesOtherCreatureFromGraveyard,
				counterLabel, b11PutCounterOnEachCreatureYouControl)),
		},
	})
}

// celesRuneKnightLoot is the enters trigger: discard any number of
// cards, then draw that many plus one.
func celesRuneKnightLoot(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	hand := 0
	if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Hand != nil {
		hand = p.Hand.Size()
	}
	return b39MayDiscardThenDraw(hand, false,
		"Celes, Rune Knight — discard any number of cards, then draw that many plus one",
		func(discarded int) int { return discarded + 1 })(ctx)
}

// celesOtherCreatureFromGraveyard is the second ability's trigger and
// its intervening if in one: another creature the source's controller
// controls entered, and it either came out of a graveyard or was cast
// from one.
func celesOtherCreatureFromGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	if !ok || !c.IsCreature() {
		return false
	}
	return ev.EnteredFrom == game.ZoneGraveyard || c.Provenance.FromZone == game.ZoneGraveyard
}
