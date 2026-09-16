package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hashaton, Scarab's Fist — 1/3 Legendary Creature — Zombie Wizard
// for {W}{B}:
//
//	"Whenever you discard a creature card, you may pay {2}{U}. If you
//	 do, create a tapped token that's a copy of that card, except
//	 it's a 4/4 black Zombie."
//
// Three pieces, and the interesting one is the third:
//
//   - The discard trigger is EventDiscardCard, which carries the
//     discarded card, so the creature check reads the type line out
//     of the graveyard (the same route Mary Read takes).
//   - The optional cost is the MayPay primitive — "you may pay {N}.
//     If you do" is PayUnless with the consequence on the other
//     answer, and it now shares that prompt and that auto-tap.
//   - The token is a copy of a specific card rather than a template,
//     which is what CreateTokenCopy was built for. The copy carries
//     the discarded card's oracle ID, so its triggered abilities,
//     statics, keywords and activated abilities all resolve through
//     the ordinary catalog hooks — a copy of a catalog card really
//     behaves like the card.
//
// "Except it's a 4/4 black Zombie" REPLACES the creature types
// rather than adding to them (CR 707.9a): a discarded Human Wizard
// comes back as a Zombie, not a Zombie Human Wizard. Eternalize,
// which keeps the types, spells them out ("a 4/4 black Zombie Snake
// Wizard") — that contrast is the evidence. Supertypes survive, so a
// copy of a legendary creature is legendary.
//
// Sandbox simplifications, both declared rather than papered over:
//
//   - The copy is read from wherever the discarded card now sits
//     rather than from last-known information in the graveyard. In
//     practice the two agree — the copiable values are printed
//     values and don't change by zone — but a card exiled in
//     response to the trigger would be copied out of exile instead
//     of failing over to LKI. There is no LKI store reachable from
//     an ability's Effect.
//   - The token's ETB *triggers* fire, but a copied card whose ETB
//     lives in Spec.AsEnters rather than Spec.Triggered does not get
//     that clause, because CreateTokenForEffect does not call
//     fireETBHookLocked. See the note on CreateTokenCopy.
func init() {
	Register(Spec{
		OracleID:     "db266661-f783-4907-9e52-6963eec05431",
		Name:         "Hashaton, Scarab's Fist",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Token copies skip the enters-the-battlefield effect on some cards, and a discarded card exiled in response is still copied."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDiscardCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return discardedByYou(ev, source) && eventCardHasType(ev, g, "creature")
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				// Capture the discarded card's ID, not the card and
				// not the game: undo resolves this closure against a
				// cloned game, and an instance ID is stable across
				// the clone while a *Card pointer is not.
				discarded := ev.CardID
				return game.NewTriggeredItem(source, "Hashaton, Scarab's Fist — pay {2}{U} to copy the discarded creature",
					func(g *game.Game, item *game.StackItem) error {
						return MayPay{
							Chooser:  item.Controller,
							Cost:     "{2}{U}",
							Question: "Hashaton, Scarab's Fist — pay {2}{U} to create a tapped 4/4 black Zombie copy?",
							OnPay: func(ctx *Context) error {
								return CreateTokenCopy{
									Controller: item.Controller,
									Copy:       discarded,
									N:          1,
									Except:     hashatonZombieException,
								}.Apply(ctx)
							},
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}

// hashatonZombieException is the "except it's a tapped 4/4 black
// Zombie" clause. Tapped rides on the template because
// CreateTokenForEffect copies it wholesale — the token is never
// untapped on the battlefield, which is what "create a TAPPED token"
// says and what an OnETB tap would get subtly wrong.
func hashatonZombieException(t *game.Card) {
	t.Power = 4
	t.Toughness = 4
	t.VariableToughness = false // a printed 4, not a `*` stand-in (#683)
	t.Colors = []string{"B"}
	t.TypeLine = retypedTypeLine(t.TypeLine, "Zombie")
	t.Tapped = true
}
