package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// J. Jonah Jameson — Legendary Creature — Human Citizen {2}{R}:
//
//	"When J. Jonah Jameson enters, suspect up to one target creature.
//	 (A suspected creature has menace and can't block.)
//	 Whenever a creature you control with menace attacks, create a
//	 Treasure token."
//
// The second ability reads the attacker's EFFECTIVE keywords as it is
// declared, so a printed menace, a suspected creature's menace (CR
// 701.60c) and a menace granted until end of turn all count, and it
// fires once per attacking creature, as printed ("a creature", not
// "one or more").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de861715-fd0b-493e-9a7c-c470a23044c0",
		Name:         "J. Jonah Jameson",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("J. Jonah Jameson — suspect up to one target creature", SuspectEachLegalTarget),
				UpToOneTargetCreature("up to one target creature"),
			),
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !attackDeclaredByYou(ev, source.Controller) {
					return false
				}
				attacker, ok := g.LookupCardForEffect(ev.CardID)
				return ok && game.HasKeyword(&attacker, "menace")
			}, "J. Jonah Jameson — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
