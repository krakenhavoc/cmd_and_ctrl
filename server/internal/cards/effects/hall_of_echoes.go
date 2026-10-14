package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hall of Echoes — Land (Reality Fracture):
//
//	"{T}: Add {C}.
//	 {5}: This land becomes a copy of target creature you control until
//	 end of turn. The "legend rule" doesn't apply to permanents you
//	 control this turn."
//
// The second sentence is the half the engine lacked (#3515, ADR 0032's
// amendment of 2026-10-14). `Spec.LegendRuleExemptions` is a static read
// live off the battlefield with no duration, so a land declaring it would
// have switched the legend rule off for its controller for the whole
// game. The ability registers `LegendRuleDoesntApplyThisTurn` instead:
// a ModLegendRuleDoesntApply record naming the ability's controller,
// swept at cleanup (CR 514.2), carried by the snapshot, read by the same
// state-based action as Mirror Box. It covers "permanents you control"
// only: an opponent's duplicate legends still go to their graveyard.
//
// The two halves are separate instructions. The copy lands on this land
// (BecomeCopy leaves a land that has left, or come back as a new object,
// alone, #1432) and the exemption is registered either way. If the target
// is no longer legal on resolution the whole ability does nothing
// (CR 608.2b), which is why neither half runs then.
//
// A land that becomes a copy of a creature is no longer a land and
// loses its {T}: Add {C}, as printed ("becomes a copy" replaces the
// copiable values; no "in addition to" clause). Both effects end at the
// same cleanup, so the land is a land again as the exemption ends and no
// pair of legends is left behind by Hall itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc0a51d4-5b7e-4987-b1fe-349fd118a625",
		Name:         "Hall of Echoes",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{5}: This land becomes a copy of target creature you control until end of turn. The \"legend rule\" doesn't apply to permanents you control this turn.",
			Cost:    ManaCost("{5}"),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect:  hallOfEchoesBecomeCopy,
		}},
	})
}

// hallOfEchoesBecomeCopy is both sentences of the ability: the land
// becomes a copy of the target creature until end of turn, and the
// legend rule stops applying to its controller's permanents this turn.
func hallOfEchoesBecomeCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	if err := (BecomeCopy{
		Targets: []uuid.UUID{ctx.Source()},
		Of:      targets[0].ID,
		Label:   "Hall of Echoes — becomes a copy until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	return LegendRuleDoesntApplyThisTurn{}.Apply(ctx)
}

// LegendRuleDoesntApplyThisTurn is "the legend rule doesn't apply to
// permanents <player> controls this turn" (Hall of Echoes; #3515, ADR
// 0032's amendment of 2026-10-14). Player defaults to the resolving
// item's controller ("permanents you control"). A stored
// ModLegendRuleDoesntApply record swept at cleanup (CR 514.2); for a
// printed, durationless "the legend rule doesn't apply" static use
// `Spec.LegendRuleExemptions`.
type LegendRuleDoesntApplyThisTurn struct {
	Player uuid.UUID
}

func (l LegendRuleDoesntApplyThisTurn) Apply(ctx *Context) error {
	player := l.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	ctx.Game.LegendRuleDoesntApplyThisTurnForEffect(ctx.Source(), player, "the legend rule doesn't apply to your permanents this turn")
	return nil
}
