package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Book of Vile Darkness — Legendary Artifact — Book {B}{B}{B}
// (Adventures in the Forgotten Realms, #5683):
//
//	"At the beginning of your end step, if you lost 2 or more life this
//	 turn, create a 2/2 black Zombie creature token.
//	 {T}, Exile The Book of Vile Darkness and artifacts you control named
//	 Eye of Vecna and Hand of Vecna: Create Vecna, a legendary 8/8 black
//	 Zombie God creature token with indestructible and all triggered
//	 abilities of the exiled cards."
//
// The end-step ability is an intervening if (CR 603.4): the life lost
// this turn is read as the end step begins and again as the ability
// resolves. The tally only rises within a turn.
//
// The cost exiles the Book (ExileThis) and every artifact its controller
// controls named Eye of Vecna or Hand of Vecna
// (ExileYourPermanentsNamed): the board decides the set, so nothing is
// picked, and with neither on the battlefield the cost is the tap and the
// Book alone. It is not a mana ability and has no timing restriction.
//
// Vecna's abilities are part of the token's text, defined by the effect
// that creates it (CR 111.3), so they are its copiable values (CR 707.2):
// CreateTokenWithTriggeredAbilitiesOf reads each exiled CARD as it is in
// exile when the ability resolves, the Book first and then the others in
// the order the cost exiled them, and gives the token their triggered
// abilities and nothing else, so the Book's own activated ability and
// the Hand's equip abilities do not come along. A card the cost did not
// leave in exile (a commander its owner moved to the command zone), or
// one that has left exile since, is not an exiled card and lends nothing;
// a token the cost exiled has ceased to exist and is not a card. Vecna
// enters with the abilities, so the Eye's "when this enters" triggers for
// Vecna's own arrival.
//
// No simplification.
func init() {
	const label = "{T}, Exile The Book of Vile Darkness and artifacts you control named Eye of Vecna and Hand of Vecna: Create Vecna, a legendary 8/8 black Zombie God creature token with indestructible and all triggered abilities of the exiled cards."
	Register(Spec{
		OracleID:     "fcc666bc-6fea-44e0-94bc-462c742db528",
		Name:         "The Book of Vile Darkness",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b18LifeLostThisTurn(g, source.Controller) >= 2
			}, "The Book of Vile Darkness — create a 2/2 black Zombie creature token",
				Do(If{
					Cond: CondLifeLost{AtLeast: 2},
					Then: []Applier{CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1}},
				})),
		},
		Activated: []ActivatedAbility{{
			Label:   label,
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(TapCost(), ExileThis(), ExileYourPermanentsNamed("artifact", "Eye of Vecna", "Hand of Vecna")),
			Effect:  bookOfVileDarknessCreateVecna,
		}},
	})
}

// bookOfVileDarknessCreateVecna is the activated ability's resolution:
// Vecna, with all triggered abilities of the cards this activation's cost
// exiled that are still in exile.
func bookOfVileDarknessCreateVecna(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var lenders []game.Card
	for _, id := range append([]uuid.UUID{item.SourceCardID}, ctx.Paid().Exiled...) {
		z := g.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if c, ok := g.LookupCardForEffect(id); ok && !IsToken(c) {
			lenders = append(lenders, c)
		}
	}
	return CreateTokenWithTriggeredAbilitiesOf(ctx, TokenCard("8/8 black Legendary Zombie God named Vecna with indestructible"), lenders)
}
