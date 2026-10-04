package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Artificer Class — Enchantment — Class, {1}{U}:
//
//	(Gain the next level as a sorcery to add its ability.)
//	The first artifact spell you cast each turn costs {1} less to cast.
//	{1}{U}: Level 2
//	When this Class becomes level 2, reveal cards from the top of your
//	library until you reveal an artifact card. Put that card into your
//	hand and the rest on the bottom of your library in a random order.
//	{5}{U}: Level 3
//	At the beginning of your end step, create a token that's a copy of
//	target artifact you control.
//
// Wizard Class's shape (ADR 0071), one line per level:
//
//   - Level 1 is a cost modifier with no gate: CR 716.3 gives a Class
//     the ability in its top section at all times. "The first artifact
//     spell you cast each turn" reads the caster's per-turn tally of
//     artifact spells (game.CastTally.Artifact, counted off the spell
//     on the stack as it becomes cast). A cost is determined (CR
//     601.2f) before the spell being priced is counted, so a tally of
//     zero means this is the first. An artifact spell cast earlier in
//     the turn, before the Class was on the battlefield, still counts,
//     so no later artifact spell that turn is discounted.
//   - Level 2 is BecomesLevel, gated at level 2 (CR 716.2a). The
//     reveal stops on the first artifact CARD (a token in a library is
//     never one, CR 108.2b). That card goes to its owner's hand and the
//     rest of the run goes to the bottom in a random order. A library
//     with no artifact card reveals itself entirely, nothing reaches
//     the hand, and every revealed card goes to the bottom.
//   - Level 3 is a targeted end-step trigger gated at level 3. Below
//     level 3 it does not exist, so it neither triggers nor asks for a
//     target. The token copies the artifact's copiable values (CR
//     707.2) through CreateTokenCopy; a target that left in response
//     is skipped (CR 608.2b).
//
// The level-up abilities are LevelUp: sorcery speed and "only if this
// Class is level N-1" (CR 716.2a) come with it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6dccf583-1045-4058-86cc-ebbcc8de080e",
		Name:         "Artificer Class",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "The first artifact spell you cast each turn costs {1} less to cast.",
				YourSpell(), ArtifactSpell(), NoArtifactSpellCastYetThisTurn()),
		},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{1}{U}")),
			LevelUp(3, ManaCost("{5}{U}")),
		},
		Triggered: []game.TriggeredAbility{
			BecomesLevel(2, "Artificer Class — level 2: reveal until an artifact card, put it into your hand",
				artificerClassRevealUntilArtifact),
			AtLevel(3, Targeting(
				AtYourEndStep("Artificer Class — create a token that's a copy of target artifact you control",
					func(g *game.Game, item *game.StackItem) error {
						return TokenCopyOfSingleTarget(item, NewContext(g, item))
					}),
				TargetPermanent("target artifact you control", Artifact(), YouControl()),
			)),
		},
	})
}

// artificerClassRevealUntilArtifact is the level-2 trigger's body.
func artificerClassRevealUntilArtifact(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	run, hit := revealUntil(ctx, player, func(c game.Card) bool { return c.IsArtifact() },
		"Artificer Class — reveal until an artifact card")
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  run,
		Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
		All:    true,
		Label:  "Artificer Class — put the artifact card into your hand",
		Then:   TakeRestOnBottomInRandomOrder,
	}.Apply(ctx)
}
