package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sugar Coat — Enchantment — Aura {2}{U} (Bloomburrow, #4467):
//
//	"Flash
//	 Enchant creature or Food
//	 Enchanted permanent is a colorless Food artifact with "{2}, {T},
//	 Sacrifice this artifact: You gain 3 life" and loses all other card
//	 types and abilities."
//
// One continuous effect in three layers, as Song of the Dryads is (CR 613.6:
// it starts in layer 4, so the later halves keep applying even if the Aura's
// own abilities are taken away): layer 4 sets the card types to Artifact and
// the subtypes to Food (so the host stops being a creature and loses its
// creature subtypes; supertypes stay, CR 205.4b), layer 5 makes it colorless,
// layer 6 removes every ability and grants the Food's own (Supper for
// Spiders' bundle: the same printed "{2}, {T}, Sacrifice this artifact: You
// gain 3 life"). The host keeps its name, mana cost and mana value, its
// counters and token-ness (the ruling). "Enchant creature or Food" is the
// legality CR 704.5m re-checks: a creature turned into a Food is still a Food,
// so the Aura stays; Equipment on the former creature unattaches (CR 704.5n),
// and Auras that can't enchant a Food go to the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1caa8c8a-ff45-4eee-a834-6e5e1fb54721",
		Name:            "Sugar Coat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         TargetPermanent("enchant creature or Food", Or(Creature(), Subtype("Food"))),
		Static: []game.StaticAbility{
			{
				Label:                 "Enchanted permanent is a Food artifact and loses all other card types.",
				Layer:                 game.Layer4Type,
				ContinuesAfterRemoval: true,
				AppliesTo:             AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					c.Types = []string{"Artifact"}
					c.SetSubtypes(nil)
					c.Subtypes = append(c.Subtypes, source.TextSubtypes("Food")...)
				},
			},
			SetAttachedColors(),
			LosesAllAbilitiesAndHas(AttachedToSource, supperForSpidersFood),
		},
	})
}
