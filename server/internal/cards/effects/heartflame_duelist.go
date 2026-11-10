package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heartflame Duelist // Heartflame Slash — Creature — Human Knight {1}{W},
// 3/1 // Instant — Adventure {2}{R} (Wilds of Eldraine, #4471):
//
//	Heartflame Duelist — "Instant and sorcery spells you control have
//	 lifelink."
//	Heartflame Slash — "Heartflame Slash deals 3 damage to any target."
//
// The static is Pestilent Spirit's spell-keyword grant with lifelink
// (Lo and Li's reader): the damage an instant or sorcery you control deals
// gains life for you. The Adventure's own damage is lifelinked only while
// the Duelist is on the battlefield, since the card is a spell on the
// stack then (not a creature) and the grant reads the battlefield.
//
// No simplification.
const heartflameDuelistOracleID = "3ca07da2-29a1-40ce-a7f3-4838e95ca1d9"

func init() {
	Register(Spec{
		OracleID:     heartflameDuelistOracleID,
		Name:         "Heartflame Duelist",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(SpellsYouControlHave(Or(Instant(), Sorcery()), "lifelink"),
				"Instant and sorcery spells you control have lifelink."),
		},
	})
	Register(Spec{
		OracleID:     heartflameDuelistOracleID + "#1",
		Name:         "Heartflame Slash",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(3),
	})
}
