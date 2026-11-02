package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lady Octopus, Inspired Inventor — Legendary Creature — Human Scientist
// Villain {U}, 0/2 (Marvel's Spider-Man, #4460):
//
//		"Whenever you draw your first or second card each turn, put an
//		 ingenuity counter on Lady Octopus.
//		 {T}: You may cast an artifact spell from your hand with mana value less
//		 than or equal to the number of ingenuity counters on Lady Octopus
//		 without paying its mana cost."
//
//	  - The trigger fires once per card: the draw's ordinal is counted off
//	    this turn's draw events up to and including the one that fired it, so
//	    the third card of a three-card draw does nothing, and a card you did
//	    not draw (a replaced draw) is not counted.
//	  - The activated ability is the free cast from hand every such card
//	    shares (offerFreeHandCast: the cast is made during the resolution,
//	    CR 608.2g, and the player may decline). The clause is data the
//	    permission carries: the spell is an ARTIFACT spell (judged on the
//	    face being cast) with mana value at most the counters, read as the
//	    ability resolves, from Lady Octopus's last-known counters if she has
//	    left. A spell with X in its cost has X = 0 (CR 107.3b). Additional
//	    costs are still paid.
//
// No simplification.
func init() {
	const label = "{T}: You may cast an artifact spell from your hand with mana value less than or equal to the number of ingenuity counters on Lady Octopus without paying its mana cost."
	Register(Spec{
		OracleID:     "8e51fdb8-4f83-4e94-b36c-40a501075ecb",
		Name:         "Lady Octopus, Inspired Inventor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, ladyOctopusYourFirstOrSecondDraw,
				"Lady Octopus, Inspired Inventor — put an ingenuity counter on Lady Octopus",
				Do(CounterOnThis{Kind: ladyOctopusCounter, N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: label,
			// The free artifact may be a creature that blocks, so the
			// ability can answer a combat (ADR 0142).
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ladyOctopusOffer(NewContext(g, item))
			},
		}},
	})
}

const ladyOctopusCounter = "ingenuity"

// ladyOctopusYourFirstOrSecondDraw is the trigger condition: the card the
// event drew is the controller's first or second this turn.
func ladyOctopusYourFirstOrSecondDraw(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDrawCard || ev.Actor != source.Controller {
		return false
	}
	n := 0
	for _, e := range g.EventsThisTurn() {
		if e.Kind != game.EventDrawCard || e.Actor != source.Controller || e.Seq > ev.Seq {
			continue
		}
		n++
	}
	return n >= 1 && n <= 2
}

// ladyOctopusOffer is the activated ability's resolution.
func ladyOctopusOffer(ctx *Context) error {
	n := ladyOctopusCounters(ctx.Game, ctx.Source())
	offerFreeHandCast(ctx, ctx.Controller(), ladyOctopusCastOffer, "an artifact spell with mana value "+strconv.Itoa(n)+" or less",
		freeCastSpell{filter: game.PermissionFilter{CardType: "Artifact"}}.withManaValueAtMost(n))
	return nil
}

// ladyOctopusCounters is the number of ingenuity counters on Lady Octopus:
// as she is on the battlefield, else as she last was (CR 608.2h).
func ladyOctopusCounters(g *game.Game, id uuid.UUID) int {
	if onBattlefield(g, id) {
		if c, ok := g.LookupCardForEffect(id); ok {
			return c.Counters[ladyOctopusCounter]
		}
	}
	return g.LastKnownCountersForEffect(id)[ladyOctopusCounter]
}

// ladyOctopusCastOffer is the offer's keyed continuation (#4246), so a table
// waiting on it is a restore point.
var ladyOctopusCastOffer = registerFreeHandCastOffer("choose-cards/lady-octopus-cast", "Lady Octopus, Inspired Inventor")
