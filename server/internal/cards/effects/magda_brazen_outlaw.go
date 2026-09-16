package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magda, Brazen Outlaw — Legendary Creature — Dwarf Berserker {1}{R},
// 2/1 (EDHREC rank 1235):
//
//	"Other Dwarves you control get +1/+0.
//	 Whenever a Dwarf you control becomes tapped, create a Treasure
//	 token.
//	 Sacrifice five Treasures: Search your library for an artifact or
//	 Dragon card, put that card onto the battlefield, then shuffle."
//
// Two of the three abilities are here whole. The anthem is a Layer
// 7c static over other Dwarves (effective subtypes, so a changeling
// counts). The Treasure trigger reads two event kinds, because the
// engine taps an attacker without an EventTapCard — see
// b11DwarfYouControlBecameTapped for why an EventAttack whose
// creature is now tapped is "became tapped", and why a vigilance
// Dwarf is not. Magda herself is a Dwarf, so her own attack pays.
//
// SANDBOX GAP, weaker than printed: the third ability is omitted.
// "Sacrifice five Treasures" is a five-permanent sacrifice cost, and
// AbilityCost.SacrificeOther pays exactly one permanent (Stoneforge
// Mystic's posture: a cost with no shape leaves the ability out
// rather than shipping a cheaper one). The card is still recognisably
// Magda — the lord and the Treasure engine are what every Dwarf deck
// plays her for.
func init() {
	Register(Spec{
		OracleID:     "3d268e48-3004-4384-bf52-e63243fb5e02",
		Name:         "Magda, Brazen Outlaw",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The five-Treasure tutor isn't implemented — Magda pumps Dwarves and makes Treasures only."},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID &&
					target.Controller == source.Controller &&
					target.IsCreature() && target.HasSubtype("Dwarf")
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
			},
		}},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b11DwarfYouControlBecameTapped(ev, source, g)
			}, "Magda, Brazen Outlaw — create a Treasure", Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
