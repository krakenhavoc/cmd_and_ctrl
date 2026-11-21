package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aurelia, Exemplar of Justice — Legendary Creature — Angel {2}{R}{W}, 2/5
// (Ravnica Remastered, #5948):
//
//	"Flying
//	 Mentor
//	 At the beginning of combat on your turn, choose up to one target
//	 creature you control. Until end of turn, that creature gets +2/+0,
//	 gains trample if it's red, and gains vigilance if it's white."
//
// Mentor is Blade Instructor's trigger. The beginning-of-combat trigger has an
// "up to one" target (CR 601.2c: none is a legal choice, and the ability still
// goes on the stack). The colours are read as the ability resolves, after the
// layers, so a creature that is both red and white gains both keywords and a
// colourless one gains neither; the +2/+0 comes either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2e1e0385-1537-4cb6-9d24-e4593d9a8f4a",
		Name:            "Aurelia, Exemplar of Justice",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Mentor("Aurelia, Exemplar of Justice — mentor"),
			Targeting(AtBeginningOfYourCombat("Aurelia, Exemplar of Justice — up to one target creature you control gets +2/+0, trample if red, vigilance if white",
				gapB74AureliaBoost),
				TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1)),
		},
	})
}

// gapB74AureliaBoost is the beginning-of-combat body: +2/+0 until end of turn
// on the target, then trample if it is red and vigilance if it is white.
func gapB74AureliaBoost(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := firstLegalCardTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	if err := (BoostUntilEOT{Target: id, Power: 2, Label: "Aurelia, Exemplar of Justice — +2/+0"}).Apply(ctx); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	var grant []string
	for _, col := range c.EffectiveColors() {
		switch col {
		case "R":
			grant = append(grant, "trample")
		case "W":
			grant = append(grant, "vigilance")
		}
	}
	if len(grant) == 0 {
		return nil
	}
	return GrantKeywordUntilEOT{Target: id, Keywords: grant, Label: "Aurelia, Exemplar of Justice — gains trample if red, vigilance if white"}.Apply(ctx)
}
