package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nameless Inversion — Kindred Instant — Shapeshifter, {1}{B} (Lorwyn
// Eclipsed, #4458):
//
//	"Changeling (This card is every creature type.)
//	 Target creature gets +3/-3 and loses all creature types until end
//	 of turn."
//
// One record until end of turn (CR 611.2a), so both halves share one
// timestamp: the +3/-3 is layer 7c and "loses all creature types" is a
// layer-4 type change (game.LoseAllCreatureTypesMod, CR 613.1d, #4691).
// Every creature type goes, a changeling's every-creature-type included,
// and the card types and other subtypes stay (CR 205.1a). While it lasts
// a lord stops pumping the creature, and a later "gains all creature
// types" applies on top and wins (CR 613.7). A 3-toughness creature dies
// to the -3 as a state-based action (CR 704.5f).
//
// Changeling is printed and works in every zone, so the card in a
// graveyard is every creature type; nothing here arranges that.
//
// No Purpose: the change is partly a benefit and partly a harm, which
// TargetPurpose.UntilEndOfTurn leaves undeclared (#4312), so the card is
// on untilEndOfTurnSpellExempt and the bot prices it as removal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d6911456-42cb-4b23-a560-dc98784530f5",
		Name:            "Nameless Inversion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Targets:         TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return untilEndOfTurn(ctx, item.Targets[0].ID, nil,
				"Nameless Inversion — +3/-3 and loses all creature types",
				game.ModifyPTMod(3, -3), game.LoseAllCreatureTypesMod())
		},
	})
}
