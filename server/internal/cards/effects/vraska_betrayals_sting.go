package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vraska, Betrayal's Sting — Legendary Planeswalker — Vraska
// {4}{B}{B/P}, loyalty 6:
//
//	"Compleated ({B/P} can be paid with {B} or 2 life. If life was paid,
//	 this planeswalker enters with two fewer loyalty counters.)
//	 0: You draw a card and lose 1 life. Proliferate.
//	 −2: Target creature becomes a Treasure artifact with '{T},
//	     Sacrifice this artifact: Add one mana of any color' and loses
//	     all other card types and abilities.
//	 −9: If target player has fewer than nine poison counters, they get
//	     a number of poison counters equal to the difference."
//
// Compleated is ADR 0183's replacement on the starting-loyalty
// placement (effects.Compleated, #2946). It reads only the printed
// Phyrexian symbol paid with life on the cast, so Vraska enters with 6
// when {B/P} was paid with {B} and 4 when it was paid with 2 life, and a
// doubler beside it is ordered by her controller (CR 616.1; the ruling:
// "other replacement effects ... apply as normal").
//
// 0: the draw and the life loss, then proliferate, whose choice is the
// player's (Proliferate{} asks).
//
// −2 is The Irencrag's shape: one indefinite effect (CR 611.2a) pinned
// to the creature, so it ends only if the permanent leaves and comes
// back as a new object (CR 400.7). Layer 4 sets its types to Artifact —
// Treasure, which takes every other card type and subtype with them and
// keeps its supertypes (the ruling: "It will retain any supertypes it
// had"). Layer 6 removes all its abilities and then grants the Treasure
// mana ability, the removal first so the grant survives it.
//
// −9: the difference is computed as the ability resolves and then
// placed through the player-counter window, so a replacement such as
// Melira, the Living Cure changes how many arrive (the ruling).
//
// No simplification.
const vraskaTreasureGrant = "vraska-betrayals-sting/treasure-mana"

func init() {
	Register(Spec{
		OracleID:        "05d09c55-f551-4cb4-a8d2-4811dccc848d",
		Name:            "Vraska, Betrayal's Sting",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"compleated"},
		Replacements:    []game.ReplacementEffect{Compleated()},
		Grants: []AbilityGrant{{
			Key: vraskaTreasureGrant,
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true, Sacrifice: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}, Sacrifice this artifact: Add one mana of any color",
			}},
			Text: "{T}, Sacrifice this artifact: Add one mana of any color.",
		}},
		Activated: []ActivatedAbility{
			{
				Label:  "0: You draw a card and lose 1 life. Proliferate.",
				Cost:   LoyaltyCost(0),
				Effect: vraskaDrawLoseProliferate,
			},
			{
				Label:   "−2: Target creature becomes a Treasure artifact with \"{T}, Sacrifice this artifact: Add one mana of any color\" and loses all other card types and abilities.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetCreature("target creature"),
				Effect:  vraskaTreasureify,
			},
			{
				Label:   "−9: If target player has fewer than nine poison counters, they get a number of poison counters equal to the difference.",
				Cost:    LoyaltyCost(-9),
				Targets: TargetPlayer("target player"),
				Effect:  vraskaPoisonToNine,
			},
		},
	})
}

func vraskaDrawLoseProliferate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if err := (GainLife{Player: item.Controller, Amount: -1}).Apply(ctx); err != nil {
		return err
	}
	return Proliferate{}.Apply(ctx)
}

// vraskaTreasureify turns the still-legal target into a Treasure with no
// other abilities, for as long as it stays the same object.
func vraskaTreasureify(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ScopedEffectFor{
			Target: t.ID,
			Mods: []game.Mod{
				game.SetTypesMod([]string{"Artifact"}, []string{"Treasure"}),
				game.LoseAllAbilitiesMod(),
				game.GrantAbilitiesMod(vraskaTreasureGrant),
			},
			Duration: g.PinnedTo(game.IndefiniteDuration(), t.ID),
			Label:    "Vraska, Betrayal's Sting — a Treasure artifact",
		}.Apply(ctx)
	}
	return nil
}

// vraskaPoisonToNine tops the still-legal target player up to nine
// poison counters.
func vraskaPoisonToNine(g *game.Game, item *game.StackItem) error {
	for _, t := range NewContext(g, item).LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return vraskaTopUpPoison(g, item.Controller, t.ID)
	}
	return nil
}

func vraskaTopUpPoison(g *game.Game, placer, player uuid.UUID) error {
	p := g.PlayerByIDForEffect(player)
	if p == nil {
		return nil
	}
	have := p.Counters[game.CounterPoison]
	if have >= 9 {
		return nil
	}
	return g.AddPlayerCounterByForEffect(placer, player, game.CounterPoison, 9-have)
}
