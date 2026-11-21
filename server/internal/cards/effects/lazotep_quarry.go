package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lazotep Quarry — Land — Desert (Modern Horizons 3 Commander, #5948):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice a creature: Add one mana of any color.
//	 {X}{2}, {T}, Sacrifice a Desert: Exile target creature card with mana
//	 value X from your graveyard. Create a token that's a copy of it, except
//	 it's a 4/4 black Zombie. Activate only as a sorcery."
//
// The second mana ability sacrifices a creature as part of its cost (CR
// 605.1a, it is still a mana ability). The third is the Scarab God's
// exile-then-token-copy with a mana-value-X target: X is announced with the
// activation (CR 107.3a) and the target clause reads it (WithManaValueEqualsX),
// checked at announce and again as the ability resolves (CR 608.2b). The Desert
// sacrificed may be the Quarry itself, which is a Desert. The copy is a 4/4
// black Zombie that REPLACES the copied creature's creature types (CR 707.9b,
// scarabGodZombieExceptionFor); the exile and the copy are one instruction, so
// a card that was not exiled makes no token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0d2fa39a-9cac-4a1f-bb1e-b6162e6d5169",
		Name:         "Lazotep Quarry",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, SacrificeOther: SacrificeACreature().SacrificeOther},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}, Sacrifice a creature: Add one mana of any color",
				Answers:  game.AnswerSacOutlet,
			},
		},
		Activated: []ActivatedAbility{{
			Label: "{X}{2}, {T}, Sacrifice a Desert: Exile target creature card with mana value X from your graveyard. " +
				"Create a token that's a copy of it, except it's a 4/4 black Zombie. Activate only as a sorcery.",
			Cost: Plus(ManaCost("{X}{2}"), TapCost(), SacrificeN(1, "a Desert", HasSubtype("Desert"))),
			Targets: TargetCardInGraveyard("target creature card with mana value X from your graveyard",
				Creature(), YouOwn()).WithManaValueEqualsX(),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ExileTargetThenTokenCopy(g, item, lazotepQuarryZombieCopy)
			},
		}},
	})
}

// lazotepQuarryZombieCopy is the exile's keyed "then": the token copy, a
// 4/4 black Zombie.
var lazotepQuarryZombieCopy = ExileThenTokenCopyTail("exile-then-token-copy/lazotep-quarry", func(ctx *Context) func(t *game.Card) {
	return scarabGodZombieExceptionFor(ctx.Game.TextReaderForEffect(ctx.Item))
})
