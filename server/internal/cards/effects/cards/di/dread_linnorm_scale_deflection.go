package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const dreadLinnormOracleID = "1bf04a2a-7f96-412c-9d91-034bef26e1b6"

// Dread Linnorm // Scale Deflection — Adventure.
//
//	Dread Linnorm {6}{G} Creature — Snake Dragon 7/6
//	  "This creature can't be blocked by creatures with power 3 or less."
//	Scale Deflection {3}{G} Instant — Adventure
//	  "Put two +1/+1 counters on target creature and untap it. It gains
//	   hexproof until end of turn."
//
// The evasion is a block rule judged at declaration (CR 509.1b) against each
// blocker's current power. Scale Deflection's target is re-checked as it
// resolves (CR 608.2b); a creature that left gets nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     dreadLinnormOracleID,
		Name:         "Dread Linnorm",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(3), "creatures with power 3 or less"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     dreadLinnormOracleID + "#1",
		Name:         "Scale Deflection",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "hexproof")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"hexproof"}, Label: "Scale Deflection — hexproof"}.Apply(ctx)
		},
	})
}
