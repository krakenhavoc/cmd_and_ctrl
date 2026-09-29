package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enter the Enigma — Sorcery {U}:
//
//	"Target creature can't be blocked this turn.
//	 Draw a card."
//
// The evasion grant is Artful Dodge's own RestrictUntilEOT clause —
// a restriction on the DEFENDING player's legal blocks, carried on
// the named creature, that nothing (reach, flying, "can block
// creatures with flying") gets around.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "67554654-e751-4679-9e92-f3588525ae4f",
		Name:         "Enter the Enigma",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (RestrictUntilEOT{
				Target:       item.Targets[0].ID,
				Restrictions: game.CantBeBlocked,
				Label:        "Enter the Enigma — can't be blocked",
			}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
