package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marshaling Cry — Sorcery {1}{W}{W} (Future Sight, #5594):
//
//	"Creatures you control get +1/+1 and gain vigilance until end of
//	 turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)
//	 Flashback {3}{W} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Devoted Paladin's team pump as a spell, with the shared cycling row and
// a flashback cast from the graveyard. The creatures are the ones the
// caster controls as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "0dc30c06-3900-44e8-9e63-c232e42c1b5c",
		Name:             "Marshaling Cry",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{W}")},
		Activated:        []ActivatedAbility{Cycling("{2}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			yours := And(Creature(), YouControl())
			if err := (BoostUntilEOT{Match: yours, Power: 1, Toughness: 1, Label: "Marshaling Cry — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Match: yours, Keywords: []string{"vigilance"}, Label: "Marshaling Cry — vigilance"}.Apply(ctx)
		},
	})
}
