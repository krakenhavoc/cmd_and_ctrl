package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
// # The lifelink counter
//
// CR 122.1b: the engine reads keyword counters itself (ADR 0101), so
// the stolen creature has lifelink for as long as it has the counter,
// whether or not Sorin is still on the battlefield. Until ADR 0101 a
// static on the planeswalker carried the rule and stopped when Sorin
// left.
func init() {
	effects.Register(effects.Spec{
		OracleID:        SorinOfHouseMarkovOracleID,
		Name:            "Sorin of House Markov",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.Extort("Sorin of House Markov"),
			// The intervening if's first check is asked as the phase
			// begins: without the life gain it does not trigger.
			effects.On(game.EventStepBegan, effects.AllOf(effects.StepBegan(game.StepPostcombatMain, true), sorinGainedThree),
				"Sorin of House Markov — exile him, then return him transformed", sorinFlip),
		},
	})

	effects.Register(effects.Spec{
		OracleID:     SorinOfHouseMarkovOracleID + "#1",
		Name:         "Sorin, Ravenous Neonate",
		Completeness: effects.CompletenessFull,
		// Printed loyalty reaches a card through deck import (ADR 0032
		// §1); this is the fallback for fixtures and the dev spawner.
		StartingLoyalty: 3,
		Triggered: []game.TriggeredAbility{
			effects.Extort("Sorin, Ravenous Neonate"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+2: Create a Food token.",
				Cost:  effects.LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Template: effects.FoodToken(), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "−1: Sorin deals damage equal to the amount of life you gained this turn to any target.",
				Cost:    effects.LoyaltyCost(-1),
				Targets: effects.TargetAny(),
				Effect:  sorinDamageForLifeGained,
			},
			{
				Label:   "−6: Gain control of target creature. It becomes a Vampire in addition to its other types. Put a lifelink counter on it if you control a white permanent other than that creature or Sorin.",
				Cost:    effects.LoyaltyCost(-6),
				Targets: effects.TargetCreature("target creature"),
				Effect:  sorinTakeAndTurn,
			},
		},
	})
}

// sorinOfHouseMarkovOracleID is shared by both faces.
const SorinOfHouseMarkovOracleID = "84e84532-498b-40ea-9510-46fb2403939b"

// sorinGainedThreeThisTurn is the flip's intervening if.
func sorinGainedThreeThisTurn(g *game.Game, player uuid.UUID) bool {
	return effects.B15LifeGainedThisTurn(g, player) >= 3
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
	return effects.ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
}

// sorinDamageForLifeGained is the −1. The amount is read as the ability
// resolves, so life gained in response counts.
func sorinDamageForLifeGained(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return effects.DealDamage{
			Source: item.SourceCardID,
			Target: t.ID,
			Amount: effects.B15LifeGainedThisTurn(g, item.Controller),
		}.Apply(ctx)
	}
	return nil
}

// sorinTakeAndTurn is the −6, in printed order.
func sorinTakeAndTurn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
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
	if err := (effects.GainControl{
		Target:     target,
		Controller: you,
		Duration:   game.IndefiniteDuration(),
		Label:      "Sorin, Ravenous Neonate — gain control (no stated duration)",
	}).Apply(ctx); err != nil {
		return err
	}
	if err := (effects.ScopedEffectFor{
		Target:   target,
		Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Vampire")...)},
		Duration: g.PinnedTo(game.IndefiniteDuration(), target),
		Label:    "Sorin, Ravenous Neonate — it becomes a Vampire in addition to its other types",
	}).Apply(ctx); err != nil {
		return err
	}
	if !sorinControlsAnotherWhitePermanent(g, you, target, item.SourceCardID, ctx.TextColor("W")) {
		return nil
	}
	return effects.AddCounter{Target: target, Kind: game.CounterLifelink, N: 1}.Apply(ctx)
}

// sorinControlsAnotherWhitePermanent is the −6's condition: a white
// permanent `you` control other than the stolen creature and Sorin.
// Colours are read post-layer (HasColor reads EffectiveColors). `white`
// is the colour word as Sorin's text now reads it (CR 612.2).
func sorinControlsAnotherWhitePermanent(g *game.Game, you, stolen, sorin uuid.UUID, white string) bool {
	g.RecomputeLayersIfStaleLocked()
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you || c.InstanceID == stolen || c.InstanceID == sorin {
			continue
		}
		if c.HasColor(white) {
			return true
		}
	}
	return false
}
