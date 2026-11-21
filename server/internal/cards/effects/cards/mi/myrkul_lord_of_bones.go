package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Myrkul, Lord of Bones — {4}{W}{B}{G} Legendary Creature — God 7/5
//
//	"As long as your life total is less than or equal to half your starting
//	 life total, Myrkul has indestructible.
//	 Whenever another nontoken creature you control dies, you may exile it.
//	 If you do, create a token that's a copy of that card, except it's an
//	 enchantment and loses all other card types."
//
// Nightmare Shepherd's shape: the card is exiled from the graveyard (only if
// it is still that object there), and the token copies it as it last existed
// with its card types replaced by Enchantment (CR 707.9b, 205.3d: creature
// subtypes go with the creature type), so the copy is not a creature, and
// keeps the copied abilities.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11329dd8-b10b-4659-b747-fbdbecf16308",
		Name:         "Myrkul, Lord of Bones",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GapB214SelfIndestructibleAtHalfLife("Myrkul")},
		Triggered: []game.TriggeredAbility{
			effects.Cmdr2S47ExileThenCopyTrigger(
				"Myrkul, Lord of Bones — exile it; if you do, create a token copy that's an enchantment",
				"Myrkul, Lord of Bones — exile that creature card and create a token copy of it that's an enchantment?",
				"myrkul-lord-of-bones/copy",
				func(t *game.Card) {
					v := game.PrintedValues{TypeLine: t.TypeLine, GrantedAbilities: t.GrantedAbilities}
					v.SetCardTypes("Enchantment")
					t.TypeLine = v.TypeLine
					t.GrantedAbilities = v.GrantedAbilities
				}),
		},
	})
}
