package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancestral Anger — Sorcery {R} (Secrets of Strixhaven, #4456):
//
//	"Target creature gains trample and gets +X/+0 until end of turn, where
//	 X is 1 plus the number of cards named Ancestral Anger in your
//	 graveyard.
//	 Draw a card."
//
// The spell is on the stack as it resolves, so it is not itself one of the
// cards counted; the others named Ancestral Anger in the caster's
// graveyard are, read as the spell resolves. The draw is the spell's own
// step and happens whenever the spell resolves at all (a spell whose
// only target is illegal does not resolve, CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0828e8d-f01f-4088-9123-6d923ddb3242",
		Name:         "Ancestral Anger",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "trample"))},
		OnResolve:    ancestralAngerResolve,
	})
}

// ancestralAngerResolve gives the target trample and +X/+0, then draws.
func ancestralAngerResolve(item *game.StackItem, ctx *Context) error {
	x := 1
	if self, ok := ctx.Game.LookupCardForEffect(item.SourceCardID); ok {
		if p := ctx.PlayerByID(ctx.Controller()); p != nil && p.Graveyard != nil {
			for _, c := range p.Graveyard.Cards {
				if c.Name == self.Name && c.InstanceID != item.SourceCardID {
					x++
				}
			}
		}
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Ancestral Anger — trample until end of turn"}).Apply(ctx); err != nil {
			return err
		}
		if err := (BoostUntilEOT{Target: t.ID, Power: x, Label: "Ancestral Anger — +X/+0 until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return DrawCards{N: 1}.Apply(ctx)
}
