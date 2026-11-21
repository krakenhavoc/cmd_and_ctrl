package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Exterminator Magmarch — Artifact Creature — Phyrexian Construct {2}{B}{R},
// 5/3:
//
//	"Whenever you cast an instant or sorcery spell that targets only a
//	 single nonland permanent an opponent controls, if another opponent
//	 controls one or more nonland permanents that spell could target,
//	 choose one of those permanents. Copy that spell. The copy targets the
//	 chosen permanent.
//	 {1}{B}: Regenerate this creature."
//
// The trigger fires for an instant or sorcery its controller cast with
// exactly one target, a nonland permanent an opponent controls, and only
// when a nonland permanent controlled by ANOTHER opponent is a legal target
// of that spell in its place (game.SpellCouldTargetInsteadForEffect: hexproof
// and protection from the caster, and a clause that excludes it, rule one
// out). That "if" is checked again as the trigger resolves (CR 603.4); with
// one such permanent it is chosen, with several its controller picks one,
// and the copy of the spell (Ivy's retargeted copy: controlled by the
// caster, targeting the chosen permanent in place of the original) goes on
// the stack above the spell. Regenerate is the usual shield.
func init() {
	const name = "Exterminator Magmarch"
	Register(Spec{
		OracleID:     "f42c0b5b-7a86-4f79-a5d3-859ec25ed870",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventCast || ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || (!spell.IsInstant() && !spell.IsSorcery()) {
					return false
				}
				from, ok := gapB44MagmarchSoleTarget(g, ev.CardID, source.Controller)
				return ok && len(gapB44MagmarchCandidates(g, ev.CardID, source.Controller, from)) > 0
			}, name+" — choose another opponent's nonland permanent that spell could target; copy that spell, the copy targeting it",
				gapB44MagmarchTrigger),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{1}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}

func gapB44MagmarchTrigger(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	spell, caster := item.Trigger.Event.CardID, item.Controller
	from, ok := gapB44MagmarchSoleTarget(g, spell, caster)
	if !ok {
		return nil
	}
	cands := gapB44MagmarchCandidates(g, spell, caster, from)
	switch len(cands) {
	case 0:
		return nil
	case 1:
		g.CopySpellRetargetedForEffect(spell, caster, from, cands[0])
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  caster,
		Source:   item.SourceCardID,
		Question: "Exterminator Magmarch — choose one of those permanents; the copy of that spell targets it",
		Cards:    cands,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  gapB44MagmarchChosen,
		Carry:    []uuid.UUID{spell, from},
	})
	return nil
}

// gapB44MagmarchChosen copies the spell with the permanent chosen as its
// target. An on-disk key.
var gapB44MagmarchChosen = game.RegisterChooseCardsThen("choose-cards/exterminator-magmarch-copy", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) < 2 || len(r.Picked) == 0 {
		return nil
	}
	g.CopySpellRetargetedForEffect(r.Carry[0], r.Chooser, r.Carry[1], r.Picked[0])
	return nil
})
