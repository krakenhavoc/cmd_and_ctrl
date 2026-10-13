package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// text_box.go — the card-side vocabulary for exchanging text boxes
// (CR 612.5, CR 701.12h; ADR 0160). The engine half is game/text_box.go.

// EntersExchangingTextBoxes is "As <name> enters, you may exchange his
// text box and another creature's" (Deadpool, Trading Card; ADR 0160
// Decision 6). It is the copy choice's prompt with the exchange flag:
// the entering permanent's controller picks one of `candidates`, or
// nothing, before it lands, and the exchange is registered once it has
// landed and before EventETB, so its enters triggers are the gained
// text's. Indefinite: nothing in the printed text ends it.
//
// Not targeting, so hexproof, shroud and protection do not narrow it
// (CR 115.10a).
func EntersExchangingTextBoxes(
	name string,
	candidates func(g *game.Game, controller uuid.UUID, self uuid.UUID) []uuid.UUID,
) game.ReplacementEffect {
	return game.ReplacementEffect{
		RowLabel:        "May exchange text boxes with another creature as it enters.",
		EntryOnly:       true,
		Watches:         []game.EventKind{game.EventZoneMove},
		SelfReplacement: true,
		Label:           name,
		PromptQuestion:  name + " — exchange text boxes with…?",
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove &&
				ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID
		},
		Controller: EnteringPermanentChooser,
		CopySelector: &game.CopySelector{
			Candidates: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) []uuid.UUID {
				controller, self := copyChooserAndSelf(ev, src)
				return candidates(g, controller, self)
			},
			ExchangeText: true,
		},
	}
}

// anotherCreatureOnBattlefield is "another creature": any creature on
// the battlefield but the entering permanent, whoever controls it.
func anotherCreatureOnBattlefield(g *game.Game, _ uuid.UUID, self uuid.UUID) []uuid.UUID {
	return copyCandidates(g, self, func(c game.Card) bool { return c.IsCreature() })
}

// ExchangeTargetTextBoxesWhileSourceRemains is the resolution of "choose
// two target creatures. For as long as this <permanent> remains on the
// battlefield, exchange the text boxes of those creatures" (Exchange of
// Words). CR 701.12a: both targets must still be legal, or none of the
// exchange happens; and a source that has already left ends the
// duration before it begins (CR 611.2b), so nothing happens then either.
func ExchangeTargetTextBoxesWhileSourceRemains(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.Targets()
	if len(targets) != 2 {
		return nil
	}
	for _, t := range targets {
		if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
			return nil
		}
	}
	d, ok := g.ForAsLongAsOnBattlefieldDuration(item.SourceCardID)
	if !ok {
		return nil
	}
	g.ExchangeTextBoxesForEffect(item.SourceCardID, targets[0].ID, targets[1].ID, d, item.Label)
	return nil
}
