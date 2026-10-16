package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyclave Apparition — Creature — Kor Spirit {1}{W}{W}:
//
//	"When this creature enters, exile up to one target nonland,
//	 nontoken permanent you don't control with mana value 4 or less.
//	 When this creature leaves the battlefield, the exiled card's owner
//	 creates an X/X blue Illusion creature token, where X is the mana
//	 value of the exiled card."
//
// The two halves are linked abilities (CR 607.2a): the leave trigger
// reads what THIS object's enter trigger exiled, from the same event
// record Ossification uses (b27ExiledWith), so an Apparition that left
// and came back has no memory of its first exile, and one that leaves
// before its own enter trigger resolves exiles the card for good and
// pays no token — the rules' outcome for that ordering.
//
// Unlike an Oblivion Ring, nothing comes back: the exile is permanent
// and the LEAVE is what pays the owner, in the currency of the card's
// mana value, as a token they create (not the Apparition's controller).
//
// The exiled card is remembered, not looked up in exile: a card that
// something else has already moved out of exile (CR 608.2h) still pays its
// owner, X being the mana value it had as it last existed in exile
// (#3859). A mana value is a printed fact that does not change with the
// zone, and it is read with {X} counted as 0, which is what a card in
// exile has (CR 202.3e), whatever the card is doing now.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d90af00a-d322-4265-9954-7b1e80702e18",
		Name:         "Skyclave Apparition",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters(skyclaveApparitionExileLabel, b27ExileChosenTarget),
				TargetPermanent("up to one target nonland, nontoken permanent you don't control with mana value 4 or less",
					Nonland(), Not(IsTokenPredicate()), OpponentControls(), ManaValueLE(4)).WithCount(0, 1)),
			On(game.EventLTB, Self, "Skyclave Apparition — the exiled card's owner creates an X/X blue Illusion creature token",
				skyclaveApparitionPayOwner),
		},
	})
}

const skyclaveApparitionExileLabel = "Skyclave Apparition — exile up to one target nonland, nontoken permanent you don't control with mana value 4 or less"

// skyclaveApparitionPayOwner gives the exiled card's owner an X/X blue
// Illusion, X being the exiled card's mana value.
func skyclaveApparitionPayOwner(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range b27EverExiledWith(g, item.SourceCardID, skyclaveApparitionExileLabel) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if err := illusionTokenFor(ctx, c.Owner, c.ManaValueWithX(0)); err != nil {
			return err
		}
	}
	return nil
}

// illusionTokenFor creates an X/X blue Illusion creature token under
// `owner`.
func illusionTokenFor(ctx *Context, owner uuid.UUID, x int) error {
	return CreateToken{
		Controller: owner,
		Template: game.Card{
			Name: "Illusion", TypeLine: "Token Creature — Illusion",
			Power: x, Toughness: x, Colors: []string{"U"},
		},
		N: 1,
	}.Apply(ctx)
}
