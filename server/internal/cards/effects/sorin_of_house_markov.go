package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sorin of House Markov // Sorin, Ravenous Neonate (#1117) — a
// transforming legendary creature whose back face is a planeswalker.
//
// Front face, Legendary Creature — Human Noble {1}{B}, 1/4:
//
//	"Lifelink
//	 Extort
//	 At the beginning of each of your postcombat main phases, if you
//	 gained 3 or more life this turn, exile Sorin, then return him to
//	 the battlefield transformed under his owner's control."
//
// Back face, Legendary Planeswalker — Sorin, loyalty 3:
//
//	"Extort
//	 +2: Create a Food token.
//	 −1: Sorin deals damage equal to the amount of life you gained
//	     this turn to any target.
//	 −6: Gain control of target creature. It becomes a Vampire in
//	     addition to its other types. Put a lifelink counter on it if
//	     you control a white permanent other than that creature or
//	     Sorin."
//
// The back face registers under "<oracle_id>#1" (game.CatalogKey), as
// every double-faced card's does.
//
// # The flip
//
// "Exile Sorin, then return him … transformed" is ADR 0079's SECOND
// verb, ExileAndReturnTransformed: two zone changes and a new object
// (CR 400.7), so the planeswalker enters with its printed loyalty 3
// (CR 306.5b) — a face's StartingLoyalty rides SetFace — and not with
// anything the creature had. "Under his OWNER's control" leaves the
// Controller zero. "If you gained 3 or more life this turn" is an
// intervening if (CR 603.4), read off the turn tally as the phase
// begins and again as the trigger resolves; lifelink damage counts,
// because it is life gain (CR 702.15b).
//
// # The −6
//
// Gain control with no stated duration is indefinite (CR 611.2a), and
// "it becomes a Vampire in addition to its other types" is a layer-4
// subtype add pinned to that object (CR 611.2c), the Legend of
// Kyoshi's data record. The lifelink counter's condition is read as
// the ability resolves, after the theft: a white permanent you control
// other than the stolen creature and Sorin himself.
//
// # Declared simplification (the back face's caveat)
//
// The engine reads no keyword counters of its own (CR 122.1b), so the
// lifelink counter is honoured by a static on the planeswalker that
// grants lifelink to every creature with a lifelink counter — Vraska
// Joins Up's posture (b24KeywordCounterGrant). While Sorin is on the
// battlefield that is exactly the rule; once he leaves, the counter
// stays on the creature and stops granting lifelink. Weaker than
// printed.
func init() {
	Register(Spec{
		OracleID:        sorinOfHouseMarkovOracleID,
		Name:            "Sorin of House Markov",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			Extort("Sorin of House Markov"),
			// The intervening if's first check is asked as the phase
			// begins: without the life gain it does not trigger.
			On(game.EventStepBegan, AllOf(StepBegan(game.StepPostcombatMain, true), sorinGainedThree),
				"Sorin of House Markov — exile him, then return him transformed", sorinFlip),
		},
	})

	Register(Spec{
		OracleID:     sorinOfHouseMarkovOracleID + "#1",
		Name:         "Sorin, Ravenous Neonate",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The lifelink counter from Sorin's −6 only gives lifelink while Sorin is on the battlefield — once he leaves, the counter stays but stops working.",
		},
		// Printed loyalty reaches a card through deck import (ADR 0032
		// §1); this is the fallback for fixtures and the dev spawner.
		StartingLoyalty: 3,
		Triggered: []game.TriggeredAbility{
			Extort("Sorin, Ravenous Neonate"),
		},
		Static: []game.StaticAbility{b24KeywordCounterGrant("lifelink")},
		Activated: []ActivatedAbility{
			{
				Label: "+2: Create a Food token.",
				Cost:  LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Template: FoodToken(), N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "−1: Sorin deals damage equal to the amount of life you gained this turn to any target.",
				Cost:    LoyaltyCost(-1),
				Targets: TargetAny(),
				Effect:  sorinDamageForLifeGained,
			},
			{
				Label:   "−6: Gain control of target creature. It becomes a Vampire in addition to its other types. Put a lifelink counter on it if you control a white permanent other than that creature or Sorin.",
				Cost:    LoyaltyCost(-6),
				Targets: TargetCreature("target creature"),
				Effect:  sorinTakeAndTurn,
			},
		},
	})
}

// sorinOfHouseMarkovOracleID is shared by both faces.
const sorinOfHouseMarkovOracleID = "84e84532-498b-40ea-9510-46fb2403939b"

// sorinGainedThreeThisTurn is the flip's intervening if.
func sorinGainedThreeThisTurn(g *game.Game, player uuid.UUID) bool {
	return b15LifeGainedThisTurn(g, player) >= 3
}

// sorinGainedThree is the same condition as a trigger predicate.
func sorinGainedThree(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return sorinGainedThreeThisTurn(g, source.Controller)
}

// sorinFlip exiles Sorin and returns him transformed, if the life gain
// is still there as the trigger resolves (CR 603.4).
func sorinFlip(g *game.Game, item *game.StackItem) error {
	if !sorinGainedThreeThisTurn(g, item.Controller) {
		return nil
	}
	return ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(NewContext(g, item))
}

// sorinDamageForLifeGained is the −1. The amount is read as the ability
// resolves, so life gained in response counts.
func sorinDamageForLifeGained(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return DealDamage{
			Source: item.SourceCardID,
			Target: t.ID,
			Amount: b15LifeGainedThisTurn(g, item.Controller),
		}.Apply(ctx)
	}
	return nil
}

// sorinTakeAndTurn is the −6, in printed order.
func sorinTakeAndTurn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var target uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			target = t.ID
			break
		}
	}
	if target == uuid.Nil {
		return nil
	}
	you := item.Controller
	if err := (GainControl{
		Target:     target,
		Controller: you,
		Duration:   game.IndefiniteDuration(),
		Label:      "Sorin, Ravenous Neonate — gain control (no stated duration)",
	}).Apply(ctx); err != nil {
		return err
	}
	if err := (ScopedEffectFor{
		Target:   target,
		Mods:     []game.Mod{game.AddSubtypesMod("Vampire")},
		Duration: g.PinnedTo(game.IndefiniteDuration(), target),
		Label:    "Sorin, Ravenous Neonate — it becomes a Vampire in addition to its other types",
	}).Apply(ctx); err != nil {
		return err
	}
	if !sorinControlsAnotherWhitePermanent(g, you, target, item.SourceCardID) {
		return nil
	}
	return AddCounter{Target: target, Kind: "lifelink", N: 1}.Apply(ctx)
}

// sorinControlsAnotherWhitePermanent is the −6's condition: a white
// permanent `you` control other than the stolen creature and Sorin.
// Colours are read post-layer (HasColor reads EffectiveColors).
func sorinControlsAnotherWhitePermanent(g *game.Game, you, stolen, sorin uuid.UUID) bool {
	g.RecomputeLayersIfStaleLocked()
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you || c.InstanceID == stolen || c.InstanceID == sorin {
			continue
		}
		if c.HasColor("W") {
			return true
		}
	}
	return false
}
