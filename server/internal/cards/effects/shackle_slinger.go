package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shackle Slinger — Creature — Human Soldier {2}{U}, 3/2 (Outlaws of
// Thunder Junction, #4468):
//
//	"Whenever you cast your second spell each turn, choose target
//	 creature an opponent controls. If it's tapped, put a stun counter on
//	 it. Otherwise, tap it."
//
// Brimstone Roundup's second-spell read (YouCastYourSecondSpellEachTurn)
// with a target. Whether the creature is tapped is read as the ability
// resolves (CR 608.2h): a tapped one gets one stun counter (CR 122.1d),
// an untapped one is tapped and gets none. A target gone by resolution
// gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "59e8747a-3e3f-48d2-aaee-af94ecee378b",
		Name:         "Shackle Slinger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{Targeting(
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Shackle Slinger — stun target creature an opponent controls if it's tapped; otherwise tap it",
				shackleSlingerStunOrTap),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		)},
	})
}

// shackleSlingerStunOrTap is "If it's tapped, put a stun counter on it.
// Otherwise, tap it."
func shackleSlingerStunOrTap(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(ts[0].ID)
	if !ok {
		return nil
	}
	if c.Tapped {
		return AddCounter{Target: ts[0].ID, Kind: game.CounterStun, N: 1}.Apply(ctx)
	}
	return TapTarget{Target: ts[0].ID}.Apply(ctx)
}
