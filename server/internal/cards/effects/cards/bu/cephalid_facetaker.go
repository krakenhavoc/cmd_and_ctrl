package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const cephalidFacetakerGrant = "cephalid-facetaker/cant-be-blocked"

// Cephalid Facetaker — Creature — Octopus Rogue {2}{U}, 1/4:
//
//	"This creature can't be blocked.
//	 At the beginning of combat on your turn, you may have this creature
//	 become a copy of another target creature until end of turn, except
//	 it's 1/4 and has 'This creature can't be blocked.'"
//
// The target is chosen as the trigger goes on the stack ("another" keeps
// the Facetaker itself out), the "you may" is asked as it resolves
// (CR 603.5). A duration copy (BecomeCopy, CR 707.2) whose except clause is
// the CR 707.9a pair: its power and toughness are set to 1 and 4 in the
// copiable values, and the "can't be blocked" ability is a granted bundle,
// since the copy has lost the printed one. Both ride the copy, so they
// end with it at end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cfb41a2b-4d8b-47e3-a074-95a1bd502227",
		Name:         "Cephalid Facetaker",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		Grants: []effects.AbilityGrant{{
			Key:    cephalidFacetakerGrant,
			Static: []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
			Text:   "This creature can't be blocked.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Optional(cephalidFacetakerTrigger(), "Cephalid Facetaker — become a copy of the target creature until end of turn?"),
		},
	})
}

// cephalidFacetakerCopy becomes a copy of the picked creature.
func cephalidFacetakerCopy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var of uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			of = t.ID
			break
		}
	}
	if of == uuid.Nil {
		return nil
	}
	return effects.BecomeCopy{
		Targets:  []uuid.UUID{ctx.Source()},
		Of:       of,
		Duration: effects.CopyUntilEndOfTurn,
		Except: func(v *game.PrintedValues) {
			v.SetPowerToughness(1, 4)
			v.GrantAbility(cephalidFacetakerGrant)
		},
		Label: "Cephalid Facetaker — becomes a copy until end of turn",
	}.Apply(ctx)
}

// cephalidFacetakerTrigger is the beginning-of-combat trigger with its
// "another target creature" clause.
func cephalidFacetakerTrigger() game.TriggeredAbility {
	t := effects.AtBeginningOfYourCombat("Cephalid Facetaker — you may have it become a copy of another target creature until end of turn, except it's 1/4 and can't be blocked",
		cephalidFacetakerCopy)
	t.TargetsFrom = effects.AnotherTarget(func(other effects.CardPredicate) *game.TargetSpec {
		return effects.TargetCreature("another target creature", other)
	})
	return t
}
