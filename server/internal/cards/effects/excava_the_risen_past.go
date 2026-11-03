package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Excava, the Risen Past — Legendary Creature — Spirit Horse {2}{R}{W},
// 3/3 (Secrets of Strixhaven Commander, #4834):
//
//	"Flying, haste
//	 Whenever Excava attacks, return up to one target artifact, creature,
//	 or non-Aura enchantment card with mana value 3 or less from your
//	 graveyard to the battlefield with a finality counter on it. It's a
//	 1/1 Spirit creature with flying in addition to its other types. (If
//	 a creature with a finality counter on it would die, exile it
//	 instead.)"
//
// The target is chosen as the attack trigger goes on the stack and
// re-checked as it resolves (CR 608.2b); a card with {X} in its cost
// counts X as 0 in the graveyard (CR 202.3e, the ruling).
//
// "It's a 1/1 Spirit creature with flying" is how the card enters
// (CR 614.1c), so it rides the return as an entry overwrite (ADR 0210):
// the creature type and the Spirit subtype are added in layer 4
// (CR 205.1b), flying in layer 6 and the base 1/1 in layer 7b
// (CR 613.4b), one record at the entry's timestamp, pinned to the
// object for as long as it is on the battlefield. It is a creature from
// the moment it lands, so "whenever a creature enters" sees a returned
// artifact. An artifact or enchantment creature card is 1/1 too, and a
// later effect that sets its base power and toughness wins by timestamp
// (the rulings). The finality counter is an entry counter, and the
// finality exile is the engine's (CR 122.1h). "Spirit" is a word of the
// card's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "77216636-f76d-4c5e-b1ac-ab1844753f05",
		Name:            "Excava, the Risen Past",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Excava, the Risen Past — return up to one card with mana value 3 or less as a 1/1 Spirit with a finality counter", excavaReturn),
				TargetCardInGraveyard("up to one target artifact, creature, or non-Aura enchantment card with mana value 3 or less from your graveyard",
					YouOwn(), Or(Artifact(), Creature(), And(Enchantment(), Not(Subtype("Aura")))), ManaValueLE(3)).WithCount(0, 1)),
		},
	})
}

// excavaReturn is the attack trigger's resolution.
func excavaReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	return putCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, []uuid.UUID{targets[0].ID}, game.ZoneEntryOptions{
		Controller:         item.Controller,
		EntersWithCounters: map[string]int{game.CounterFinality: 1},
		EntersAs: &game.EntersAs{
			AddTypes:    []string{"Creature"},
			AddSubtypes: ctx.TextSubtypes("Spirit"),
			Keywords:    []string{"flying"},
			BasePT:      &game.EntersAsPT{Power: 1, Toughness: 1},
			Label:       "Excava, the Risen Past — a 1/1 Spirit creature with flying in addition to its other types",
		},
	}, game.TailRef{}, 0)
}
