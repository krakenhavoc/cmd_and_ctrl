package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dissection Practice — Instant {B} (Secrets of Strixhaven, #4456):
//
//	"Target opponent loses 1 life and you gain 1 life.
//	 Up to one target creature gets +1/+1 until end of turn.
//	 Up to one target creature gets -1/-1 until end of turn."
//
// Three target clauses, in printed order (slot 0 the opponent, slot 1 the
// creature that grows, slot 2 the one that shrinks); the two creature
// clauses may name the same creature, since the card does not say
// "another" (CR 601.2c). Each clause is re-judged as the spell resolves
// (CR 608.2b): an opponent who became illegal loses nothing, and the
// creatures are handled on their own. "You gain 1 life" touches no target,
// so it happens whenever the spell resolves at all (CR 608.2b: a spell
// fizzles only when every one of its targets is illegal).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "73905f94-9665-4d82-8927-725d2c4fc4bd",
		Name:         "Dissection Practice",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPlayer("target opponent", Opponent()),
			UpToOneTargetCreature("up to one target creature gets +1/+1"),
			UpToOneTargetCreature("up to one target creature gets -1/-1"),
		),
		Purpose: ForTargets(
			game.TargetPurpose{Slot: 0, LifeLoss: 1},
			BenefitUntilEOT(1, 1, 1),
			HarmUntilEOT(2, -1, -1),
		),
		OnResolve: dissectionPracticeResolve,
	})
}

// dissectionPracticeResolve is the drain, then the +1/+1, then the -1/-1,
// each against its own clause's still-legal pick.
func dissectionPracticeResolve(item *game.StackItem, ctx *Context) error {
	if p, ok := ctx.ClauseTarget(0); ok && p.Kind == game.TargetPlayer {
		if err := ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, p.ID, -1); err != nil {
			return err
		}
	}
	if err := (GainLife{Player: ctx.Controller(), Amount: 1}).Apply(ctx); err != nil {
		return err
	}
	if c, ok := ctx.ClauseTarget(1); ok && c.Kind == game.TargetCard {
		if err := (BoostUntilEOT{Target: c.ID, Power: 1, Toughness: 1, Label: "Dissection Practice — +1/+1 until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	if c, ok := ctx.ClauseTarget(2); ok && c.Kind == game.TargetCard {
		return BoostUntilEOT{Target: c.ID, Power: -1, Toughness: -1, Label: "Dissection Practice — -1/-1 until end of turn"}.Apply(ctx)
	}
	return nil
}
