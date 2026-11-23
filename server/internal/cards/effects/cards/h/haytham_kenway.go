package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Haytham Kenway — Legendary Creature — Human Knight {2}{W}{U}, 3/3
// (Assassin's Creed, #5948):
//
//	"Protection from Assassins
//	 Other Knights you control get +2/+2 and have protection from Assassins.
//	 When Haytham Kenway enters, for each opponent, exile up to one target
//	 creature that player controls until Haytham Kenway leaves the
//	 battlefield."
//
// Protection from creatures of the type is a printed keyword on Haytham
// and a layer 6 grant on the other Knights you control (Riders of Gavony's
// token), beside Lord of the Accursed's TribalAnthem for the +2/+2, the lord
// excluded and read through the layers.
//
// The enters trigger has one "up to one target creature" clause per opponent
// (taken in turn order, three at this table), each limited to that opponent's
// creatures. Each chosen creature is
// exiled until this permanent leaves (Banisher Priest's ExileUntil, CR 610.3),
// each one still legal at resolution (CR 608.2b); Haytham gone before the trigger
// resolves exiles nothing.
//
// No simplification.
func init() {
	const label = "Haytham Kenway — for each opponent, exile up to one target creature that player controls until Haytham Kenway leaves the battlefield"
	t := effects.Targeting(effects.On(game.EventETB, effects.Self, label, haythamExile), haythamTargets)
	knights := effects.TribeFilter{Tribes: []string{"Knight"}, Others: true, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:        "ee6a9bac-cf5b-4edd-8cb2-874d8dd0e62a",
		Name:            "Haytham Kenway",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.ProtectionFromCreaturesOfType("Assassin")},
		Static: []game.StaticAbility{
			effects.TribalAnthem(knights, 2, 2),
			effects.TribalKeywordGrant(knights, game.ProtectionFromCreaturesOfType("Assassin")),
		},
		Triggered: []game.TriggeredAbility{t},
	})
}

// haythamTargets is the enters trigger's target clauses: one "up to one target
// creature that player controls" for each opponent, the opponents taken in turn
// order after Haytham's controller (the table seats at most three). Each clause
// admits only its own opponent's creatures, read as the trigger is put on the
// stack from the table's seat order (a pure function of the game and the
// controller, so no board-reading TargetsFrom is needed); a clause whose opponent
// has left the game or has no creature offers nothing and is skipped.
var haythamTargets = effects.Clauses(
	haythamOpponentClause("first", 0),
	haythamOpponentClause("second", 1),
	haythamOpponentClause("third", 2))

// haythamOpponentClause is "up to one target creature <that opponent> controls"
// for the opponent `nth` places after the caster in turn order.
func haythamOpponentClause(ordinal string, nth int) *game.TargetSpec {
	return effects.TargetCreature("up to one target creature your "+ordinal+" opponent in turn order controls",
		func(g *game.Game, caster uuid.UUID, c game.Card) bool {
			opp, ok := haythamNthOpponent(g, caster, nth)
			return ok && c.Controller == opp
		}).WithCount(0, 1)
}

// haythamNthOpponent is the nth (from 0) player still in the game after the
// caster in turn order.
func haythamNthOpponent(g *game.Game, caster uuid.UUID, nth int) (uuid.UUID, bool) {
	start := -1
	for i, p := range g.Seats {
		if p != nil && p.ID == caster {
			start = i
		}
	}
	if start < 0 {
		return uuid.Nil, false
	}
	seen := 0
	for k := 1; k < len(g.Seats); k++ {
		p := g.Seats[(start+k)%len(g.Seats)]
		if p == nil || p.Eliminated || p.ID == caster {
			continue
		}
		if seen == nth {
			return p.ID, true
		}
		seen++
	}
	return uuid.Nil, false
}

// haythamExile exiles each still-legal target until Haytham leaves.
func haythamExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !effects.OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	seen := map[uuid.UUID]bool{}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || seen[t.ID] {
			continue
		}
		seen[t.ID] = true
		if err := (effects.ExileUntil{Target: t.ID, ThisLeaves: true, Label: "Haytham Kenway — the exiled creature returns when Haytham Kenway leaves the battlefield"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
