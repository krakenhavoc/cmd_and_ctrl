package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brash Taunter — Creature — Goblin {4}{R}, 1/1:
//
//	"Indestructible
//	 Whenever this creature is dealt damage, it deals that much damage
//	 to target opponent.
//	 {2}{R}, {T}: This creature fights another target creature."
//
// Indestructible rides PrintedKeywords. The trigger is Screaming
// Nemesis's shape (b35SelfWasDealtDamage) with Wrathful Red Dragon's
// body (damagedCreatureReflectsChosenAmount) aimed at "target
// opponent": the controller picks the opponent as the trigger goes on
// the stack, and the Taunter, as the source, deals the event's amount
// on resolution. Its colour and keywords are what a prevention or
// lifelink check sees.
//
// The fight is Polukranos's (b30SourceFightsFirstLegalTarget): each
// deals damage equal to its power to the other (CR 701.14), and a
// Taunter that has left the battlefield fights nothing. A Taunter
// damaged by the fight reflects that damage too, since it triggers on
// any damage.
//
// One declared simplification, weaker than printed and the same one
// Screaming Nemesis carries: the engine emits one damage event per
// SOURCE, so a Taunter blocked by two creatures fires twice — once per
// blocker's damage, each with its own opponent target — where the
// printed card fires once for the total. The same damage is reflected
// either way; only the split differs.
func init() {
	Register(Spec{
		OracleID:        "3e41648f-c5a9-4b26-b97e-8176fe5e9c85",
		Name:            "Brash Taunter",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"If two or more sources damage it at the same time, it reflects each source's damage separately instead of the total in one go."},
		PrintedKeywords: []string{"indestructible"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b35SelfWasDealtDamage(ev, source)
			},
			Targets: TargetPlayer("target opponent", Opponent()),
			Key:     "Brash Taunter — it deals that much damage to target opponent",
			Effect:  damagedCreatureReflectsChosenAmount,
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, {T}: This creature fights another target creature.",
			Cost:    Plus(ManaCost("{2}{R}"), TapCost()),
			Targets: Another(TargetCreature("another target creature")),
			Effect:  b30SourceFightsFirstLegalTarget,
		}},
	})
}
