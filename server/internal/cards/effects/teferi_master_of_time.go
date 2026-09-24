package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teferi, Master of Time — Legendary Planeswalker — Teferi {2}{U}{U},
// starting loyalty 3:
//
//	"You may activate loyalty abilities of Teferi on any player's
//	 turn any time you could cast an instant.
//	 +1: Draw a card, then discard a card.
//	 −3: Target creature you don't control phases out.
//	 −10: Take two extra turns after this one."
//
// THE STATIC is `Spec.ActivationTimings` (#1208), The Wandering
// Emperor's clause with the condition removed — which is the whole
// difference between the two cards and the reason
// `ThisSourcesLoyaltyAbilitiesAtInstantSpeed` takes `when` rather
// than baking one in.
//
// "OF TEFERI" IS A SELF-REFERENCE (CR 201.5): a card that names
// itself in its own text means THAT OBJECT, so this opens the window
// for this Teferi and not for a second one, and not for an opponent's
// Teferi, Time Raveler. The predicate compares instance IDs, which
// is that rule with nothing added.
//
// "ON ANY PLAYER'S TURN" is not modelled as a clause, because it is
// not one: it is what "any time you could cast an instant" already
// means, spelled out for players who would otherwise read CR 606.3
// into it. The read has no turn test on the grant side.
//
// CR 606.3's OTHER half still applies and is not this seam: one
// loyalty activation per turn per planeswalker. Teferi, Master of
// Time's reputation comes from having FOUR printed versions of
// himself (each a separate object with its own activation), not from
// activating one of them twice.
//
// WHAT IS WIRED
//
//   - The static, in full.
//   - +1 as b16DrawThenDiscard: the draw resolves and the discard is
//     a prompt, which is the shape every "draw then discard" in the
//     catalog already uses.
//   - −3 as PhaseOut (#1199, ADR 0084). Phasing is the engine's
//     since #1251, and the reminder text's whole content — attached
//     permanents come along, counters and damage survive, no
//     enters/leaves trigger fires, it phases in on its controller's
//     untap step — is the primitive's, not this card's.
//
// WHAT IS NOT WIRED
//
//   - The −10. "Take two extra turns after this one" is the extra
//     turns seam (docs/engine-seams.md): turn identity now handles a
//     seat taking two turns in a row, but the machinery still has no
//     queue or way to insert those turns. The ability is
//     registered with no effect rather than omitted, so the row is
//     visible, the loyalty is charged and the players resolve the
//     extra turns between themselves — the sandbox posture #259
//     asks for.
func init() {
	Register(Spec{
		OracleID:     "e802fb53-7cf5-46bc-8a0b-f99cf5c20f74",
		Name:         "Teferi, Master of Time",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The −10 doesn't take the two extra turns — the engine can't add a turn yet."},
		ActivationTimings: []game.ActivationTiming{
			ThisSourcesLoyaltyAbilitiesAtInstantSpeed(
				"You may activate loyalty abilities of Teferi on any player's turn any time you could cast an instant.",
				nil),
		},
		// The fallback for tokens, fixtures and the dev spawner;
		// an imported deck reads printed loyalty (ADR 0032 §1).
		StartingLoyalty: 3,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Draw a card, then discard a card.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b16DrawThenDiscard(g, item, 1, 1)
				},
			},
			{
				Label: "−3: Target creature you don't control phases out.",
				Cost:  LoyaltyCost(-3),
				Targets: TargetCreature("target creature you don't control",
					OpponentControls()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					// CR 608.2b: a target that became illegal in
					// response is skipped, and PhaseOut with an
					// empty list is a resolution that does nothing.
					return PhaseOut{Targets: legalTargetCards(item, g)}.Apply(ctx)
				},
			},
			{
				Label: "−10: Take two extra turns after this one.",
				Cost:  LoyaltyCost(-10),
				Effect: func(_ *game.Game, _ *game.StackItem) error {
					// Deliberately empty — see the caveat above. Not
					// an error: an ability the engine cannot carry
					// out is a sandbox gap, and EventEffectError is
					// what the catalog soak fails a game on.
					return nil
				},
			},
		},
	})
}
