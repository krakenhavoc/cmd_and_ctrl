package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fireblade Artist — Creature — Human Shaman {B}{R}, 2/2 (Eventide):
//
//	"Haste
//	 At the beginning of your upkeep, you may sacrifice a creature. When
//	 you do, this creature deals 2 damage to target opponent or
//	 planeswalker."
//
// Ruthless Lawbringer's shape: the sacrifice is asked as the upkeep
// trigger resolves, a "may" with a floor of zero, over every creature you
// control (Fireblade Artist included). Only once a creature has really
// been sacrificed is the damage created, as a reflexive trigger that
// chooses its target after the sacrifice (CR 603.12). With the Artist gone
// by then, the damage is dealt by it as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "52630eaf-995e-4d08-be99-888a38b598b2",
		Name:            "Fireblade Artist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Fireblade Artist — you may sacrifice a creature", firebladeArtistSacrifice),
		},
	})
}

func firebladeArtistSacrifice(g *game.Game, item *game.StackItem) error {
	return effects.MaySacrificeOne(g, item, "Fireblade Artist — sacrifice a creature to deal 2 damage to target opponent or planeswalker?",
		effects.Creature(), false, firebladeArtistSacrificed)
}

// firebladeArtistSacrificed is the "when you do". An on-disk key: never
// renamed, never reused.
var firebladeArtistSacrificed = effects.RegisterSacrificedOne("prompt-run/fireblade-artist-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.WhenYouDo("Fireblade Artist — 2 damage to target opponent or planeswalker", firebladeArtistDamageBody).Apply(ctx)
})

// firebladeArtistDamageBody is the reflexive trigger. An on-disk key:
// never renamed, never reused.
var firebladeArtistDamageBody = game.ReflexiveBody("fireblade-artist/damage", effects.SimpleBody(effects.DealDamageToTheTarget(2)),
	effects.ConstTargets(effects.TargetOpponentOrPlaneswalker))
