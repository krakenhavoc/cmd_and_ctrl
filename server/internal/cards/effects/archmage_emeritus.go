package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archmage Emeritus — Creature — Human Wizard, {2}{U}{U}, 2/2:
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery
//	 spell, draw a card."
//
// A 2/2 that turns every cantrip into two cards. Magecraft is just a
// named trigger condition, not machinery of its own — the ability
// watches EventCast and gates on three things: the caster is this
// creature's controller, the spell is an instant or a sorcery, and
// the spell is not Archmage Emeritus itself (it cannot be — it is a
// creature — but the check costs nothing and documents the intent).
//
// One trigger per spell, and it goes on the stack ABOVE the spell
// that caused it, so the card is drawn before the instant resolves.
// That is the printed behaviour and it falls out of the harvester
// for free: EventCast fires at announce, the trigger drains onto the
// stack at the next priority boundary, and the stack is last-in
// first-out.
//
// The type test reads the printed type line via IsInstant() /
// IsSorcery(). A spell on the stack has no post-layer characteristic
// to consult, so printed is the only reading available and is also
// the right one for the overwhelming majority of cases.
//
// DECLARED SIMPLIFICATION — THE "OR COPY" HALF NEVER FIRES. Printed
// magecraft triggers on casting OR on copying an instant or sorcery,
// and the copy half is what makes the card explosive alongside storm
// and Fork effects. The engine emits no event for a spell COPY —
// there is no copy-a-spell primitive in the catalog and no
// EventSpellCopied for one to emit — so only the cast half is
// watched. This makes the card strictly worse than printed: it
// draws fewer cards, never more. When spell copying lands, this
// ability gains a second EventKind in Watches and the note goes.
func init() {
	Register(Spec{
		OracleID: "8305d576-21d8-4ce7-8eda-a7cd9793aca5",
		Name:     "Archmage Emeritus",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller || ev.CardID == source.InstanceID {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (c.IsInstant() || c.IsSorcery())
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Archmage Emeritus — magecraft, draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
