package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wandering Archaic // Explore the Vastlands — Creature — Avatar {5},
// 4/4 // Sorcery {3} (slice 296-m):
//
//	Wandering Archaic: "Whenever an opponent casts an instant or
//	 sorcery spell, they may pay {2}. If they don't, you may copy that
//	 spell. You may choose new targets for the copy."
//	Explore the Vastlands: "Each player looks at the top five cards of
//	 their library and may reveal a land card and/or an instant or
//	 sorcery card from among them. Each player puts the cards they
//	 revealed this way into their hand and the rest on the bottom of
//	 their library in a random order. Each player gains 3 life."
//
// Front face only. Its trigger is Rhystic Study's own pay-unless
// shape one level deeper: PayUnless asks the CASTER (ev.Actor, not
// this controller) to pay {2}; declining opens a second, independent
// "you may copy that spell" question to Wandering Archaic's
// controller (MayChoice, the same "you may [do X] mid-resolution"
// primitive Rhystic Study's draw uses), and a "yes" copies the spell
// with CR 707.10c's "choose new targets" prompt.
//
// The back face, Explore the Vastlands, is left uncatalogued: "each
// player" independently choosing to reveal a land AND/OR an
// instant-or-sorcery card from the SAME five-card look, with only the
// chosen cards to hand and the rest to the bottom, has no primitive
// today. Every existing "look at top N" shape
// (LookAtTopThenMayPutOntoBattlefield, Impulse's take-one-to-hand) is
// one player choosing at most one card against a single predicate;
// this is one player choosing up to two cards against two independent
// predicates, done by every player at once. Casting the sorcery side
// still works — it just resolves as an uncatalogued sandbox spell,
// same as any other card the catalog hasn't reached.
func init() {
	Register(Spec{
		OracleID:     "6556c4c0-b10d-4208-821b-0c0a49abd188",
		Name:         "Wandering Archaic",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The sorcery side, Explore the Vastlands, isn't automated — casting it resolves as an ordinary sandbox spell.",
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor == source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (c.IsInstant() || c.IsSorcery())
			}, "Wandering Archaic — they may pay {2} or you may copy that spell", wanderingArchaicPayOrCopy),
		},
	})
}

func wanderingArchaicPayOrCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	spellID := ctx.Trigger().Event.CardID
	caster := ctx.Trigger().Event.Actor
	controller := item.Controller
	return PayUnless{
		Chooser:  caster,
		Cost:     "{2}",
		Question: "Wandering Archaic — pay {2}?",
		OnDecline: func(ctx *Context) error {
			return MayChoice{
				Player:   controller,
				Question: "Wandering Archaic — copy that spell?",
				OnYes: func(ctx *Context) error {
					return CopySpell{
						StackID:          spellID,
						Controller:       controller,
						ChooseNewTargets: true,
					}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	}.Apply(ctx)
}
