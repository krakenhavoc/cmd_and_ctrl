package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gtc_5565_helpers.go — shared bodies for Gatecrash, batch 1 (#5565).

// gtcBloodrush is Gatecrash's bloodrush (Pyrewild Shaman's shape): an
// activated ability that works only from the hand, whose cost is the mana
// and discarding this card, and whose effect is "Target attacking creature
// gets +P/+T [and gains <keyword>] until end of turn." Both halves end at
// cleanup (CR 514.2). A target that stopped attacking or left is illegal,
// and the ability does nothing (CR 608.2b).
func GtcBloodrush(name, mana string, power, toughness int, keyword string) ActivatedAbility {
	effect := fmt.Sprintf("Target attacking creature gets %+d/%+d until end of turn.", power, toughness)
	resolve := BoostTheTargetUntilEOT(power, toughness, name+" — bloodrush")
	var keywords []string
	if keyword != "" {
		effect = fmt.Sprintf("Target attacking creature gets %+d/%+d and gains %s until end of turn.", power, toughness, keyword)
		resolve = SocB27TargetGetsAndGains(power, toughness, keyword, name+" — bloodrush")
		keywords = []string{keyword}
	}
	return ActivatedAbility{
		Label:   "Bloodrush — " + mana + ", Discard this card: " + effect,
		Cost:    game.AbilityCost{Mana: mana, DiscardSelf: true},
		Zones:   []game.ZoneKind{game.ZoneHand},
		Purpose: ForTargets(BenefitUntilEOT(0, power, toughness, keywords...)),
		Targets: TargetCreature("target attacking creature", AttackingCreature()),
		Effect: func(g *game.Game, item *game.StackItem) error {
			return resolve(item, NewContext(g, item))
		},
	}
}

// gtcBattalion is "Battalion — Whenever this creature and at least two
// other creatures attack, <effect>" (Boros Strike-Captain's group test,
// battalionAttack).
func GtcBattalion(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventAttack, BattalionAttack, label, effect)
}

// gtcKeyruneAnimate is a Gatecrash Keyrune's second ability: "This
// artifact becomes a P/T <colours> <subtype> artifact creature with
// <keyword> until end of turn." It is Levitating Statue's animation
// (cmdr2B13Animate): one effect pinned to this object (CR 400.7) adding
// the creature type and subtype (layer 4), setting the colours (layer 5),
// granting the keyword (layer 6) and setting the base power and toughness
// (layer 7b). It stays an artifact and keeps its mana ability, which the
// engine refuses while it is a summoning-sick creature (CR 302.6).
func GtcKeyruneAnimate(label string, power, toughness int, subtype string, colours []string, keyword string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		return Cmdr2B13Animate(NewContext(g, item), item.SourceCardID, label, power, toughness, subtype, colours, keyword)
	}
}
