package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chandra, Fire of Kaladesh // Chandra, Roaring Flame — Magic Origins (S82
// older sets), a transforming legendary creature whose back face is a
// planeswalker.
//
// Front face, Legendary Creature — Human Shaman {1}{R}{R}, 2/2:
//
//	"Whenever you cast a red spell, untap Chandra.
//	 {T}: Chandra deals 1 damage to target player or planeswalker. If Chandra has
//	 dealt 3 or more damage this turn, exile her, then return her to the
//	 battlefield transformed under her owner's control."
//
// Back face, Legendary Planeswalker — Chandra, loyalty 4:
//
//	"+1: Chandra deals 2 damage to target player or planeswalker.
//	 −2: Chandra deals 2 damage to target creature.
//	 −7: Chandra deals 6 damage to each opponent. Each player dealt damage this
//	     way gets an emblem with "At the beginning of your upkeep, this emblem
//	     deals 3 damage to you.""
//
// The back face registers under "<oracle_id>#1". The untap is Belligerent
// Hatchling's red cast trigger. The tap ability counts every damage the object
// has dealt this turn, to players and to permanents, after its own damage
// (game.DamageDealtThisTurnByObjectForEffect). The −7 deals the damage as one
// instance and gives an emblem to each opponent it was really dealt to (damage
// prevented to 0 leaves none); the emblem is Chandra, Awakened Inferno's with
// three damage, and the emblem is its source.
//
// No simplification.
func init() {
	const oracle = "6f6b8e14-c247-4dfd-8b39-38c7a5f2f354"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Chandra, Fire of Kaladesh",
		Completeness: effects.CompletenessFull,
		// The emblem is declared on the front face too: a walker with exactly
		// seven loyalty dies as the −7 is paid, a card that has changed face on
		// its way out is its front face again (CR 712.8a), and the ability still
		// resolves (CR 608.2) and files the emblem under that card's key.
		Emblem: chandraRoaringFlameEmblem(),
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.OfColor("R"), "Chandra, Fire of Kaladesh — untap Chandra", effects.UntapThisPermanentEffect),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Chandra deals 1 damage to target player or planeswalker. If Chandra has dealt 3 or more damage this turn, exile her, then return her to the battlefield transformed under her owner's control.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Effect:  effects.OldB322ChandraTap,
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Chandra, Roaring Flame",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Emblem:          chandraRoaringFlameEmblem(),
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Chandra deals 2 damage to target player or planeswalker.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
				Targets: effects.TargetPlayerOrPlaneswalker(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DamageToFirstTargetStep(2)(item, effects.NewContext(g, item))
				},
			},
			{
				Label:   "−2: Chandra deals 2 damage to target creature.",
				Cost:    effects.LoyaltyCost(-2),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DamageToFirstTargetStep(2)(item, effects.NewContext(g, item))
				},
			},
			{
				Label:   "−7: Chandra deals 6 damage to each opponent. Each player dealt damage this way gets an emblem with \"At the beginning of your upkeep, this emblem deals 3 damage to you.\"",
				Cost:    effects.LoyaltyCost(-7),
				Purpose: game.Purpose{DamageEachOpponent: 6},
				Effect:  effects.OldB322ChandraUltimate,
			},
		},
	})
}

// chandraRoaringFlameEmblem is the emblem Chandra, Roaring Flame's −7 gives:
// Chandra, Awakened Inferno's upkeep damage with three damage, the emblem the
// source of it.
func chandraRoaringFlameEmblem() *effects.EmblemSpec {
	return &effects.EmblemSpec{
		Label: "Chandra, Roaring Flame emblem",
		Text:  "At the beginning of your upkeep, this emblem deals 3 damage to you.",
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Chandra, Roaring Flame emblem — 3 damage to you", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.DealDamage{Source: ctx.Source(), Target: ctx.Controller(), Amount: 3}.Apply(ctx)
			}),
		},
	}
}
