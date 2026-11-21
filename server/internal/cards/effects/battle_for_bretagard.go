package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Battle for Bretagard — Enchantment — Saga {1}{G}{W} (Commander Masters,
// #5948):
//
//	"I — Create a 1/1 white Human Warrior creature token.
//	 II — Create a 1/1 green Elf Warrior creature token.
//	 III — Choose any number of artifact tokens and/or creature tokens you
//	 control with different names. For each of them, create a token that's a
//	 copy of it."
//
// Chapter III asks you for a set of your own tokens (artifact tokens and
// creature tokens, so an artifact creature token counts once) whose names
// are all different; the engine refuses a pick with two of the same name
// (a set rule, CR 608.2d). Every copy is created at once, one per chosen
// token (CR 707.2 copiable values), as a single creation event.
//
// No simplification.
func init() {
	const name = "Battle for Bretagard"
	Register(Spec{
		OracleID:     "8294539d-7982-4ec9-8392-99eed78542f5",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "create a 1/1 white Human Warrior creature token"),
				Do(CreateToken{Template: TokenCard("1/1 white Human Warrior"), N: 1})),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "create a 1/1 green Elf Warrior creature token"),
				Do(CreateToken{Template: TokenCard("1/1 green Elf Warrior"), N: 1})),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "choose any number of artifact tokens and/or creature tokens you control with different names; for each of them, create a token that's a copy of it"),
				func(g *game.Game, item *game.StackItem) error {
					return ChoosePermanents{
						Question:   "Battle for Bretagard — choose any number of artifact and/or creature tokens you control with different names",
						Candidates: gap110BretagardCandidates,
						Keyed:      gap110BretagardPick,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// gap110BretagardCandidates offers the chooser's artifact tokens and
// creature tokens, bounded by how many different names there are.
func gap110BretagardCandidates(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	var ids []uuid.UUID
	names := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != of || !IsToken(c) || !(c.IsCreature() || c.HasCardType("artifact")) {
			continue
		}
		ids = append(ids, c.InstanceID)
		names[c.Name] = true
	}
	if len(ids) == 0 {
		return nil, 0, 0
	}
	return ids, 0, len(names)
}

var gap110BretagardPick = RegisterPermanentsPick("prompt-run/battle-for-bretagard-copies", PermanentsDef{
	Validate: func(picked []game.Card) bool {
		seen := map[string]bool{}
		for _, c := range picked {
			if seen[c.Name] {
				return false
			}
			seen[c.Name] = true
		}
		return true
	},
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		cards := picked.Cards()
		if len(cards) == 0 {
			return nil
		}
		return CreateTokenCopiesTogether{Controller: ctx.Controller(), Copies: cards}.Apply(ctx)
	},
})
