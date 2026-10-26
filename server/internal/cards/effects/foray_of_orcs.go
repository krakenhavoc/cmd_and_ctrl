package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Foray of Orcs — Sorcery {3}{R}:
//
//	"Amass Orcs 2. When you do, Foray of Orcs deals X damage to target
//	 creature an opponent controls, where X is the amassed Army's
//	 power."
//
// Amass hands the Army it chose to its Then (CR 701.47c), and the
// "when you do" is a CR 603.12 reflexive trigger that carries that Army
// in its payload. The target is chosen as the trigger goes on the
// stack, and X is the Army's power as the trigger resolves, so a pump
// or a Mauhúr-boosted counter count in response is counted, and an Army
// that died by then is read as it last stood.
//
// Amass always happens, so the reflexive trigger always exists. With no
// opposing creature it has no legal target and is removed (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60bfcaef-9014-4707-b681-cc3920bf223e",
		Name:         "Foray of Orcs",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Amass{
				Subtype:  "Orc",
				N:        2,
				ThenTail: forayOfOrcsReflexive,
			}.Apply(ctx)
		},
	})
}

// forayOfOrcsReflexive is "when you do, the Army deals damage equal to its
// power to target creature an opponent controls" as a keyed tail (#4253
// slice 6). An on-disk key.
var forayOfOrcsReflexive = AmassTail("foray-of-orcs/reflexive-damage", func(ctx *Context, army uuid.UUID) error {
	return ReflexiveTrigger{
		Label: "Foray of Orcs — damage equal to the amassed Army's power",
		Cards: []uuid.UUID{army},
		Body:  forayOfOrcsDamageBody,
	}.Apply(ctx)
})

// forayOfOrcsDamage is the reflexive body: the chosen creature takes
// damage equal to the Army's power, which is zero when the Army is not
// to be found.
func forayOfOrcsDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	power := 0
	if army := ctx.PayloadCards(); len(army) > 0 {
		power = b29PowerAsItLastStood(g, army[0])
	}
	if power <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: power}.Apply(ctx)
	}
	return nil
}
