package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nissa, Leyline Tamer — Legendary Creature — Elf Wizard {3}{W}{U}{B}{R}, 4/5:
//
//	"Deathtouch, vigilance
//	 Landfall — Whenever a land you control enters, draw a card. Then if
//	 this is the first time this ability has resolved this turn, reveal
//	 cards from the top of your library until you reveal a creature
//	 card. Put that card onto the battlefield and the rest on the bottom
//	 of your library in a random order."
//
// Zimone, Mystery Unraveler's "first time this ability has resolved
// this turn": the resolution tally counts this resolution as it runs,
// so the first one reads 1. The tally is per object (#936, CR 400.7),
// so a Nissa that left and came back this turn starts again. Every
// landfall draws; only the first reveals. The reveal is Atla Palani's
// RevealUntilThenPutOntoBattlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6067b8d0-5b00-402c-967d-e35a1304fafc",
		Name:            "Nissa, Leyline Tamer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "vigilance"},
		Triggered: []game.TriggeredAbility{
			Landfall(nissaLeylineTamerLabel, nissaLeylineTamerLandfall),
		},
	})
}

const nissaLeylineTamerLabel = "Nissa, Leyline Tamer — landfall: draw a card, then the first time this turn reveal until a creature card and put it onto the battlefield"

func nissaLeylineTamerLandfall(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if g.ResolvedThisTurn(item.SourceCardID, nissaLeylineTamerLabel) > 1 {
		return nil
	}
	return RevealUntilThenPutOntoBattlefield{
		Match:  game.Card.IsCreature,
		Reason: "Nissa, Leyline Tamer — revealed until a creature card",
	}.Apply(ctx)
}
