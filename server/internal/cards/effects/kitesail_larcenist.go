package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kitesail Larcenist — Creature — Human Pirate {2}{U}, 2/3:
//
//	"Flying, ward {1}
//	 When this creature enters, for each player, choose up to one
//	 other target artifact or creature that player controls. For as
//	 long as this creature remains on the battlefield, the chosen
//	 permanents become Treasure artifacts with '{T}, Sacrifice this
//	 artifact: Add one mana of any color' and lose all other
//	 abilities."
//
// The enters trigger is three existing shapes (#3863):
//
//   - THE TARGETS. "For each player, choose up to one … that player
//     controls" is one "up to one" clause per player, each bound to its
//     player: the controller's own (another artifact or creature, so
//     never the Larcenist itself), then the 1st, 2nd and 3rd seats after
//     theirs in turn order (controlledByNthOpponent, Lich's Relic's
//     binding). A clause for a seat the table does not have offers
//     nothing. Binding is the point: a permanent that changes hands in
//     response is an illegal target for the clause it was chosen under
//     (CR 608.2b), even if it moved to another player.
//   - THE TREASURE. Vraska, Betrayal's Sting's three mods at one
//     timestamp: layer 4 sets the types to Artifact — Treasure, which
//     takes every other card type and subtype with it and keeps the
//     supertypes (CR 205.1a); layer 6 removes every ability, then grants
//     the Treasure mana ability as a registered bundle (ADR 0093), the
//     removal first so the grant survives it.
//   - THE DURATION. "For as long as this creature remains on the
//     battlefield" is CR 611.2b (DurationWhileSourceRemains): a
//     Larcenist that leaves before its trigger resolves changes nothing,
//     and one that leaves later gives every chosen permanent its
//     characteristics back. The chosen set is locked as the ability
//     resolves (CR 611.2c), each permanent pinned to the object it is,
//     so one that leaves and returns is a new object the effect no
//     longer reaches (CR 400.7).
//
// A chosen permanent that is sacrificed for its Treasure mana is gone for
// good, as printed.
//
// No simplification.
const (
	kitesailLarcenistTreasureGrant = "kitesail-larcenist/treasure-mana"
	kitesailLarcenistEntersLabel   = "Kitesail Larcenist — the chosen permanents become Treasure artifacts"
)

func init() {
	enters := WhenThisEnters(kitesailLarcenistEntersLabel, kitesailLarcenistTreasureify)
	enters.Targets = kitesailLarcenistClauses()
	Register(Spec{
		OracleID:        "2452be47-cc23-47f7-a3a1-fec900bb0119",
		Name:            "Kitesail Larcenist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []AbilityGrant{{
			Key: kitesailLarcenistTreasureGrant,
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true, Sacrifice: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}, Sacrifice this artifact: Add one mana of any color",
			}},
			Text: "{T}, Sacrifice this artifact: Add one mana of any color.",
		}},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{1}"), "Kitesail Larcenist — ward {1}"),
			enters,
		},
	})
}

// kitesailLarcenistClauses is "for each player, up to one other target
// artifact or creature that player controls": the controller's clause
// first, then one per opponent seat in turn order.
func kitesailLarcenistClauses() *game.TargetSpec {
	artifactOrCreature := Or(Artifact(), Creature())
	yours := Another(TargetPermanent("up to one other target artifact or creature you control",
		YouControl(), artifactOrCreature)).WithCount(0, 1)
	var rest []*game.TargetSpec
	for n := 1; n < game.MaxPlayers; n++ {
		rest = append(rest, TargetPermanent(
			"up to one target artifact or creature controlled by the "+nthOpponentWords[n]+" opponent after you in turn order",
			controlledByNthOpponent(n), artifactOrCreature).WithCount(0, 1))
	}
	return Clauses(yours, rest...)
}

// kitesailLarcenistTreasureify turns every still-legal chosen permanent
// into a Treasure with no other abilities, for as long as the Larcenist
// remains on the battlefield.
func kitesailLarcenistTreasureify(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	chosen := map[uuid.UUID]bool{}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			chosen[t.ID] = true
		}
	}
	if len(chosen) == 0 {
		return nil
	}
	dur, ok := DurationWhileSourceRemains(ctx, ctx.Source())
	if !ok {
		return nil
	}
	return ScopedEffectFor{
		Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return chosen[c.InstanceID] },
		Mods: []game.Mod{
			game.SetTypesMod([]string{"Artifact"}, []string{"Treasure"}),
			game.LoseAllAbilitiesMod(),
			game.GrantAbilitiesMod(kitesailLarcenistTreasureGrant),
		},
		Duration: dur,
		Label:    "Kitesail Larcenist — a Treasure artifact",
	}.Apply(ctx)
}
