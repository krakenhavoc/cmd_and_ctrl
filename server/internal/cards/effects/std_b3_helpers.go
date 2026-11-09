package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b3_helpers.go — shared pieces of the Standard set leftovers, batches
// 3 and 4 (#4452).

// stdB3RemoveChargeCountersBody is the delayed half of Systems Override:
// "remove ten charge counters from it at the beginning of the next end
// step". The listed permanents are the payload; one that has fewer than ten
// loses what it has, and one that is gone loses nothing. An on-disk key.
var stdB3RemoveChargeCountersBody = game.SimpleDelayedBody("systems-override/remove-ten-charge-counters",
	func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range item.Targets {
			if t.Kind != game.TargetCard {
				continue
			}
			c, ok := g.LookupCardForEffect(t.ID)
			if !ok || !onBattlefield(g, t.ID) {
				continue
			}
			n := c.Counters[game.CounterCharge]
			if n > 10 {
				n = 10
			}
			if n <= 0 {
				continue
			}
			if err := (AddCounter{Target: t.ID, Kind: game.CounterCharge, N: -n}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})

// stdB3ScheduleChargeRemoval is the "if you do" half: told how many charge
// counters landed, it schedules their removal at the next end step.
var stdB3ScheduleChargeRemoval = CardTail("systems-override/schedule-charge-removal", func(ctx *Context, o game.TailOutcome) error {
	if o.Amount <= 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Systems Override — remove ten charge counters",
		Cards: o.Tail.Carry,
		Body:  stdB3RemoveChargeCountersBody,
	}.Apply(ctx)
})

// stdB3Taken is a uuid helper: the single-element slice.
func stdB3Taken(id uuid.UUID) []uuid.UUID { return []uuid.UUID{id} }

// WarpForArtifactAndRedCreatureCardsInHand is Tannuk, Steadfast Second's
// "Artifact cards and red creature cards in your hand have warp {cost}"
// (CR 702.185a): a granted alternative cost (ADR 0118) that reaches only
// casts from hand, only of an artifact card or a red creature card, and
// that carries warp's exile clause (game.GrantedAltCostWarp, which the
// entry hook reads off the key). A card that prints its own warp keeps it
// as a second offer; the caster picks. The key is an on-disk identity.
func WarpForArtifactAndRedCreatureCardsInHand(cost string) game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:       game.GrantedAltCostWarp,
			Label:     "Warp " + cost,
			ManaCost:  cost,
			WarpExile: true,
		},
		Zones:       []game.ZoneKind{game.ZoneHand},
		SpellsAnyOf: []game.PermissionFilter{{CardType: "artifact"}, {CreatureOnly: true, ColorsAny: "R"}},
		SpellsText:  "artifact cards and red creature cards",
	}
}

// printedDroneBlocksOnlyFliersToken is Station Monitor's Drone: a 1/1
// colorless Drone artifact creature token with flying and "This token can
// block only creatures with flying." (teoe 3). Into the Fae Court's Faerie
// block rule on an artifact. Its slug is an on-disk identity: never
// renamed, never reused.
func printedDroneBlocksOnlyFliersToken() tokenTemplate {
	return tokenTemplate{
		Slug: "drone-blocks-only-fliers",
		Card: game.Card{
			Name:      "Drone",
			TypeLine:  "Token Artifact Creature — Drone",
			Power:     1,
			Toughness: 1,
			Keywords:  []string{"flying"},
		},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
		Text: "Flying\nThis token can block only creatures with flying.",
	}
}

// StdB3DroneToken is the Drone's card value for CreateToken.
func StdB3DroneToken() game.Card { return tokenFromCatalog(printedDroneBlocksOnlyFliersToken) }

// stdB3ArtifactsYouControl is "artifacts you control": the source's
// controller's permanents that are artifacts, read post-layer.
func stdB3ArtifactsYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.IsArtifact()
}

// stdB3AreAlsoFood is "<applies> are Foods in addition to their other
// types" (Ragost, Deft Gastronaut): a layer-4 subtype add (CR 205.1b) that
// takes nothing away, and does nothing to a permanent that is already a
// Food. An object that is not an artifact cannot have an artifact type
// (CR 205.3d), so a non-artifact the predicate matches is left alone.
// AreAlsoEquipment's sibling.
func stdB3AreAlsoFood(label string, applies func(target *game.Card, g *game.Game, source *game.Card) bool) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer4Type,
		AppliesTo: applies,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			if !eotHasType(c.Types, "Artifact") || eotHasType(c.Subtypes, "Food") {
				return
			}
			c.Subtypes = append(append([]string(nil), c.Subtypes...), "Food")
		},
		Label: label,
	}
}

// ChooseCreatureOrWarpedCardCost is Close Encounter's "As an additional cost
// to cast this spell, choose a creature you control or a warped creature
// card you own in exile" (CR 702.185b): the mandatory slot's one component.
// The caster names the one object on cast_spell as reveal_ids; nothing is
// revealed or moved. The spell reads the chosen object's power at
// resolution through Context.TappedPower (the payment record's one
// object): the creature as it is then, or as it last existed on the
// battlefield if it has left (CR 608.2h); a warped card in exile has
// its printed power.
func ChooseCreatureOrWarpedCardCost() *game.AdditionalCost {
	return &game.AdditionalCost{
		Reveal: &game.RevealCost{ChooseCreatureOrWarped: true},
		Label:  "Choose a creature you control or a warped creature card you own in exile",
	}
}
