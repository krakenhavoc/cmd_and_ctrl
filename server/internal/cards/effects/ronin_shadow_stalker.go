package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ronin, Shadow Stalker — {2}{B} Legendary Creature — Human Rogue Hero 3/3:
//
//	"Pay 2 life: Add two mana of any one color. Spend this mana only to
//	 cast Equipment spells or activate equip abilities. Activate only once
//	 each turn.
//	 {T}, Sacrifice an Equipment attached to Ronin: Target creature gets
//	 -4/-4 until end of turn. Activate only as a sorcery."
//
// The mana ability's cost is life alone (CR 119.4, validated before
// anything is paid), so it taps nothing and a summoning-sick Ronin can
// fire it. "Two mana of any one color" is OneColorOfAmount(2): one colour
// pick, two mana of it. "Activate only once each turn" is the declared
// OncePerTurn bit (#1621), per object and label, refused by the click
// path, the enumerator and the auto-tapper alike.
//
// "Spend this mana only to cast Equipment spells or activate equip
// abilities" is one restriction with two halves (#4600, CR 106.6): an
// Equipment spell (ManaRestrictCast with the Equipment subtype) or the
// cost of an equip ability (ManaRestrictEquip, matched against
// ManaSpendContext.Equip, which the activation, the legal-move enumerator
// and the cost preview set from the row). It can't pay a creature spell,
// a non-Equipment artifact spell, or an Equipment's other activated
// abilities: those are activations whose source is an Equipment, which
// the source's types alone could not tell from an equip ability.
//
// The second ability is a sorcery-speed activated ability whose cost taps
// Ronin and sacrifices an Equipment attached to him (AttachedToThis,
// Faunsbane Troll's clause), so it needs one on him.
//
// No simplification. Marvel Super Heroes (#4455, seam #4600).
func init() {
	Register(Spec{
		OracleID:     "8ceb1acf-5ff3-4f1a-af04-9f07ad454526",
		Name:         "Ronin, Shadow Stalker",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Life: 2},
			Produced: OneColorOfAmount(2),
			Label:    "Pay 2 life: Add two mana of any one color. Spend this mana only to cast Equipment spells or activate equip abilities. Activate only once each turn.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Equipment")},
				[]string{game.ManaRestrictEquip},
			)},
			OncePerTurn: true,
		}},
		Activated: []ActivatedAbility{{
			Label: "{T}, Sacrifice an Equipment attached to Ronin: Target creature gets -4/-4 until end of turn. Activate only as a sorcery.",
			Cost: Plus(TapCost(), game.AbilityCost{
				SacrificeOther: AttachedToThis("an Equipment attached to Ronin", HasSubtype("Equipment")),
			}),
			Targets:      TargetCreature("target creature"),
			SorcerySpeed: true,
			Purpose:      ForTargets(HarmUntilEOT(0, -4, -4)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return BoostUntilEOT{Target: t.ID, Power: -4, Toughness: -4,
							Label: "Ronin, Shadow Stalker — -4/-4"}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
