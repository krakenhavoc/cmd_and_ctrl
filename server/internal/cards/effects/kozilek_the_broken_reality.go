package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kozilek, the Broken Reality — Legendary Creature — Eldrazi {9}, 9/9:
//
//	"When you cast this spell, up to two target players each manifest
//	 two cards from their hands. For each card manifested this way, you
//	 draw a card. (To manifest a card, put it onto the battlefield face
//	 down as a 2/2 creature. Turn it face up any time for its mana cost
//	 if it's a creature card.)
//	 Other colorless creatures you control get +3/+2."
//
// The cast trigger fires from the stack and resolves above the spell,
// so a countered Kozilek still resolves it (the 2024-06-07 ruling).
// "Up to two" is a target count of zero to two: with no target nothing
// is manifested and nothing is drawn.
//
// Each target chooses its own two cards, in secret and together, and a
// player holding two cards or fewer is not asked: they manifest
// everything they hold, so a lone card is manifested (the ruling). The
// manifested cards enter face down under their owner's control, in one
// entry (EachManifestsFromHand, ADR 0189 decision 3). The caster draws
// one card per card that actually arrived, so a replacement that keeps
// a manifested card off the battlefield lowers the draw by one.
//
// The anthem reads layer 7c after layer 5, like Forsaken Monument's:
// a face-down 2/2 is colorless and a creature, so it is 5/4 under
// Kozilek, and Kozilek itself, which "other" excludes, stays 9/9.
//
// The 2024-06-07 ruling's second half, that a manifested card with
// morph or disguise may also be turned face up for that cost, is the
// double offer ADR 0189 decision 4 builds (slice 5). It applies to
// every manifest and cloak card, not to this one, so it is tracked on
// the face-down-objects roadmap row and not as a caveat here.
func init() {
	cast := WhenYouCastThisSpell("Kozilek, the Broken Reality — up to two target players each manifest two cards from their hands",
		Do(EachManifestsFromHand{
			Players: kozilekTargetPlayers,
			N:       2,
			Then:    kozilekDrawPerManifest,
		}))
	cast.Targets = TargetPlayer("up to two target players").WithCount(0, 2)
	Register(Spec{
		OracleID:     "39f4bf0e-47d5-4ae8-a504-71acda849850",
		Name:         "Kozilek, the Broken Reality",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{cast},
		Static: []game.StaticAbility{
			Labelled(b16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					target.IsColorless() && target.InstanceID != source.InstanceID
			}, 3, 2), "Other colorless creatures you control get +3/+2."),
		},
	})
}

// kozilekTargetPlayers is the cast trigger's players that are still
// legal targets as it resolves.
func kozilekTargetPlayers(ctx *Context) []uuid.UUID {
	var out []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			out = append(out, t.ID)
		}
	}
	return out
}

// kozilekDrawPerManifest is "for each card manifested this way, you
// draw a card".
func kozilekDrawPerManifest(ctx *Context, manifested []uuid.UUID) error {
	return DrawCards{N: len(manifested)}.Apply(ctx)
}
