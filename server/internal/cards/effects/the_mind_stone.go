package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Mind Stone — Legendary Artifact — Infinity Stone {1}{W}
// (#1321, tracker #889):
//
//	"Indestructible
//	 {T}: Add {W}.
//	 {5}{W}, {T}: Harness The Mind Stone. (Once harnessed, its ∞
//	 ability is active.)
//	 ∞ — At the beginning of your end step, exile up to one other
//	 target nonland permanent you control, then return that card to
//	 the battlefield under its owner's control."
//
// The proof card for the HARNESSED designation (ADR 0071 amendment,
// #1321): a "∞ — [Ability]" trigger declares `ActiveWhen:
// effects.Harnessed()` and does not exist until CR 701.64a's own
// activated ability, `Harness`, has set `Card.Harnessed`. Everything
// else here is ordinary catalog machinery that happened to have no
// entry yet.
//
// INDESTRUCTIBLE rides PrintedKeywords, same as every other
// indestructible permanent.
//
// {T}: ADD {W} is an ordinary single-colour ManaAbility, Sol Ring's
// shape.
//
// THE HARNESS ABILITY is `Harness`'s whole point: a plain
// `{5}{W}, {T}` cost (Plus(ManaCost, TapCost)) whose effect is
// `HarnessForEffect`. Nothing gates its activation — CR 701.64a lets
// the ability be activated a second time and simply do nothing, so a
// player tapping the Stone again after it is harnessed pays five mana
// for no reason, exactly as printed.
//
// THE ∞ ABILITY is Thassa, Deep-Dwelling's end-step blink
// (thassa_deep_dwelling.go) one clause wider: "nonland permanent" in
// place of "creature", and "under ITS OWNER'S control" in place of
// "under YOUR control" — which is why Flicker's Controller is left at
// its zero value here instead of being set to the trigger's
// controller. "Up to one OTHER" excludes the Stone itself through
// AnotherTarget, exactly as Thassa's does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b175e826-09e8-4fae-9f2e-b902f95b282d",
		Name:            "The Mind Stone",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []ActivatedAbility{
			Harness("{5}{W}, {T}: Harness The Mind Stone.", Plus(ManaCost("{5}{W}"), TapCost())),
		},
		Triggered: []game.TriggeredAbility{theMindStoneEndStepFlicker()},
	})
}

// theMindStoneEndStepFlicker is "∞ — At the beginning of your end
// step, exile up to one other target nonland permanent you control,
// then return that card to the battlefield under its owner's
// control." Gated on Harnessed() so it does not exist at all until
// the Stone's own activated ability has set the designation
// (CR 702.186b).
func theMindStoneEndStepFlicker() game.TriggeredAbility {
	t := AtYourEndStep("The Mind Stone — blink another nonland permanent you control",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			// Controller left at its zero value: "under its OWNER's
			// control", not "under your control" — the clause Thassa,
			// Deep-Dwelling's own end-step blink does NOT print.
			return Flicker{Target: id}.Apply(ctx)
		})
	t.TargetsFrom = AnotherTarget(func(other CardPredicate) *game.TargetSpec {
		return TargetPermanent("up to one other target nonland permanent you control",
			YouControl(), Nonland(), other).WithCount(0, 1)
	})
	t.ActiveWhen = Harnessed()
	return t
}
