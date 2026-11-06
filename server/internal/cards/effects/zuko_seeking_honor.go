package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zuko, Seeking Honor — Legendary Creature — Human Noble {2}{B/R}, 2/2
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"Firebending 1 (Whenever this creature attacks, add {R}. This mana
//	 lasts until end of combat.)
//	 Whenever you cast a noncreature spell, Zuko gains first strike until
//	 end of turn.
//	 Whenever Zuko deals combat damage to a player, put a +1/+1 counter
//	 on him."
//
// Firebending is the set's shared attack trigger (firebending, Boiling
// Rock Rioter's row): it uses the stack and isn't a mana ability. The
// cast trigger is Dwarven Strike Force's thisCreatureUntilEOT with no
// pump, so a Zuko that has left gains nothing. The combat-damage trigger
// is a CounterOnThis.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6e737c52-cbe0-493f-ae46-7bf068e04882",
		Name:         "Zuko, Seeking Honor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Zuko, Seeking Honor — firebending 1", Do(firebending(1))),
			WheneverYouCast(Noncreature(), "Zuko, Seeking Honor — first strike until end of turn",
				thisCreatureUntilEOT("Zuko, Seeking Honor — first strike until end of turn", 0, 0, "first strike")),
			WheneverThisDealsCombatDamageToAPlayer("Zuko, Seeking Honor — put a +1/+1 counter on him",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
