package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zur, Eternal Schemer — Legendary Creature — Human Wizard {W}{U}{B}, 1/4
// (Dominaria United):
//
//	"Flying
//	 Enchantment creatures you control have deathtouch, lifelink, and
//	 hexproof.
//	 {1}{W}: Target non-Aura enchantment you control becomes a creature in
//	 addition to its other types and has base power and toughness each equal
//	 to its mana value."
//
// The three keywords are layer-6 grants to the enchantment creatures the
// controller has (Zur itself is no enchantment). The activated ability has no
// duration: it adds the Creature type (layer 4) and sets the base power and
// toughness (layer 7b) as a scoped effect pinned to that object, lasting until
// it leaves the battlefield, as Karn's animation does for a stated duration.
// The mana value is read once, as the ability resolves (CR 608.2h); a target
// that is no longer a legal non-Aura enchantment you control then does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cfc5dc9c-6f26-460b-a624-3681b8584eb3",
		Name:            "Zur, Eternal Schemer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			KeywordGrant(zurEnchantmentCreatureYouControl, "deathtouch"),
			KeywordGrant(zurEnchantmentCreatureYouControl, "lifelink"),
			KeywordGrant(zurEnchantmentCreatureYouControl, "hexproof"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: Target non-Aura enchantment you control becomes a creature in addition to its other types and has base power and toughness each equal to its mana value.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{1}{W}"),
			Targets: TargetPermanent("target non-Aura enchantment you control", Enchantment(), Not(OfSubtype("Aura")), YouControl()),
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
						Duration: g.PinnedTo(game.IndefiniteDuration(), t.ID),
						Label:    "Zur, Eternal Schemer — becomes a creature with base power and toughness equal to its mana value",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}

// zurEnchantmentCreatureYouControl is "enchantment creatures you control".
func zurEnchantmentCreatureYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.IsCreature() && target.IsEnchantment()
}
