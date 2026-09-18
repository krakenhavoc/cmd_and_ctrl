package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unruly Catapult — Artifact Creature — Construct {2}{R}, 0/4 (EDHREC
// rank 4288):
//
//	"Defender
//	 {T}: This creature deals 1 damage to each opponent.
//	 Whenever you cast an instant or sorcery spell, untap this
//	 creature."
//
// A wall that turns a spellslinger deck's turn into damage: every
// instant or sorcery untaps it, so the tap ability fires once per
// spell rather than once per turn. In a three-opponent game that is 3
// damage a spell, which is why the Catapult is a Commander card and
// was a draft common.
//
// The untap is a TRIGGER, not a cost reduction or a static, and the
// ordering is the card: casting the spell puts the untap on the stack
// ABOVE it, so the untap resolves first and the Catapult can be tapped
// again in response to your own spell. That is how a storm turn gets
// more than one activation out of it.
//
// Defender is a canonical keyword read by the attack-declaration gate,
// so the 0/4 stays home; nothing about the tap ability requires
// attacking, which is the point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b2b0a5d0-0084-43a4-b6dc-e893ef42cdb7",
		Name:            "Unruly Catapult",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label: "{T}: This creature deals 1 damage to each opponent",
			Cost:  TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()), "Unruly Catapult — untap this creature",
				func(g *game.Game, item *game.StackItem) error {
					return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
	})
}
