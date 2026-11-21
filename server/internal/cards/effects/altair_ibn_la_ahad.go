package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Altaïr Ibn-La'Ahad — Legendary Creature — Human Assassin {R}{W}{B}, 3/3
// (Assassin's Creed, #5948):
//
//	"First strike
//	 Whenever Altaïr attacks, exile up to one target Assassin creature card
//	 from your graveyard with a memory counter on it. Then for each creature
//	 card you own in exile with a memory counter on it, create a tapped and
//	 attacking token that's a copy of it. Exile those tokens at end of
//	 combat."
//
// The exile is Mari, the Killing Quill's: the counter waits for the exile
// to happen (a commander may be sent to the command zone instead, and then
// gets none), and the second sentence runs after it whether or not there was
// a target ("up to one"). "Each creature card you own in exile with a memory
// counter" is every such card, from this and earlier attacks and from any
// other source of memory counters, read once as the sentence resolves; each
// gets a copy token (CR 707.2, copiable values) that enters tapped and
// attacking, with its attack target chosen by you (CR 508.4), and the token
// carries Delina, Wild Mage's end-of-combat exile. The cards stay in exile
// with their counters, so every later attack copies them again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "43dc4b0c-e437-4abc-a3fb-48bd3098de88",
		Name:            "Altaïr Ibn-La'Ahad",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Grants: []AbilityGrant{{
			Key:  altairExileGrant,
			Text: "At end of combat, exile this token.",
			Triggered: []game.TriggeredAbility{
				AtEachStep(game.StepEndCombat, "Altaïr's token — exile this token at end of combat", func(g *game.Game, item *game.StackItem) error {
					return ExileTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
			},
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventAttack, ThisAttacked,
				"Altaïr Ibn-La'Ahad — exile up to one target Assassin creature card from your graveyard with a memory counter on it; then for each creature card you own in exile with a memory counter on it, create a tapped and attacking token that's a copy of it; exile those tokens at end of combat",
				altairAttack),
				TargetCardInGraveyard("up to one target Assassin creature card from your graveyard",
					YouOwn(), Creature(), HasSubtype("Assassin")).WithCount(0, 1)),
		},
	})
}

// altairExileGrant is the copies' end-of-combat exile. An on-disk key:
// never renamed, never reused.
const altairExileGrant = "altair/exile-at-end-of-combat"

func altairAttack(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if id, ok := b16FirstLegalTargetCard(ctx); ok {
		return ExileTarget{
			Target:   id,
			ThenTail: withCarry(ctx.ItemTail(altairExiled), id),
		}.Apply(ctx)
	}
	return altairCopies(ctx)
}

// altairExiled is the clause after the exile: the memory counter, then the
// copies. An on-disk key.
var altairExiled = CardTail("altair-ibn-la-ahad/after-exile", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 1 {
		if err := (AddCounter{Target: tailCarry0(o), Kind: altairMemoryCounter, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return altairCopies(ctx)
})

const altairMemoryCounter = "memory"

// altairCopies creates a tapped and attacking copy of each creature card the
// controller owns in exile with a memory counter on it.
func altairCopies(ctx *Context) error {
	g := ctx.Game
	if g.Exile == nil {
		return nil
	}
	var cards []game.Card
	for _, c := range g.Exile.Cards {
		if c.Owner == ctx.Controller() && c.Counters[altairMemoryCounter] > 0 && c.IsCreature() {
			cards = append(cards, c)
		}
	}
	for _, c := range cards {
		tmpl, ok := TokenCopyTemplate(g, c.InstanceID)
		if !ok {
			continue
		}
		TokenCopyGainsAbility(&tmpl, altairExileGrant)
		if err := (CreateTokensAttackingYourChoice{
			Template: tmpl,
			Copy:     true,
			N:        1,
			Tapped:   true,
			Label:    "Altaïr Ibn-La'Ahad",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
