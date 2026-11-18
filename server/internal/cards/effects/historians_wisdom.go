package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Historian's Wisdom — Enchantment — Aura {2}{G} (Neon Dynasty, #5493):
//
//	"Enchant artifact or creature
//	 When this Aura enters, if enchanted permanent is a creature with the
//	 greatest power among creatures on the battlefield, draw a card.
//	 As long as enchanted permanent is a creature, it gets +2/+1."
//
// The enters trigger has an intervening "if" (CR 603.4): it is checked as
// the Aura enters, and again as the trigger resolves. The creature must have
// power at least equal to every creature on the battlefield (a tie for
// greatest counts), read through the layers, so the +2/+1 this Aura gives is
// already in the comparison.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c2c460d-8059-41ff-b2d7-374aa498f011",
		Name:         "Historian's Wisdom",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant artifact or creature", Or(Artifact(), Creature())),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static:       []game.StaticAbility{b43PumpAttachedWhen(cmdr2B19AttachedIsCreature, 2, 1)},
		Triggered: []game.TriggeredAbility{
			historiansWisdomDraw(),
		},
	})
}

// historiansWisdomDraw is the enters trigger: its AppliesTo is the check as
// it triggers (the Aura's current host), InterveningIf the one as it
// resolves.
func historiansWisdomDraw() game.TriggeredAbility {
	t := WhenThisEnters("Historian's Wisdom — draw a card", Do(DrawCards{N: 1}))
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return when(ev, source, lki, g) && source.AttachedTo.Kind == game.TargetCard &&
			neoB12HostHasGreatestPower(g, source.AttachedTo.ID)
	}
	return InterveningIf(neoB12EnchantedHasGreatestPower, t)
}
