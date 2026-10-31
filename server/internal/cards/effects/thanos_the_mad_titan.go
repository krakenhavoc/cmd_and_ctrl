package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thanos, the Mad Titan — {R}{W}{B} Legendary Creature — Eternal Villain
// 4/4 (Marvel Super Heroes, #4455):
//
//	"Deathtouch, lifelink
//	 Power-up — {C}{W}{U}{B}{R}{G}: Put two +1/+1 counters on Thanos.
//	 Choose odd or even. Destroy each other creature with mana value of
//	 the chosen quality. (Activate each power-up ability only once. Reduce
//	 the cost by his mana cost if he entered this turn. Zero is even.)"
//
// The counters go on first, then the controller is asked odd or even as
// the ability resolves (CR 608.2c, in printed order; a PickOption keyed to
// thanosDestroyParity, so a table waiting on it restores). The sweep spares
// Thanos himself ("each other creature") and reads each creature's mana
// value as it is now, a token and a face-down creature being 0, which is
// even (CR 202.3e). It is a destruction, so indestructible creatures live.
//
// No simplification.
func init() {
	ab := PowerUp("{C}{W}{U}{B}{R}{G}: Put two +1/+1 counters on Thanos. Choose odd or even. Destroy each other creature with mana value of the chosen quality.",
		ManaCost("{C}{W}{U}{B}{R}{G}"), thanosPowerUp)
	ab.Purpose = game.Purpose{Answers: game.AnswerRemove | game.AnswerPump, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}}
	Register(Spec{
		OracleID:        "090d815b-b2c8-48cd-a0b3-dc662d5352ac",
		Name:            "Thanos, the Mad Titan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "lifelink"},
		Activated:       []ActivatedAbility{ab},
	})
}

func thanosPowerUp(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
		return err
	}
	return PickOption{
		Question: "Thanos — choose odd or even",
		Options:  []game.ChoiceOption{{Label: "Odd"}, {Label: "Even"}},
		ThenKey:  thanosDestroyParity,
	}.Apply(ctx)
}

// thanosDestroyParity destroys each other creature whose mana value has the
// chosen parity. An on-disk key.
var thanosDestroyParity = OptionPickThen("option-pick/thanos-destroy-parity", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 {
		return nil
	}
	odd := r.Index == 0
	thanos := ctx.Source()
	return DestroyAllMatching{
		Match: func(g *game.Game, _ uuid.UUID, c game.Card) bool {
			if !c.IsCreature() || c.InstanceID == thanos {
				return false
			}
			mv, ok := g.ManaValueForEffect(c)
			return ok && (mv%2 != 0) == odd
		},
	}.Apply(ctx)
})
