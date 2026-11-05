package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ezri Dax, Newly Joined — Legendary Creature — Trill Advisor {1}{W}, 2/2
// (Star Trek Commander, #4974):
//
//	"Flash
//	 You may have Ezri Dax enter as a copy of any creature card in your
//	 graveyard that was put there from the battlefield this turn, except
//	 her name is Ezri Dax, Newly Joined and she's a legendary Trill in
//	 addition to her other types."
//
// Body Double's enters-as-a-copy of a graveyard card (#4797, CR 707.2,
// CR 614.1c), narrowed to the creature cards in her controller's own
// graveyard that arrived from the battlefield during this turn
// (Card.PutInGraveyardFrom and PutInGraveyardThisTurn, which the next
// turn clears). Not targeting. The except clause is part of the copy
// effect (CR 707.9b): the copy is named Ezri Dax, Newly Joined, is
// legendary, and is a Trill as well as whatever it copied. Declined, or
// with no such card, she enters as her printed 2/2 self.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d4fb650a-94fd-4361-9648-45d6ad64b7a2",
		Name:            "Ezri Dax, Newly Joined",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf("Ezri Dax, Newly Joined", ezriDaxCandidates,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.SetName("Ezri Dax, Newly Joined")
					v.AddSupertype("Legendary")
					v.AddSubtype("Trill")
				}),
		},
	})
}

// ezriDaxCandidates is "any creature card in your graveyard that was put
// there from the battlefield this turn".
func ezriDaxCandidates(g *game.Game, controller uuid.UUID, _ uuid.UUID) []uuid.UUID {
	if g == nil {
		return nil
	}
	p := g.PlayerByIDForEffect(controller)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var out []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() && c.PutInGraveyardThisTurn && c.PutInGraveyardFrom == game.ZoneBattlefield {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
