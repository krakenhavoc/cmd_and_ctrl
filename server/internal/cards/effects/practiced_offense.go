package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Practiced Offense — Sorcery {2}{W} (Secrets of Strixhaven, #4456):
//
//	"Put a +1/+1 counter on each creature target player controls. Target
//	 creature gains your choice of double strike or lifelink until end of
//	 turn.
//	 Flashback {1}{W} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Two target clauses (Clauses, read by slot), each re-checked on its own as the
// spell resolves and skipped when its target has left or is no longer legal
// (CR 608.2b), so neither half stops the other. The counters go on the
// creatures the targeted player controls as the spell resolves, so a creature
// already on the battlefield under their control is in the set; the creature
// gained from the second clause may be any creature, the player's or not, and
// it gets its keyword after the counters (CR 608.2c). The keyword is chosen by
// the caster as the spell resolves (Orcish Medicine's PickOption shape) and
// lasts until end of turn. The flashback is the keyword helper's alternative
// cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "99678cc9-59c8-4358-b847-8c38ffb4a050",
		Name:         "Practiced Offense",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPlayer("target player"),
			TargetCreature("target creature"),
		),
		Purpose:          ForTargets(BenefitUntilEOT(1, 0, 0)),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{1}{W}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if who, ok := ctx.ClauseTarget(0); ok && who.Kind == game.TargetPlayer {
				if err := socB32PutCounterOnEachCreaturePlayerControls(ctx, who.ID); err != nil {
					return err
				}
			}
			creature, ok := ctx.ClauseTarget(1)
			if !ok || creature.Kind != game.TargetCard || !onBattlefield(ctx.Game, creature.ID) {
				return nil
			}
			return PickOption{
				Question: "Practiced Offense — the creature gains which until end of turn?",
				Options:  []game.ChoiceOption{{Label: "Double strike"}, {Label: "Lifelink"}},
				ThenKey:  practicedOffenseKeyword,
				Carry:    []uuid.UUID{creature.ID},
			}.Apply(ctx)
		},
	})
}

// practicedOffenseKeyword is the keyword chosen, keyed (#4253): the target
// rides the prompt as data. An on-disk key: never renamed, never reused.
var practicedOffenseKeyword = OptionPickThen("option-pick/practiced-offense-keyword", func(ctx *Context, r game.OptionPicked) error {
	keywords := []string{"double strike", "lifelink"}
	if r.Index < 0 || r.Index >= len(keywords) || len(r.Carry) == 0 || r.Carry[0] == uuid.Nil {
		return nil
	}
	if !onBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   r.Carry[0],
		Keywords: []string{keywords[r.Index]},
		Label:    "Practiced Offense — " + keywords[r.Index],
	}.Apply(ctx)
})
