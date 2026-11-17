package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nissa, Ascended Animist — Legendary Planeswalker — Nissa
// {3}{G}{G}{G/P}{G/P}, starting loyalty 7 (Phyrexia: All Will Be One,
// #5479):
//
//	"Compleated ({G/P} can be paid with {G} or 2 life. For each {G/P} paid
//	 with life, this planeswalker enters with two fewer loyalty counters.)
//	 +1: Create an X/X green Phyrexian Horror creature token, where X is
//	     Nissa's loyalty.
//	 −1: Destroy target artifact or enchantment.
//	 −7: Until end of turn, creatures you control get +1/+1 for each Forest
//	     you control and gain trample."
//
// Compleated is ADR 0183's replacement (effects.Compleated), counted once
// for each symbol paid with life. The +1's X is Nissa's loyalty as the
// ability resolves, after the +1 cost put its counter on (CR 606.4), and a
// Nissa that has left is read as she last existed (CR 608.2h); the token's
// size is fixed then (CR 611.2d). The −7 counts Forests once as it resolves
// and gives the same bonus to the creatures you control at that moment
// (CR 611.2c, as Overrun does): a creature that arrives later gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9c8845d6-fa65-4731-b348-dc724fc2d285",
		Name:            "Nissa, Ascended Animist",
		Completeness:    CompletenessFull,
		StartingLoyalty: 7,
		PrintedKeywords: []string{"compleated"},
		Replacements:    []game.ReplacementEffect{Compleated()},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Create an X/X green Phyrexian Horror creature token, where X is Nissa's loyalty.",
				Cost:    LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect:  nissaAscendedAnimistPlusOne,
			},
			{
				Label:   "−1: Destroy target artifact or enchantment.",
				Cost:    LoyaltyCost(-1),
				Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
				Effect:  destroyFirstLegalTarget,
			},
			{
				Label:  "−7: Until end of turn, creatures you control get +1/+1 for each Forest you control and gain trample.",
				Cost:   LoyaltyCost(-7),
				Effect: nissaAscendedAnimistMinusSeven,
			},
		},
	})
}

func nissaAscendedAnimistPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := s82OneB20SourceLoyalty(ctx)
	return CreateToken{
		Controller: item.Controller,
		Template: game.Card{
			Name:           "Phyrexian Horror",
			TypeLine:       "Token Creature — Phyrexian Horror",
			Colors:         []string{"G"},
			Power:          x,
			Toughness:      x,
			PrintedPTKnown: true,
		},
		N: 1,
	}.Apply(ctx)
}

func nissaAscendedAnimistMinusSeven(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	forests := s82OneB20LandsControlled(g, item.Controller, ctx.TextSubtype("Forest"))
	yours := And(Creature(), YouControl())
	if forests > 0 {
		if err := (BoostUntilEOT{
			Match:     yours,
			Power:     forests,
			Toughness: forests,
			Label:     "Nissa, Ascended Animist — +1/+1 for each Forest you control",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return GrantKeywordUntilEOT{
		Match:    yours,
		Keywords: []string{"trample"},
		Label:    "Nissa, Ascended Animist — trample",
	}.Apply(ctx)
}
