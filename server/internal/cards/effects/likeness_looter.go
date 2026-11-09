package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Likeness Looter — Creature — Faerie Shapeshifter {U}{B}, 1/1 (Wilds of
// Eldraine, #4471):
//
//	"Flying
//	 {T}: Draw a card, then discard a card.
//	 {X}: This creature becomes a copy of target creature card in your
//	 graveyard with mana value X, except it has flying and this ability.
//	 Activate only as a sorcery."
//
// The loot is Sphinx of False Conclusions' draw-then-discard. The copy is
// Dimir Doppelganger's indefinite BecomeCopy (a copy effect with no
// duration, CR 611.2a, in layer 1, CR 707.9a), reading the card's own
// printed values (the ruling: it copies nothing of what the card was
// before it was in the graveyard), without exiling it. X is announced with
// the activation and the target clause reads it (WithManaValueEqualsX, as
// Chthonian Nightmare), checked at announce and again as the ability
// resolves; a card with {X} in its cost has X = 0 there. "Except it has
// flying and this ability" is the flying keyword and a grant of this very
// ability as a named bundle, so a Looter that has become something else
// can copy again, and its own loot is gone with the rest of the copied
// card's text, as printed. A copy made again is a new copy with a fresh
// use of any once-each-turn limit (the ruling).
//
// No simplification.
const likenessLooterGrant = "likeness-looter/this-ability"

var likenessLooterAbility = ActivatedAbility{
	Label: "{X}: This creature becomes a copy of target creature card in your graveyard with mana value X, except it has flying and this ability. Activate only as a sorcery.",
	Cost:  ManaCost("{X}"),
	Targets: TargetCardInGraveyard("target creature card in your graveyard with mana value X",
		Creature(), YouOwn()).WithManaValueEqualsX(),
	SorcerySpeed: true,
	Effect: func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			c, ok := g.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			own := game.OwnPrintedValues(c)
			return BecomeCopy{
				Targets:  []uuid.UUID{item.SourceCardID},
				Of:       t.ID,
				Values:   &own,
				Duration: CopyIndefinite,
				Except: func(v *game.PrintedValues) {
					v.AddKeyword("flying")
					v.GrantAbility(likenessLooterGrant)
				},
				Label: "Likeness Looter — becomes a copy",
			}.Apply(ctx)
		}
		return nil
	},
}

func init() {
	Register(Spec{
		OracleID:        "e04560ea-f2bd-4cb7-8f65-3aa0dd58fbdf",
		Name:            "Likeness Looter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []AbilityGrant{{
			Key:       likenessLooterGrant,
			Text:      "{X}: This creature becomes a copy of target creature card in your graveyard with mana value X, except it has flying and this ability. Activate only as a sorcery.",
			Activated: []ActivatedAbility{likenessLooterAbility},
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Draw a card, then discard a card.",
				Purpose: game.Purpose{Draws: 1, Discards: 1, Answers: game.AnswerValue},
				Cost:    TapCost(),
				Effect:  func(g *game.Game, item *game.StackItem) error { return lootOne(g, item, 1) },
			},
			likenessLooterAbility,
		},
	})
}
