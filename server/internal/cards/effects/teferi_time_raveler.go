package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teferi, Time Raveler — Legendary Planeswalker — Teferi for
// {1}{W}{U}, starting loyalty 4:
//
//	"Each opponent can cast spells only any time they could cast a
//	sorcery.
//	+1: Until your next turn, you may cast sorcery spells as though
//	they had flash.
//	−3: Return up to one target artifact, creature, or enchantment
//	to its owner's hand. Draw a card."
//
// The card issue #334 is named after, and the first planeswalker in
// the catalog whose loyalty abilities do anything. It spent a sprint
// carrying two caveats that were both the same missing primitive, and
// #1195 built it: the whole card is wired.
//
// THE STATIC is `Spec.CastTimings` with the "each opponent" clause
// (`OpponentsCastAtSorcerySpeed`), derived from the battlefield on
// every query. Two things fall out of the derivation rather than
// needing a rule here: a Teferi that has lost its abilities
// (CR 613.1f) stops restricting, and CR 101.2's "can't beats can" is
// the ORDER the one read applies, so an opponent's Vedalken Orrery
// does not get them past this.
//
// THE +1 is the other home — a statement STORED on the player, with
// ADR 0063's "until your next turn" window (CR 611.2b: it ends as
// that turn BEGINS, CR 500.1, not at its cleanup). It has to be
// stored rather than derived because the planeswalker may die before
// your next turn comes and the permission does not die with it. The
// filter is `SorceryOnly` rather than `InstantOrSorceryOnly`: the
// clause says sorcery spells, and an instant needs no opening.
//
// WHAT IS NOT WIRED
//
//   - Nothing on this card. The −3's "up to one target" is a Min 0
//     clause, so it is activatable on an empty board and still draws
//     — the draw is the half you always get, and forcing a target
//     would have made Teferi unusable exactly when you most want the
//     card. That is a deliberate departure from the "up to one"
//     modelling Aang uses (an optional trigger with one required
//     target): Aang's shape needs a prompt to decline, and an
//     activated ability has no prompt to decline — the player simply
//     confirms the picker with nothing selected.
func init() {
	Register(Spec{
		OracleID:     "ae7604bb-4818-45a3-960c-cf3d83f15964",
		Name:         "Teferi, Time Raveler",
		Completeness: CompletenessFull,
		CastTimings:  []game.CastTimingRule{OpponentsCastAtSorcerySpeed()},
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner, which never see Scryfall data.
		StartingLoyalty: 4,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Until your next turn, you may cast sorcery spells as though they had flash.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GrantCastTiming{
						Filter:            game.PermissionFilter{SorceryOnly: true},
						UntilYourNextTurn: true,
						Label:             "Until your next turn, you may cast sorcery spells as though they had flash.",
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "−3: Return up to one target artifact, creature, or enchantment to its owner's hand. Draw a card.",
				Cost:    LoyaltyCost(-3),
				Targets: upToOneArtifactCreatureOrEnchantment(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					// The bounce is the optional half; the draw is
					// not. A target that became illegal between
					// announce and resolution (CR 608.2b) is
					// skipped and the draw still happens.
					for _, t := range ctx.Targets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (BounceToHand{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
				},
			},
		},
	})
}

// upToOneArtifactCreatureOrEnchantment is Teferi's −3 clause. Min 0
// is the "up to one" — the picker's Done button is live with
// nothing selected.
func upToOneArtifactCreatureOrEnchantment() *game.TargetSpec {
	spec := TargetPermanent(
		"up to one target artifact, creature, or enchantment",
		Or(Artifact(), Creature(), Enchantment()),
	)
	spec.Min = 0
	return spec
}
