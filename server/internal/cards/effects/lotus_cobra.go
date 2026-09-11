package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lotus Cobra — Creature — Snake, {1}{G}, 2/1:
//
//	"Landfall — Whenever a land you control enters, add one mana of
//	 any color."
//
// Two mana that turns every land drop into a ritual, and every
// fetchland into two. Landfall is a named trigger condition, not
// machinery: EventETB with a land under your control.
//
// # Why this is writable now
//
// The batch-02 triage (#295) listed Lotus Cobra as ready and it is,
// but only since AddMana landed with the roadmap's batch 01. The
// mana here comes from a TRIGGERED ability, which uses the stack —
// so it cannot be a ManaAbility (CR 605.3a) and needs the spell-side
// mana primitive. That is not a simplification: real Lotus Cobra's
// trigger really does use the stack, which is why it can be
// responded to and why the mana arrives a beat after the land.
//
// # "One mana of any color"
//
// Pipe syntax, "{W|U|B|R|G}", which queues the colour pick AddMana
// queues for any multi-option slot. AddMana narrows a pipe to the
// controller's commander colour identity UNCONDITIONALLY — unlike
// ManaAbility it has no IgnoreCommanderIdentity escape hatch — so
// that is engine posture here rather than a choice made on this
// card. It costs nothing in practice: a colour outside the deck's
// identity is a colour with nothing in the deck to spend it on. It
// would matter for a card that adds mana to be given away or spent
// on a cost outside the deck, and no such card is in the catalog.
//
// # Scope of the trigger
//
// "A land YOU control enters" — the land's controller, not its
// owner, and not restricted to lands PLAYED. A land put onto the
// battlefield by a fetchland, a Harrow or a Cultivate triggers it,
// which is the entire reason Lotus Cobra and fetchlands are played
// together.
//
// The Cobra itself entering is not a land and so never self-
// triggers; no exclusion is needed.
//
// The mana empties with the pool at the end of the step (CR 106.4).
// A Cobra trigger in an opponent's turn adds mana that must be spent
// that step — printed behaviour, and the reason the card is played
// on your own turn.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID: "8ad91f64-ccab-4edc-bd54-b2ee9267d614",
		Name:     "Lotus Cobra",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsLand() && c.Controller == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Lotus Cobra — landfall, add one mana of any color",
					func(g *game.Game, item *game.StackItem) error {
						return AddMana{
							Player:   item.Controller,
							Produced: "{W|U|B|R|G}",
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
