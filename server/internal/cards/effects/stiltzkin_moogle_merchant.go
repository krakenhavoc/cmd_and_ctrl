package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stiltzkin, Moogle Merchant — Legendary Creature — Moogle {W}, 1/2
// (Final Fantasy, #4462):
//
//	"Lifelink
//	 {2}, {T}: Target opponent gains control of another target permanent
//	 you control. If they do, you draw a card."
//
// Two clauses (slot 0 the opponent, slot 1 the permanent, never
// Stiltzkin: ExcludeSource). The gift is Humble Defector's indefinite
// layer-2 control change (CR 611.2a). "If they do" reads the report of
// that change: a permanent that left the battlefield, or an opponent who
// is no longer a legal target, hands over nothing and draws nothing.
//
// No simplification.
func init() {
	other := PermanentYouControl("another target permanent you control")
	other.ExcludeSource = true
	Register(Spec{
		OracleID:        "958c2a27-9fb1-42b4-a0ef-296584763a98",
		Name:            "Stiltzkin, Moogle Merchant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Target opponent gains control of another target permanent you control. If they do, you draw a card.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetPlayer("target opponent", Opponent()).Then(other),
			Effect:  Do(Instruction(stiltzkinGive)),
		}},
	})
}

// stiltzkinGive hands the permanent over and then draws.
func stiltzkinGive(ctx *Context) error {
	opp, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	perm, ok := ctx.ClauseTarget(1)
	if !ok {
		return nil
	}
	if !ctx.Game.GainControlForEffect(ctx.Source(), perm.ID, opp.ID, game.IndefiniteDuration(),
		"Stiltzkin, Moogle Merchant — target opponent gains control of it") {
		return nil
	}
	return DrawCards{N: 1}.Apply(ctx)
}
