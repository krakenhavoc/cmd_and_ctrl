package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phenax, God of Deception — Legendary Enchantment Creature — God
// {3}{U}{B}, 4/7:
//
//	"Indestructible
//	 As long as your devotion to blue and black is less than seven,
//	 Phenax isn't a creature.
//	 Creatures you control have "{T}: Target player mills X cards, where
//	 X is this creature's toughness.""
//
// #4092 (Captain N'ghathrod's list). The God clause is the shared
// two-colour Theros static at threshold seven (godUnlessDevotionToColors,
// the Mogis shape); a permanent with a hybrid blue-black symbol counts
// once (CR 700.5).
//
// The granted ability is a layer-6 bundle on every creature you control
// (CR 113.10, ADR 0093; Burning Anger's shape): each creature's own
// ability, so its controller activates it, {T} taps THAT creature, CR
// 302.6 keeps a creature that entered this turn from using it, and "this
// creature" is the creature that carries it. X is read as the ability
// resolves, from the carrier's toughness as it stands then (a pump or an
// anthem in response counts) or, if the carrier has left the battlefield,
// as it last was (CR 608.2h). The grant is Phenax's, so it applies to
// Phenax too once he is a creature (the static is evaluated after the
// layer 4 type change), and it is gone with him: removing Phenax's
// abilities, or Phenax leaving, takes it from everything (CR 613.8a).
// While he is only an enchantment he gives it to the creatures around
// him and has none himself, because "creatures you control" does not
// include him.
//
// No simplification.
const phenaxGrant = "phenax-god-of-deception/mill"

func init() {
	Register(Spec{
		OracleID:        "046ccc8a-ec0d-4dc5-b41d-5e8a6d86812a",
		Name:            "Phenax, God of Deception",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Grants: []AbilityGrant{{
			Key:  phenaxGrant,
			Text: "{T}: Target player mills X cards, where X is this creature's toughness.",
			Activated: []ActivatedAbility{{
				Label:   "{T}: Target player mills X cards, where X is this creature's toughness.",
				Cost:    TapCost(),
				Targets: TargetPlayer("target player"),
				Effect:  phenaxMillsItsToughness,
			}},
		}},
		Static: []game.StaticAbility{
			godUnlessDevotionToColors(7, "U", "B"),
			Labelled(GrantAbilities(b16CreaturesYouControl, phenaxGrant),
				"Creatures you control have \"{T}: Target player mills X cards, where X is this creature's toughness.\""),
		},
	})
}

// phenaxMillsItsToughness is the granted ability's body: the first
// still-legal player target mills X cards, X being the toughness of the
// creature whose ability this is.
func phenaxMillsItsToughness(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	victim, ok := firstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	return MillCards{Player: victim, N: toughnessNowOrLastKnown(g, item.SourceCardID)}.Apply(ctx)
}

// toughnessNowOrLastKnown is "this creature's toughness" read as an
// ability resolves: its current toughness while it is on the battlefield,
// else the toughness it last had there (CR 608.2h). A negative toughness
// mills nothing (MillCards ignores N <= 0).
func toughnessNowOrLastKnown(g *game.Game, cardID uuid.UUID) int {
	if z := g.FindCardZoneForEffect(cardID); z != nil && z.Kind == game.ZoneBattlefield {
		if c, ok := g.LookupCardForEffect(cardID); ok {
			return c.CurrentToughness()
		}
	}
	return b29LastKnownToughnessOffBattlefield(g, cardID)
}
