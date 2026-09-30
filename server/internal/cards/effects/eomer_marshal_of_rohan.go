package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Éomer, Marshal of Rohan — Legendary Creature — Human Knight
// {2}{R}{R}, 4/4:
//
//	"Haste
//	 Whenever one or more other attacking legendary creatures you
//	 control die, untap all creatures you control. After this phase,
//	 there is an additional combat phase. This ability triggers only
//	 once each turn."
//
// "Attacking", "legendary" and "you control" are the dying creature as
// it last existed on the battlefield (CR 603.10a): the combat state,
// supertypes and controller the exit stamped on the event, so a stolen
// legend that dies attacking for you counts and a Clone of a legend
// counts. "One or more" is OncePerBatch; "only once each turn" is the
// per-object trigger tally. An attacking creature only dies during
// combat, so "after this phase" is always the combat in progress, and
// the additional combat follows it with no main phase between (ADR
// 0059 sub-PR 2b, #753; tracker #1564).
//
// No simplification.
func init() {
	const label = "Éomer, Marshal of Rohan — untap all creatures you control, additional combat"
	Register(Spec{
		OracleID:        "b2d95950-18b3-463f-94f4-299e420751dc",
		Name:            "Éomer, Marshal of Rohan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedWhileAttacking(ev, g)
				if !ok || dead.InstanceID == source.InstanceID {
					return false
				}
				return leftAsSupertype(ev, dead, "Legendary") &&
					leftUnderControlOf(ev, dead) == source.Controller &&
					!b11TriggeredThisTurn(g, source.InstanceID, label)
			}, label, untapAllYouControlThenExtraCombat)),
		},
	})
}
