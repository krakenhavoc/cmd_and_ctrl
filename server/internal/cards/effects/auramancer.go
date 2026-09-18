package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Auramancer — Creature — Human Wizard {2}{W}, 2/2 (EDHREC rank
// 4312):
//
//	"When this creature enters, you may return target enchantment card
//	 from your graveyard to your hand."
//
// The enchantment deck's Eternal Witness, and the reason it is played
// over the green one is that it comes back: a blink deck loops the
// Auramancer and rebuys the same Ghostly Prison every turn.
//
// Two "you may"s in one sentence and only one of them is a choice the
// engine has to ask about. "You MAY return" is the optional TRIGGER
// prompt (CR 603.3c — the choice is made as the ability resolves, and
// the target is still chosen when it goes on the stack). The target
// clause is mandatory once the trigger is on the stack, so a yard with
// an enchantment card in it gets a prompt and one without never puts
// the trigger on the stack at all.
//
// "Your graveyard" is YouOwn, not YouControl: a graveyard card has no
// controller, and the printed word is whose graveyard it is in.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bd5eb181-6a69-4dc2-93a0-fa000291bc3d",
		Name:         "Auramancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					WhenThisEnters("Auramancer — return an enchantment card to your hand",
						returnFirstLegalGraveyardTargetToHand),
					"Auramancer — return an enchantment card from your graveyard to your hand?"),
				TargetCardInGraveyard("target enchantment card in your graveyard", YouOwn(), Enchantment())),
		},
	})
}
