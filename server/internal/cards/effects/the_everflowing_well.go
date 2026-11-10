package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Everflowing Well // The Myriad Pools — a transforming legendary
// artifact (The Lost Caverns of Ixalan, #4470, #5403):
//
//	The Everflowing Well — Legendary Artifact {2}{U}
//	  "When The Everflowing Well enters, mill two cards, then draw two
//	   cards.
//	   Descend 8 — At the beginning of your upkeep, if there are eight
//	   or more permanent cards in your graveyard, transform The
//	   Everflowing Well."
//	The Myriad Pools — Legendary Artifact Land
//	  "{T}: Add {U}.
//	   Whenever you cast a permanent spell using mana produced by The
//	   Myriad Pools, up to one other target permanent you control
//	   becomes a copy of that spell until end of turn."
//
// The front face's entry trigger mills and then draws in printed order
// in one resolution. Descend 8 is an intervening if (CR 603.4), checked
// as the upkeep begins and again as the trigger resolves, and the
// transform turns this object over in place (CR 701.27f is
// TransformThis's own check).
//
// The back face's trigger is a spend rider on the {U} (#1547), as
// Tecutlan's is: it fires when that mana pays for a permanent spell and
// goes on the stack above it, so it resolves first (the card's ruling).
// It is the first rider with a target clause (#5403,
// game.ManaSpendTrigger.Targets): "up to one other target permanent you
// control" is chosen as the trigger is put on the stack (CR 603.3d), and
// "other" is The Myriad Pools itself. Choosing nothing is legal and the
// trigger then does nothing.
//
// "Becomes a copy of that spell" takes the spell's copiable values
// (CR 707.2) through game.CopiableValuesOfSpellForEffect: the spell as
// it is on the stack, or as it last existed there once it was countered
// (CR 608.2h). The copy is a duration copy (game/duration_copy.go) that
// ends at cleanup, and the permanent is not entering, so no "enters"
// ability of the copied spell applies (the card's ruling). Counters and
// other non-copy effects on the permanent keep applying (CR 613.1a). An
// X in the copied mana cost is 0 on the battlefield (CR 202.3e), which
// the copied values already say.
//
// No simplification.
const theEverflowingWellOracleID = "1f57a9f1-6b95-4395-bdf0-c5289b786ab1"

func init() {
	Register(Spec{
		OracleID:     theEverflowingWellOracleID,
		Name:         "The Everflowing Well",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("The Everflowing Well — mill two cards, then draw two cards",
				Do(MillCards{N: 2}, DrawCards{N: 2})),
			InterveningIf(CondDescend{N: 8}, On(game.EventBeginUpkeep, AllOf(ByYou, DescendHolds(8)),
				"The Everflowing Well — descend 8: transform it", Do(TransformThis{}))),
		},
	})

	Register(Spec{
		OracleID:       theEverflowingWellOracleID + "#1",
		Name:           "The Myriad Pools",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "{T}: Add {U}",
			SpendRiders: []game.ManaSpendRider{WhenManaSpent("The Myriad Pools", game.ManaSpendTrigger{
				Label:   "The Myriad Pools — up to one other target permanent you control becomes a copy of that spell until end of turn",
				Targets: theMyriadPoolsTargets,
				Effect:  theMyriadPoolsCopy,
			}, ManaRestrictCast, ManaRestrictAnyType("Artifact", "Creature", "Enchantment", "Planeswalker", "Battle"))},
		}},
	})
}

// theMyriadPoolsTargets is "up to one other target permanent you
// control", built fresh on every call.
func theMyriadPoolsTargets() *game.TargetSpec {
	spec := Another(TargetPermanent("up to one other target permanent you control", YouControl()))
	spec.Min = 0
	return spec
}

// theMyriadPoolsCopy is "becomes a copy of that spell until end of
// turn". The spell is the item's payload; the permanent is its target.
func theMyriadPoolsCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	spells := ctx.PayloadCards()
	if len(spells) == 0 {
		return nil
	}
	v, ok := g.CopiableValuesOfSpellForEffect(spells[0])
	if !ok {
		return nil
	}
	return BecomeCopy{
		Targets: []uuid.UUID{target},
		Of:      spells[0],
		Values:  &v,
		Label:   "The Myriad Pools — becomes a copy of that spell until end of turn",
	}.Apply(ctx)
}
