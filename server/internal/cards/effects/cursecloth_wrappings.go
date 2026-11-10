package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cursecloth Wrappings — Artifact {2}{B}{B} (Aetherdrift, #4464):
//
//	"Zombies you control get +1/+1.
//	 {T}: Target creature card in your graveyard gains embalm until end of
//	 turn. The embalm cost is equal to its mana cost. (Exile that card and
//	 pay its embalm cost: Create a token that's a copy of it, except it's a
//	 white Zombie in addition to its other types and has no mana cost.
//	 Embalm only as a sorcery.)"
//
// The anthem is Death Baron's. The grant is Q's (ADR 0219) given to a card in
// a graveyard: an ordinary graveyard AbilityGrant whose one row is the shared
// embalm body (Embalm) with its mana component read off the card that has
// the row (ItsManaCost, game.AbilityCost.ManaFromSource), and a card-level
// grant that the cleanup sweep takes at end of turn (CR 514.2) and a zone
// change takes at once (CR 400.7), the embalm's own exile cost included.
// The grant is to that card object, so a card that left the graveyard and
// came back in response is not given it (CR 608.2b). A card with no mana cost
// would owe a cost it can't pay (CR 118.6), so it is given nothing.
//
// No simplification.
func init() {
	embalm := Embalm("")
	embalm.Label = "Embalm — its mana cost, Exile this card from your graveyard: Create a token that's a copy of it, " +
		"except it's a white Zombie in addition to its other types and it has no mana cost. Activate only as a sorcery."
	embalm.Cost = Plus(ItsManaCost(), ExileThis())
	Register(Spec{
		OracleID:     "d9c4996a-35eb-4d38-9680-57d9d039ed82",
		Name:         "Cursecloth Wrappings",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       cursecloth,
			Zone:      game.ZoneGraveyard,
			Activated: []ActivatedAbility{embalm},
			Text:      "Embalm. The embalm cost is equal to its mana cost.",
		}},
		Static: []game.StaticAbility{
			b16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					target.HasSubtype(source.TextSubtype("Zombie"))
			}, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature card in your graveyard gains embalm until end of turn. The embalm cost is equal to its mana cost.",
			Cost:    TapCost(),
			Targets: TargetCardInGraveyard("target creature card in your graveyard", Creature(), YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if c, ok := g.LookupCardForEffect(t.ID); !ok || c.ManaCost == "" {
						return nil
					}
					g.GrantAbilityToGraveyardCardUntilEndOfTurnForEffect(t.ID, ctx.Source(), ctx.Controller(), cursecloth)
				}
				return nil
			},
		}},
	})
}

// cursecloth is the bundle key of the embalm the Wrappings give. On disk.
const cursecloth = "cursecloth-wrappings/embalm"
