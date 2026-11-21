package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elda, Conjurer of Spectacle — {2}{G}{W} Legendary Creature — Human
// Wizard 2/2 (Mystery Booster Commander, S82 GAP batch 3, #5948):
//
//	"The first nonlegendary creature spell you cast each turn has
//	 demonstrate. (When you cast that spell, you may copy it. If you do,
//	 choose an opponent to also copy it. Each copy becomes a token.)"
//
// The grant is a battlefield trigger watching the cast (the shape of
// Maelstrom Nexus's cascade, CR 702.144a for the keyword): the spell is
// carried as the item's object, since the copies are copies of the spell
// and not of Elda. "First" is read off the cast tally, which is bumped
// before the cast event: the creature spells cast this turn that are not
// legendary number exactly one, this spell. The demonstrate itself is the
// keyword's (Demonstrate): the controller may copy, then picks an opponent
// who copies it too, each copy on last-known information if the spell has
// left the stack.
//
// No simplification.
func init() {
	const label = "Elda, Conjurer of Spectacle"
	effects.Register(effects.Spec{
		OracleID:     "7c39639e-c611-46ee-beda-aef466170ba2",
		Name:         label,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(game.TriggeredAbility{
				Keyword: effects.KeywordDemonstrate,
				Key:     label,
				Watches: []game.EventKind{game.EventCast},
				AppliesTo: effects.SocB32YouCastASpellThat(func(spell game.Card, source *game.Card, g *game.Game) bool {
					if !spell.IsCreature() || spell.IsLegendary() {
						return false
					}
					tally := g.CastTallyFor(source.Controller)
					return tally.Total-tally.Noncreature-tally.LegendaryCreature == 1
				}),
				Build: effects.GapB45DemonstrateBuild(label),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return demonstrateSpell(effects.NewContext(g, item), item.Params.Object.ID)
				},
			}, "Elda, Conjurer of Spectacle — demonstrate: copy that spell? If you do, choose an opponent to also copy it."),
		},
	})
}

// demonstrateSpell is the demonstrate resolution for a spell that is not
// the trigger's own source (a granted demonstrate): copy it with new
// targets, then an opponent of the controller's choice copies it too.
func demonstrateSpell(ctx *effects.Context, spell uuid.UUID) error {
	if err := (effects.CopySpell{StackID: spell, ChooseNewTargets: true, FromLastKnown: true}).Apply(ctx); err != nil {
		return err
	}
	return effects.ChoosePlayer{
		Among:    effects.Opponents,
		Question: "Demonstrate — choose an opponent to also copy the spell",
		ThenKey:  effects.DemonstrateOpponentCopies,
		Carry:    []uuid.UUID{spell},
	}.Apply(ctx)
}
