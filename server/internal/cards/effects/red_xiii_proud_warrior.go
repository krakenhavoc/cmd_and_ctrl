package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Red XIII, Proud Warrior — Legendary Creature — Beast Warrior
// {1}{R}{G}, 3/3 (#3656):
//
//	"Vigilance, trample
//	 Other modified creatures you control have vigilance and trample.
//	 (Equipment, Auras you control, and counters are modifications.)
//	 Cosmo Memory — When Red XIII enters, return target Aura or
//	 Equipment card from your graveyard to your hand."
//
// Kodama of the West Tree's modified reading (b07IsModified) as a
// layer 6 grant over the OTHER creatures you control; the layer
// listener recomputes on attach and on counters, so a creature that
// gains its first counter has the keywords on the same pass. The
// return is Archaeomancer's targeted ETB over the graveyard clause.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9d302e2a-d852-4878-ac9d-d51d9924144b",
		Name:            "Red XIII, Proud Warrior",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "Other modified creatures you control have vigilance and trample.",
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.IsCreature() &&
					target.Controller == source.Controller && b07IsModified(g, *target)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, kw := range []string{"vigilance", "trample"} {
					if !eotHasAbility(c.Abilities, kw) {
						c.Abilities = append(c.Abilities, kw)
					}
				}
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets: TargetCardInGraveyard("target Aura or Equipment card in your graveyard",
				Or(OfSubtype("Aura"), OfSubtype("Equipment")), YouOwn()),
			Key:    "Red XIII, Proud Warrior — Cosmo Memory: return an Aura or Equipment card to hand",
			Effect: returnTargetCardToHand,
		}},
	})
}
