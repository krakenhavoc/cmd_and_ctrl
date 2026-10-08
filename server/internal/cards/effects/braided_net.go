package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Braided Net // Braided Quipu — a transforming artifact (#2124, ADR
// 0137):
//
//	Braided Net — Artifact {2}{U}
//	  "This artifact enters with three net counters on it.
//	   {T}, Remove a net counter from this artifact: Tap another target
//	   nonland permanent. Its activated abilities can't be activated for
//	   as long as it remains tapped.
//	   Craft with artifact {1}{U}"
//	Braided Quipu — Artifact
//	  "{3}{U}, {T}: Draw a card for each artifact you control, then put
//	   this artifact into its owner's library third from the top."
//
// The counters are Gemstone Mine's entry replacement. The lock is
// Deadlock Trap's (CantActivate and CantActivateMana, because a mana
// ability is an activated ability, CR 605.1a) held for as long as the
// target remains tapped (game.ForAsLongAsPinnedTappedDuration, ADR 0109
// §3): it ends the moment the permanent untaps, and a permanent that
// leaves is a new object it no longer follows (CR 400.7). A target that
// is not tapped once the tap has resolved gets no lock, because a "for
// as long as" duration that is already over never starts (CR 611.2b).
//
// The Quipu counts artifacts as the ability resolves, itself included,
// then tucks itself third from the top through the positioned landing
// God-Eternal Oketra uses. A Quipu that has left the battlefield by then
// still draws, and is not looked for in its new zone (CR 400.7).
//
// No simplification.
const braidedNetOracleID = "86ef06b3-2049-494d-9ecb-ca14766d3b68"

func init() {
	Register(Spec{
		OracleID:     braidedNetOracleID,
		Name:         "Braided Net",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters("net", 3, "Braided Net: enters with three net counters"),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{T}, Remove a net counter from this artifact: Tap another target nonland permanent. Its activated abilities can't be activated for as long as it remains tapped.",
				Cost:    Plus(TapCost(), RemoveCountersFromThis("net", 1)),
				Targets: Another(TargetPermanent("another target nonland permanent", Nonland())),
				Effect:  braidedNetLock,
			},
			Craft("Craft with artifact {1}{U}", "{1}{U}", CraftWith("artifact")),
		},
	})

	Register(Spec{
		OracleID:     braidedNetOracleID + "#1",
		Name:         "Braided Quipu",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:  "{3}{U}, {T}: Draw a card for each artifact you control, then put this artifact into its owner's library third from the top.",
			Cost:   Plus(ManaCost("{3}{U}"), TapCost()),
			Effect: braidedQuipuDrawThenTuck,
		}},
	})
}

// braidedNetLock taps the still-legal target and locks its activated
// abilities while it stays tapped.
func braidedNetLock(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	if err := (TapTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	dur, ok := g.ForAsLongAsPinnedTappedDuration(target)
	if !ok {
		return nil
	}
	return ScopedEffectFor{
		Target:   target,
		Mods:     []game.Mod{game.AddRestrictionsMod(game.CantActivate | game.CantActivateMana)},
		Duration: dur,
		Label:    "Braided Net — its activated abilities can't be activated while it remains tapped",
	}.Apply(ctx)
}

// braidedQuipuDrawThenTuck draws a card per artifact, then tucks the
// Quipu third from the top if it is still the permanent that activated.
func braidedQuipuDrawThenTuck(g *game.Game, item *game.StackItem) error {
	if n := b03ArtifactsControlled(g, item.Controller); n > 0 {
		if err := g.DrawNForEffect(item.Controller, n); err != nil {
			return err
		}
	}
	if g.AbilitySourceGoneForEffect(item) {
		return nil
	}
	return b22TuckThirdFromTop(g, item.SourceCardID)
}
