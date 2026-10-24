package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Karn, the Great Creator — Legendary Planeswalker — Karn {4}, loyalty 5:
//
//	"Activated abilities of artifacts your opponents control can't be
//	 activated.
//	 +1: Until your next turn, up to one target noncreature artifact
//	 becomes an artifact creature with power and toughness each equal to
//	 its mana value.
//	 −2: You may reveal an artifact card you own from outside the game or
//	 choose a face-up artifact card you own in exile. Put that card into
//	 your hand."
//
// The static is OpponentsSourcesCantActivate over artifacts with no mana
// exemption, so an opponent's Sol Ring stops tapping for mana too, as
// printed. The +1 is a scoped effect on the target: a creature type added
// in layer 4 and a base power and toughness set in layer 7b, fixed at the
// artifact's mana value as the ability resolves, until Karn's controller's
// next turn. The −2 is Coax from the Blind Eternities' shape for artifacts
// (TakeFromExileOrOutside, ADR 0198).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "a20dd48d-d344-4db1-b0e9-a2b71c3cc9d1",
		Name:         "Karn, the Great Creator",
		Completeness: CompletenessFull,
		ActivationRestrictions: []game.ActivationRestriction{
			OpponentsSourcesCantActivate(
				"Karn, the Great Creator — activated abilities of artifacts your opponents control can't be activated.",
				Artifact(), false),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Until your next turn, up to one target noncreature artifact becomes an artifact creature with power and toughness each equal to its mana value.",
				Cost:    LoyaltyCost(1),
				Targets: TargetPermanent("up to one target noncreature artifact", And(Artifact(), Noncreature())).WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						c, ok := g.LookupCardForEffect(t.ID)
						if !ok {
							continue
						}
						mv := c.ManaValue()
						return ScopedEffectFor{
							Target:   t.ID,
							Mods:     append([]game.Mod{game.AddTypesMod("Creature")}, game.SetBasePTMods(mv, mv)...),
							Duration: DurationUntilYourNextTurn(ctx, item.Controller),
							Label:    "Karn, the Great Creator — an artifact creature",
						}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label: "−2: You may reveal an artifact card you own from outside the game or choose a face-up artifact card you own in exile. Put that card into your hand.",
				Cost:  LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return TakeFromExileOrOutside{
						Pick:          karnTheGreatCreatorTake,
						Exile:         karnTheGreatCreatorExile,
						ExileQuestion: "Karn, the Great Creator — choose a face-up artifact card you own in exile, or none to reveal one from outside the game",
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// karnTheGreatCreatorTake is the −2's pick from outside the game, keyed
// (#4256), and the artifact filter its exile pick shares.
var karnTheGreatCreatorTake = registerOutsideTake("choose-cards/karn-the-great-creator-take",
	func(c game.Card) bool { return c.IsArtifact() }, true)

// karnTheGreatCreatorExile is the −2's pick from exile, keyed (#4264): a
// table waiting on it is a restore point. "None" asks the outside pick.
var karnTheGreatCreatorExile = registerExileOrOutside("choose-cards/karn-the-great-creator-exile",
	karnTheGreatCreatorTake, "Karn, the Great Creator — reveal an artifact card you own from outside the game, or none")
