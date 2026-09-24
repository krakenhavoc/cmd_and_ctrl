package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Wandering Emperor — Legendary Planeswalker — Wanderer {2}{W}{W}
// with starting loyalty 3:
//
//	"Flash
//	 As long as The Wandering Emperor entered this turn, you may
//	 activate her loyalty abilities any time you could cast an
//	 instant.
//	 +1: Put a +1/+1 counter on up to one target creature. It gains
//	     first strike until end of turn.
//	 −1: Create a 2/2 white Samurai creature token with vigilance.
//	 −2: Exile target tapped creature. You gain 2 life."
//
// S14 registered her starting loyalty and left the abilities as a
// note for a future sprint. S27 (#329 / #334) wired them, and #1208
// finished the card and fixed it.
//
// THE STATIC is `Spec.ActivationTimings` (#1208, activation_timing.go)
// — a per-player statement derived from the battlefield on every
// query, exactly as Teferi, Time Raveler's cast-side static is. It is
// the card beating the rule: CR 606.3 makes every loyalty ability
// sorcery-speed and says so nowhere on any card, and CR 101.1 lets a
// printed clause win. Three things fall out of the derivation rather
// than needing code here — an Emperor that has lost her abilities
// (CR 613.1f) stops opening the window, an Emperor bounced in
// response shuts it before the activation is validated, and CR 101.2
// keeps a future restriction ahead of this grant because the one read
// folds restrictions last.
//
// "AS LONG AS SHE ENTERED THIS TURN" is `SourceEnteredThisTurn`,
// which is `game.EnteredThisTurn` and NOT `Card.SummonedThisTurn`.
// The summoning-sickness marker survives until its controller's untap
// step, so a permanent that entered on an opponent's turn still
// carries it on yours — and she has FLASH, so she lands on somebody
// else's turn nearly every time. Using it would have given her
// instant-speed loyalty abilities for a whole turn cycle.
//
// CR 606.3's OTHER half is untouched: she still gets one loyalty
// activation per turn, and `LoyaltyActivatedThisTurn` decides it.
//
// THE ABILITIES WERE WRONG and are corrected here. S27 shipped a
// permuted, half-invented set (a +1 that made the Samurai, a −1 that
// exiled, a −2 that pumped and granted lifelink); none of the three
// is what the card prints. The primitives were all already in the
// file — this is the same three effects re-attached to the right
// costs, plus the +1/+1 counter and the 2 life the printed card has
// and the old entry did not.
func init() {
	Register(Spec{
		OracleID:        "0c7f18d5-36cb-4bc6-a358-443b97666215",
		Name:            "The Wandering Emperor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		ActivationTimings: []game.ActivationTiming{
			ThisSourcesLoyaltyAbilitiesAtInstantSpeed(
				"As long as The Wandering Emperor entered this turn, you may activate her loyalty abilities any time you could cast an instant.",
				SourceEnteredThisTurn),
		},
		// The fallback for tokens, fixtures and the dev spawner;
		// an imported deck reads printed loyalty (ADR 0032 §1).
		StartingLoyalty: 3,
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Put a +1/+1 counter on up to one target creature. It gains first strike until end of turn.",
				Cost:    LoyaltyCost(1),
				Targets: upToOneCreature(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					// "Up to one" with nothing chosen: the ability
					// resolves and does nothing, which is a
					// resolution and not an error.
					if len(item.Targets) == 0 {
						return nil
					}
					target := item.Targets[0].ID
					if err := (AddCounter{
						Target: target,
						Kind:   game.CounterPlusOne,
						N:      1,
					}).Apply(ctx); err != nil {
						return err
					}
					// Two primitives because they are two layers: a
					// +1/+1 counter is layer 7d and read off the
					// card, the keyword grant is layer 6 and
					// turn-scoped. A single entry cannot sort into
					// both — see until_end_of_turn.go.
					return GrantKeywordUntilEOT{
						Target:   target,
						Keywords: []string{"first strike"},
						Label:    "The Wandering Emperor — first strike",
					}.Apply(ctx)
				},
			},
			{
				Label: "−1: Create a 2/2 white Samurai creature token with vigilance.",
				Cost:  LoyaltyCost(-1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return CreateToken{
						Controller: ctx.Controller(),
						Template:   TokenCard("2/2 white Samurai with vigilance"),
						N:          1,
					}.Apply(ctx)
				},
			},
			{
				Label:   "−2: Exile target tapped creature. You gain 2 life.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetCreature("target tapped creature", tappedPermanent()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					// This ability has exactly one target, so a
					// target that became illegal in response makes
					// EVERY target illegal — the engine's own
					// CR 608.2b re-check (resolveTopAbilityLocked)
					// has already fizzled the whole ability before
					// this Effect ever runs, and the life gain does
					// NOT happen either. This loop only ever sees a
					// legal target; the range is defensive, not a
					// real per-target skip.
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
				},
			},
		},
	})
}

// tappedPermanent is "…that is tapped", the restriction on the
// Emperor's −2. A target predicate rather than a resolution-time
// check so an untapped creature never appears in the picker, which
// is what CR 115.4 wants: an illegal target can't be chosen.
func tappedPermanent() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Tapped }
}

// upToOneCreature is the Emperor's +1 clause. Min 0 is the "up to
// one": the picker confirms with nothing selected.
func upToOneCreature() *game.TargetSpec {
	spec := TargetCreature("up to one target creature")
	spec.Min = 0
	return spec
}
