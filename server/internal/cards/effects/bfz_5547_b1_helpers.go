package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// bfz_5547_b1_helpers.go — shared pieces of Battle for Zendikar batch 1
// (S82, #5547): rally, the Ally ability word that repeats across the set
// with a different team effect each time.

// bfzB1RallyWhen is rally's condition, "Whenever this creature or
// another Ally you control enters" (Appa, the Vigilant's and South Pole
// Voyager's condition): the source itself entering, or another permanent
// with the Ally subtype entering under its controller's control. "Ally"
// is a word of the card's text (CR 612.2), read post-layer, so a
// changeling counts.
func bfzB1RallyWhen(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.CardID == source.InstanceID ||
		b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Ally"))
}

// bfzB1Rally is a rally trigger: one trigger per entering Ally, each on
// the stack, so two Allies entering at once give two triggers.
func bfzB1Rally(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventETB, bfzB1RallyWhen, label, effect)
}

// bfzB1RallyTeamGains is "Rally — …, creatures you control gain
// <keyword> until end of turn." The creatures are fixed as the trigger
// resolves (CR 611.2c), so a creature that enters later in the turn does
// not gain the keyword from this instance; the grant ends in the cleanup
// step (CR 514.2).
func bfzB1RallyTeamGains(name, keyword string) game.TriggeredAbility {
	label := name + " — creatures you control gain " + keyword + " until end of turn"
	return bfzB1Rally(label, Do(GrantKeywordUntilEOT{
		Match:    And(Creature(), YouControl()),
		Keywords: []string{keyword},
		Label:    label,
	}))
}

// bfzB1RallyTeamGets is "Rally — …, creatures you control get +N/+N
// until end of turn." (Tajuru Beastmaster, Tajuru Warcaller), the set
// locked as it resolves (CR 611.2c).
func bfzB1RallyTeamGets(name string, n int) game.TriggeredAbility {
	label := name + " — creatures you control get " + signedPT(n, n) + " until end of turn"
	return bfzB1Rally(label, Do(BoostUntilEOT{
		Match:     And(Creature(), YouControl()),
		Power:     n,
		Toughness: n,
		Label:     label,
	}))
}
