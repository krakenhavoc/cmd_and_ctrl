package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anikthea, Hand of Erebos — Legendary Enchantment Creature — Demigod
// {2}{W}{B}{G}, 4/4 (Commander Masters):
//
//	"Menace
//	 Other enchantment creatures you control have menace.
//	 Whenever Anikthea enters or attacks, exile up to one target non-Aura
//	 enchantment card from your graveyard. Create a token that's a copy of
//	 that card, except it's a 3/3 black Zombie creature in addition to its
//	 other types."
//
// The trigger is The Scarab God's shape: the target is chosen when the
// trigger goes on the stack, the exile happens as it resolves, and only a
// card that really reached exile is copied (CR 707.2, copied as the card it
// is in exile). The exception is part of the copiable values (CR 707.9b):
// 3/3, black, a Zombie creature on top of the card's own types; a copied
// Saga or Class keeps what it is. The grant reads effective types, so an
// animated enchantment counts for the menace.
//
// No simplification.
func init() {
	otherEnchantmentCreatures := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return OtherCreaturesYouControl(target, g, source) && target.IsEnchantment()
	}
	Register(Spec{
		OracleID:        "991fd651-0cba-41ee-855a-d01a1a7c8f60",
		Name:            "Anikthea, Hand of Erebos",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Purpose:         game.Purpose{Tokens: 1},
		Static: []game.StaticAbility{
			Labelled(b16GrantKeywords(otherEnchantmentCreatures, "menace"), "Other enchantment creatures you control have menace."),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEntersOrAttacks("Anikthea, Hand of Erebos — exile up to one target non-Aura enchantment card from your graveyard and copy it as a 3/3 black Zombie creature",
				aniktheaExileAndCopy),
				TargetCardInGraveyard("up to one target non-Aura enchantment card from your graveyard",
					YouOwn(), And(Enchantment(), Not(Subtype("Aura")))).WithCount(0, 1)),
		},
	})
}

func aniktheaExileAndCopy(g *game.Game, item *game.StackItem) error {
	return ExileTargetThenTokenCopy(g, item, aniktheaZombieCopy)
}

var aniktheaZombieCopy = ExileThenTokenCopyTail("exile-then-token-copy/anikthea-hand-of-erebos", func(ctx *Context) func(t *game.Card) {
	reader := ctx.Game.TextReaderForEffect(ctx.Item)
	return func(t *game.Card) {
		v := game.PrintedValues{TypeLine: t.TypeLine}
		v.AddCardType("Creature")
		for _, sub := range reader.TextSubtypes("Zombie") {
			v.AddSubtype(sub)
		}
		t.TypeLine = v.TypeLine
		t.Power, t.Toughness = 3, 3
		t.VariableToughness = false
		t.PrintedPTKnown = true
		t.SetCopyExceptionColors(reader.TextColors("B")...)
	}
})
