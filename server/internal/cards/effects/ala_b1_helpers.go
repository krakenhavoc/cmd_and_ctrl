package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ala_b1_helpers.go — the shapes Shards of Alara, batch 1 (#5588) prints
// more than once. Append-only.

// alaB1LoneAttackerGains is "Whenever a creature you control attacks
// alone, it gains <keyword> until end of turn." (Rafiq of the Many's
// double strike, Battlegrace Angel's lifelink). "Alone" is the whole
// declaration (CR 506.5), read as the trigger is harvested, after every
// attacker is stamped; the creature that gains it is the one the event
// names, and one that left in response gains nothing (CR 400.7).
func alaB1LoneAttackerGains(name, keyword string) game.TriggeredAbility {
	return On(game.EventAttack, mshB1YouAttackAlone,
		name+" — the lone attacker gains "+keyword+" until end of turn",
		func(g *game.Game, item *game.StackItem) error {
			attacker := item.Trigger.Event.CardID
			if !onBattlefield(g, attacker) {
				return nil
			}
			return GrantKeywordUntilEOT{
				Target:   attacker,
				Keywords: []string{keyword},
				Label:    name + " — " + keyword,
			}.Apply(NewContext(g, item))
		})
}

// alaB1PowerFiveOrGreater is the Naya cards' "if you control a creature
// with power 5 or greater" (Drumhunter): current power, counters and
// anthems included, as the trigger triggers and again as it resolves
// (CR 603.4, through InterveningIf).
var alaB1PowerFiveOrGreater = CondPowerAtLeast{N: 5}

// alaB1AtYourEndStepIfPowerFive is "At the beginning of your end step, if
// you control a creature with power 5 or greater, <effect>." The trigger
// condition and the resolution re-check read the same board (CR 603.4).
func alaB1AtYourEndStepIfPowerFive(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(alaB1PowerFiveOrGreater, On(game.EventBeginEndStep,
		func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return ev.Actor == source.Controller && b24ControlsCreatureWithPowerAtLeast(g, source.Controller, 5)
		}, label, effect))
}
