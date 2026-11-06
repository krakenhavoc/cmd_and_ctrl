package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glamer Gifter — Creature — Faerie Wizard {1}{U}, 1/2 (Lorwyn Eclipsed,
// #4458):
//
//	"Flash
//	 Flying
//	 When this creature enters, choose up to one other target creature.
//	 Until end of turn, that creature has base power and toughness 4/4
//	 and gains all creature types."
//
// One until-end-of-turn record with three mods at one timestamp, the way
// Allosaurus Shepherd writes its base size and type change: base power and
// toughness at layer 7b, so pumps and counters still apply on top, and
// all creature types at layer 4 (the fact Shields of Velis Vel grants).
// The target is any creature but the Gifter itself, either player's, and
// "up to one" lets the trigger choose none. A target that left in response
// is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e67a820b-6249-478e-91db-9827e44aa60a",
		Name:            "Glamer Gifter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Glamer Gifter — up to one other target creature is a 4/4 with all creature types until end of turn",
					glamerGifterEnters),
				Another(UpToOneTargetCreature("up to one other target creature"))),
		},
	})
}

func glamerGifterEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return untilEndOfTurn(ctx, id, nil,
		"Glamer Gifter — base power and toughness 4/4 and all creature types",
		game.SetBasePowerMod(4), game.SetBaseToughnessMod(4), game.AllCreatureTypesMod())
}
