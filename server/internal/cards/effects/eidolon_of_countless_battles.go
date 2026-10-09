package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eidolon of Countless Battles — Enchantment Creature — Spirit
// {1}{W}{W}, 0/0 (EDHREC rank 3426):
//
//	"Bestow {2}{W}{W} (If you cast this card for its bestow cost, it's
//	 an Aura spell with enchant creature. It becomes a creature again
//	 if it's not attached.)
//	 This creature and enchanted creature each get +1/+1 for each
//	 creature you control and +1/+1 for each Aura you control."
//
// The Voltron Aura deck's finisher: bestowed, it counts itself as an
// Aura, and when the host dies it stays as a creature that counts
// itself as one (ADR 0141, #2862). Both counts are live reads at layer
// 7c, after layer 4 has settled what is a creature and what is an
// Aura, and "you" is its controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "441b51b0-ecd0-466d-9ca8-e754eb563aec",
		Name:         "Eidolon of Countless Battles",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Bestow("{2}{W}{W}"),
		},
		Static: []game.StaticAbility{
			PumpSelfCreatureOrAttachedPer(1, 1, creaturesAndAurasYouControl),
		},
	})
}

// creaturesAndAurasYouControl is Eidolon of Countless Battles' count:
// one for each creature its controller controls plus one for each Aura
// they control. A permanent that is both counts twice, as the printed
// sum does.
func creaturesAndAurasYouControl(g *game.Game, source *game.Card) int {
	return permanentsYouControl(g, source, func(c *game.Card) bool { return c.IsCreature() }) +
		permanentsYouControl(g, source, func(c *game.Card) bool { return c.IsAura() })
}
