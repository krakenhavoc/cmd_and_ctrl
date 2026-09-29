package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dimir Doppelganger — Creature — Shapeshifter {1}{U}{B}, 0/2:
//
//	"{1}{U}{B}: Exile target creature card from a graveyard. This
//	 creature becomes a copy of that card, except it has this ability."
//
// A duration copy with no stated duration (#1593, become_copy.go): it
// lasts until the Doppelganger leaves the battlefield or copies the next
// card, and the engine keeps only the newest record.
//
// The copy is of "that card", the one the ability exiled, read in
// whatever zone it is in once the exile has happened — its own printed
// values (game.OwnPrintedValues). The copy does not depend on the exile
// succeeding: a commander that went to the command zone instead is
// still "that card".
//
// "Except it has this ability" is a grant of the ability as a named
// bundle (CR 707.9a), so a Doppelganger that has become a Hill Giant can
// activate it again. The card's own printed ability and the granted one
// are the same ActivatedAbility value, so they cannot drift.
const dimirDoppelgangerGrant = "dimir-doppelganger/this-ability"

var dimirDoppelgangerAbility = ActivatedAbility{
	Label:   "{1}{U}{B}: Exile target creature card from a graveyard. This creature becomes a copy of that card, except it has this ability.",
	Cost:    ManaCost("{1}{U}{B}"),
	Targets: TargetCardInGraveyard("target creature card from a graveyard", Creature()),
	Effect: func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		targets := ctx.LegalTargets()
		if len(targets) == 0 {
			return nil
		}
		card := targets[0].ID
		self := ctx.Source()
		return ExileTarget{Target: card, Then: func(ctx *Context, _ bool) error {
			c, ok := ctx.Game.LookupCardForEffect(card)
			if !ok {
				return nil
			}
			own := game.OwnPrintedValues(c)
			return BecomeCopy{
				Targets:  []uuid.UUID{self},
				Of:       card,
				Values:   &own,
				Duration: CopyIndefinite,
				Except:   func(v *game.PrintedValues) { v.GrantAbility(dimirDoppelgangerGrant) },
				Label:    "Dimir Doppelganger — becomes a copy",
			}.Apply(ctx)
		}}.Apply(ctx)
	},
}

func init() {
	Register(Spec{
		OracleID:     "1916f120-4418-404a-aed5-b95ebf60d3a3",
		Name:         "Dimir Doppelganger",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       dimirDoppelgangerGrant,
			Text:      "{1}{U}{B}: Exile target creature card from a graveyard. This creature becomes a copy of that card, except it has this ability.",
			Activated: []ActivatedAbility{dimirDoppelgangerAbility},
		}},
		Activated: []ActivatedAbility{dimirDoppelgangerAbility},
	})
}
