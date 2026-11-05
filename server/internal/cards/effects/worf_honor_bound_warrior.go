package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Worf, Honor-Bound Warrior — Legendary Creature — Klingon Warrior
// Officer {R/W}{R/W}, 1/1 (Star Trek, #4973):
//
//	"Whenever you attack while creatures you control have total power 8
//	 or greater, double Worf's power until end of turn.
//	 Tap two untapped artifacts and/or creatures you control: Put a
//	 +1/+1 counter on Worf. Activate only as a sorcery."
//
// The attack trigger is Hologram Janeway's: "whenever you attack" is once
// per declaration however many creatures attack and whether or not Worf
// is one of them (OncePerBatch), and "while" is a condition of the event,
// not an intervening "if" (CR 603.4 needs the word "if"), so the total
// is read as attackers are declared and not again. The doubling reads
// Worf's power as the trigger resolves (CR 701.10b, DoublePowerUntilEOT);
// a Worf that has left does nothing.
//
// The activated ability's cost is a TapOthers cost of two, Collation
// Trimatrix's shape: not the {T} symbol, so Worf may tap himself as one
// of the two and a permanent that arrived this turn may pay (CR 302.6).
//
// No simplification.
func init() {
	const tapTwo = "two untapped artifacts and/or creatures you control"
	totalPowerEight := creaturesYouControlTotalPowerAtLeast(8)
	Register(Spec{
		OracleID:     "2ecda53d-ac24-4b28-aed8-2a5fb1449f85",
		Name:         "Worf, Honor-Bound Warrior",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && totalPowerEight(g, source.Controller)
			}, "Worf, Honor-Bound Warrior — double Worf's power until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return DoublePowerUntilEOT(NewContext(g, item), item.SourceCardID, "Worf, Honor-Bound Warrior — double power")
				})),
		},
		Activated: []ActivatedAbility{{
			Label: "Tap two untapped artifacts and/or creatures you control: Put a +1/+1 counter on Worf. Activate only as a sorcery.",
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: TargetPermanent(tapTwo, Or(Artifact(), Creature())),
				Label:  tapTwo,
			}},
			SorcerySpeed: true,
			Effect:       Do(CounterOnThis{Kind: "+1/+1", N: 1}),
		}},
	})
}
