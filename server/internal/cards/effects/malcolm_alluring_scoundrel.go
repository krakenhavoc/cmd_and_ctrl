package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Malcolm, Alluring Scoundrel — 2/1 Legendary Creature — Siren Pirate
// for {1}{U}:
//
//	"Flash
//	 Flying
//	 Whenever Malcolm, Alluring Scoundrel deals combat damage to a
//	 player, put a chorus counter on it. Draw a card, then discard a
//	 card. If there are four or more chorus counters on Malcolm, you
//	 may cast the discarded card without paying its mana cost."
//
// A two-mana flier that loots every time it connects and, from the
// fourth connection on, pays the loot back: the card you pitch is
// cast for free out of the graveyard. The counters are the clock, so
// the card rewards a table that cannot profitably block a 2/1 flier
// four times.
//
// # The order is the card
//
// Counter, then draw, then discard, then the check. The counter goes
// on FIRST, which is why the fourth hit is the one that pays — the
// "four or more" reads the counter this trigger just placed. The
// draw is above the discard (a loot, not a rummage), so the card
// pitched may be the one just drawn. And the check is read after the
// discard prompt is answered, from the live board, because the
// discard splits the resolution: an open prompt holds priority, so
// nothing can move the counters underneath it.
//
// # The free cast is a grant
//
// Declared simplification, shared with cascade, madness and suspend:
// taking the offer stamps a per-instance game.CastPermission on the
// ONE discarded card object (ADR 0066) priced at {0}, flash-timed
// (CR 608.2g — the trigger resolves in the combat damage step), and
// expiring at end of turn, rather than casting the card inside the
// trigger's resolution. The consequences, in full:
//
//   - There is a CR 117.3b response window paper does not have, and
//     the cast can be held until later in the same turn.
//   - The permission names ONE card object at its current epoch
//     (CR 400.7), so a discarded card that leaves the graveyard and
//     comes back is a new object with nothing granted, and the
//     permission is spent the moment the card reaches the stack.
//   - It expires at end of turn either way. Nothing here can leave a
//     card permanently castable, which is the shape the grant-instead-
//     of-inline-cast approach has to avoid (see game/cascade.go).
//
// An inline cast is not available to any card: the announce path has
// no frame for a half-validated cast made from inside a resolution,
// which is the whole reason the engine's four other "cast it for
// free" keywords are grants. See game/cascade.go for the argument.
//
// Chorus counters are an ordinary named counter — Card.Counters is
// string-keyed and needs no registry entry, the same way "muster" and
// "slumber" do not have one.
func init() {
	Register(Spec{
		OracleID:        "3bba3b36-f8d7-4fd6-892a-797226210b6e",
		Name:            "Malcolm, Alluring Scoundrel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(
				"Malcolm, Alluring Scoundrel — chorus counter, draw, then discard",
				malcolmChorusAndLoot),
		},
	})
}

// malcolmChorusCounter is the counter the card prints. Free-form, like
// every other card-specific counter name in the catalog.
const malcolmChorusCounter = "chorus"

// malcolmChorusThreshold is the printed "four or more".
const malcolmChorusThreshold = 4

// malcolmChorusAndLoot is the whole trigger: counter, draw, discard,
// and the offer hung off the discard's continuation so it reads the
// card that really left the hand (#1027) rather than guessing at it
// from a hand-size subtraction.
func malcolmChorusAndLoot(g *game.Game, item *game.StackItem) error {
	source, controller := item.SourceCardID, item.Controller
	// #1290: the draw/discard/offer sequence must not start until the
	// chorus counter has LANDED — a Doubling Season / Hardened Scales
	// board can pause the placement on a CR 616 ordering prompt, and
	// starting the discard prompt before that resolves would leave
	// two independent prompts open with no guaranteed order, so
	// malcolmOfferFreeCast's "four or more" check could read the
	// pre-placement count.
	return g.AddCounterThenForEffect(source, malcolmChorusCounter, 1, func(g *game.Game, _ int) error {
		ctx := NewContext(g, item)
		if err := (DrawCards{Player: controller, N: 1}).Apply(ctx); err != nil {
			return err
		}
		return g.PlayerDiscardsThenForEffect(game.DiscardPrompt{
			Player:   controller,
			Source:   source,
			N:        1,
			Question: "Malcolm, Alluring Scoundrel — discard a card",
		}, func(g *game.Game, discarded game.PromptedDiscards) error {
			return malcolmOfferFreeCast(g, controller, source, discarded.Cards())
		})
	})
}

// malcolmOfferFreeCast is the last sentence. It is reached even when
// nothing was discarded (an empty hand discards as many as it can,
// which is none) and does nothing then — there is no "the discarded
// card" to name.
//
// Captures nothing but IDs, so an undo replays it against the restored
// game.
func malcolmOfferFreeCast(g *game.Game, controller, source uuid.UUID, discarded []uuid.UUID) error {
	if len(discarded) == 0 {
		return nil
	}
	self, ok := g.LookupCardForEffect(source)
	if !ok || self.Counters[malcolmChorusCounter] < malcolmChorusThreshold {
		return nil
	}
	card := discarded[0]
	name := "the discarded card"
	if c, found := g.LookupCardForEffect(card); found {
		name = c.Name
	}
	return g.QueueMayCastForEffect(controller, source, card,
		"Malcolm, Alluring Scoundrel — cast "+name+" without paying its mana cost?",
		func(g *game.Game) error {
			g.GrantCastPermissionOverCardForEffect(card, game.CastPermission{
				Player: controller,
				// Named explicitly rather than filled in from wherever
				// the card landed: a card a replacement sent somewhere
				// else is not "the discarded card" in the graveyard,
				// and the grant should be refused rather than follow
				// it.
				Zone: game.ZoneGraveyard,
				// "{0}", not empty: an empty Cost means "pay the
				// printed cost", which is the opposite of what this
				// clause grants.
				Cost:     "{0}",
				CastOnly: true,
				Timing:   game.TimingFlash,
				Label:    "Malcolm, Alluring Scoundrel — cast it without paying its mana cost",
			})
			return nil
		}, nil)
}
