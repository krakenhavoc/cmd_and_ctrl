package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Steel Seraph — Artifact Creature — Angel {6}, 5/4 (The Brothers' War,
// #5656): "Prototype {1}{W}{W} — 3/3", "Flying", "At the beginning of
// combat on your turn, target creature you control gains your choice of
// flying, vigilance, or lifelink until end of turn."
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out. Flying
// is a printed keyword. The target is chosen when the trigger is put on
// the stack and re-checked as it resolves (CR 608.2b); the keyword is
// chosen as it resolves, by the Seraph's controller (Lunar Avenger's
// PickOption), and granted to that creature until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7ff8940f-d798-41eb-b82e-a1eba2163fe7",
		Name:            "Steel Seraph",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventStepBegan, StepBegan(game.StepBeginCombat, true),
				"Steel Seraph — target creature you control gains your choice of flying, vigilance, or lifelink until end of turn",
				steelSeraphGrant),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

// steelSeraphKeywords are the three choices, in the order the question lists them.
var steelSeraphKeywords = []string{"flying", "vigilance", "lifelink"}

func steelSeraphGrant(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	return PickOption{
		Question: "Steel Seraph — the creature gains which until end of turn?",
		Options: []game.ChoiceOption{
			{Label: "Flying"}, {Label: "Vigilance"}, {Label: "Lifelink"},
		},
		ThenKey: steelSeraphAnswer,
		Carry:   []uuid.UUID{target.ID},
	}.Apply(ctx)
}

// steelSeraphAnswer grants the chosen keyword to the carried creature.
var steelSeraphAnswer = OptionPickThen("option-pick/steel-seraph-keyword", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(steelSeraphKeywords) || len(r.Carry) == 0 {
		return nil
	}
	return GrantKeywordUntilEOT{Target: r.Carry[0], Keywords: []string{steelSeraphKeywords[r.Index]},
		Label: "Steel Seraph — " + steelSeraphKeywords[r.Index] + " until end of turn"}.Apply(ctx)
})
