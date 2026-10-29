package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beorn's Hospitality — Enchantment {1}{G} (The Hobbit):
//
//	"Landfall — Whenever a land you control enters, put a +1/+1 counter on
//	 target creature you control.
//	 {5}{G}{G}: This enchantment becomes a Bear creature in addition to its
//	 other types and gains "This creature's power and toughness are each
//	 equal to the number of lands you control." (This effect doesn't end.)"
//
// The landfall trigger is Necklace of Girion's targeted counter on the shared
// Landfall. The activated ability is one effect with no duration (CR 611.2a),
// pinned to this object (CR 400.7), so a Hospitality that leaves and returns
// is an enchantment again: the Creature type and the Bear subtype in layer 4
// (the subtype is a word of the card's text, CR 612.2), and the granted
// static in layers 6 and 7b. The static is a granted ability, not a
// characteristic-defining one (CR 604.3: a CDA is printed on its object), so
// it sets power and toughness in layer 7b, where a later effect's set or a
// +1/+1 counter lands on top, and it stops with the ability if the creature
// loses its abilities. Lands are counted live, every land you control, as the
// layer pass recomputes. It stays an enchantment as well (in addition to its
// other types), and a second activation only stacks another copy of the same
// grant.
//
// No simplification.
func init() {
	const grantKey = "beorns-hospitality/power-toughness-lands"
	landfall := Landfall("Beorn's Hospitality — put a +1/+1 counter on target creature you control", putPlusOneCounterOnEachLegalTarget)
	landfall.Targets = TargetCreature("target creature you control", YouControl())
	Register(Spec{
		OracleID:     "d9ed7252-11d1-432b-9101-ac08cd28826d",
		Name:         "Beorn's Hospitality",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: grantKey,
			Static: []game.StaticAbility{{
				Label:    "This creature's power and toughness are each equal to the number of lands you control.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7B_Set,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return source != nil && target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					n := b10LandsControlled(g, source.Controller)
					c.Power, c.Toughness = n, n
				},
			}},
			Text: "This creature's power and toughness are each equal to the number of lands you control.",
		}},
		Triggered: []game.TriggeredAbility{landfall},
		Activated: []ActivatedAbility{{
			Label:   "{5}{G}{G}: This enchantment becomes a Bear creature in addition to its other types and gains \"This creature's power and toughness are each equal to the number of lands you control.\" (This effect doesn't end.)",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{5}{G}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				self := item.SourceCardID
				if !onBattlefield(g, self) || sourceIsNewObject(g, item) {
					return nil
				}
				return GrantAbilitiesFor{
					Target: self,
					Keys:   []string{grantKey},
					Also: []game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(g.TextReaderForEffect(item).TextSubtypes("Bear")...),
					},
					Duration: g.PinnedTo(game.IndefiniteDuration(), self),
					Label:    "Beorn's Hospitality — becomes a Bear creature with power and toughness equal to your lands",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
