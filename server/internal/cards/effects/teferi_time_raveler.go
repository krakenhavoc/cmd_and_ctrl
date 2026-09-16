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
// the catalog whose loyalty abilities do anything. Both replays
// (#329 report c3ecbf98, #334 report 2b4ce9cc) show the same thing:
// a Teferi on the battlefield with four loyalty counters and
// `activated_abilities: []`, the `tapped` bit flipping true / false
// as the player clicked him. There was nothing to activate because
// AbilityCost had no loyalty component — see ADR 0032 §7.
//
// WHAT IS WIRED
//
//   - −3 in full: the loyalty payment (CR 606.6 refuses it below
//     three counters), the bounce, and the draw. "Up to one target"
//     is a Min 0 / Max 1 clause, so the −3 is activatable on an
//     empty board and still draws — the draw is the half you always
//     get, and forcing a target would have made Teferi unusable
//     exactly when you most want the card.
//
//     This is a deliberate departure from the "up to one" modelling
//     Aang uses (an optional trigger with one required target).
//     Aang's shape needs a prompt to decline, and an activated
//     ability has no prompt to decline — the player simply confirms
//     the picker with nothing selected.
//
//   - +1 as a loyalty gain, and nothing more. See below.
//
// WHAT IS NOT WIRED, AND WHY
//
//   - The +1's continuous effect. "Until your next turn, you may
//     cast sorcery spells as though they had flash" needs two
//     things the engine does not have: an "as though" cast
//     permission (the cast gate reads HasKeyword(card, "flash") off
//     the card, with no per-player override — mutations.go:597),
//     and an "until your next turn" duration, which
//     turn_scoped_statics.go:97 records as inexpressible because
//     Turn.Number counts rounds, not seat-turns.
//
//     The ability is still registered, because a planeswalker that
//     can only ever tick DOWN is a worse lie than one whose plus
//     ability protects it: the loyalty gain, the once-per-turn
//     window and the sorcery-speed gate are all real, and the label
//     the player reads in the menu is the printed text. When the
//     two missing pieces land, this Effect stops being nil and
//     nothing else about the card changes.
//
//   - The static ("Each opponent can cast spells only any time they
//     could cast a sorcery"). A cast-timing restriction imposed on
//     OTHER players is not a layer-system characteristic and has no
//     home in Spec.Static; it needs the same per-player cast gate
//     the +1 does. Opponents keep instant speed until it lands.
func init() {
	Register(Spec{
		OracleID:     "ae7604bb-4818-45a3-960c-cf3d83f15964",
		Name:         "Teferi, Time Raveler",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Opponents can still cast spells at instant speed.", "The +1 only adds loyalty — it doesn't let you cast sorceries at instant speed."},
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner, which never see Scryfall data.
		StartingLoyalty: 4,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Until your next turn, you may cast sorcery spells as though they had flash. (loyalty only — the timing permission is not implemented)",
				Cost:  LoyaltyCost(1),
				// Effect nil: the loyalty counter IS the whole of
				// what this engine can do here today. See the
				// comment above.
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
