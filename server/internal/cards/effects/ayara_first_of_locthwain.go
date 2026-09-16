package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ayara, First of Locthwain — Legendary Creature — Elf Noble
// {B}{B}{B}, 2/3 (EDHREC rank 850):
//
//	"Whenever Ayara or another black creature you control enters,
//	 each opponent loses 1 life and you gain 1 life.
//	 {T}, Sacrifice another black creature: Draw a card."
//
// Mono-black's aristocrats commander: every black body is a drain on
// the way in and a card on the way out.
//
// Colour is read from the card's computed colours (Card.HasColor:
// Scryfall's `colors`, falling back to the mana cost), so a black
// creature token counts only if its template stamps Colors — the
// catalog's Zombie does and its Faerie Rogue does not, so a
// Bitterblossom token neither drains nor feeds the sacrifice. Weaker
// than printed, and the fix is a `Colors` stamp on the template, not
// this card.
//
// "Another" on the sacrifice cost is enforced by NAME (b03NotNamed):
// a cost clause never sees its own source, and a second Ayara is the
// legend rule's problem. Sacrificing Ayara herself is refused, which
// is exactly the printed restriction.
func init() {
	Register(Spec{
		OracleID:     "388168b3-ec68-4af2-b88c-6a5ec88c15f6",
		Name:         "Ayara, First of Locthwain",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"A black creature token only counts if its token template records its colour, so some older tokens neither trigger the drain nor can be sacrificed.",
			"A second copy or token copy of Ayara can't be sacrificed to her own ability.",
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return true
				}
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && c.HasColor("B")
			}, "Ayara — each opponent loses 1, you gain 1", drainEachOpponent),
		},
		Activated: []ActivatedAbility{{
			Label: "{T}, Sacrifice another black creature: Draw a card.",
			Cost: Plus(TapCost(), game.AbilityCost{
				SacrificeOther: sacrificeSpec("another black creature",
					Creature(), OfColor("B"), b03NotNamed("Ayara, First of Locthwain")),
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
