package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fists of Flame — Instant {1}{R} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"Draw a card. Until end of turn, target creature gains trample and
//	 gets +1/+0 for each card you've drawn this turn."
//
// A spell whose only target is illegal does not resolve (CR 608.2b), so
// nothing is drawn then, as the ruling says. Otherwise the draw comes
// first, and the bonus is read from the caster's per-turn draw tally
// after it, so the card just drawn counts. The bonus is locked in as the
// spell resolves: a later draw this turn adds nothing. Cards put into
// the hand without being drawn are not in the tally.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e05abb6b-e9f4-4d9e-ad1e-7805ebc09914",
		Name:         "Fists of Flame",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "trample"))},
		OnResolve:    avB30dFistsOfFlameResolve,
	})
}

// avB30dFistsOfFlameResolve draws, then gives the target trample and
// +1/+0 for each card its controller has drawn this turn.
func avB30dFistsOfFlameResolve(item *game.StackItem, ctx *Context) error {
	if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
		return err
	}
	x := ctx.Game.TurnTallyFor(item.Controller).CardsDrawn
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Fists of Flame — trample until end of turn"}).Apply(ctx); err != nil {
			return err
		}
		if err := (BoostUntilEOT{Target: t.ID, Power: x, Label: "Fists of Flame — +1/+0 for each card drawn this turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
