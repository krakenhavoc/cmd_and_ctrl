package effects

// Zhur-Taa Druid — Creature — Human Druid {R}{G}, 1/1 (EDHREC rank
// 4194):
//
//	"{T}: Add {G}.
//	 Whenever you tap this creature for mana, it deals 1 damage to
//	 each opponent."
//
// A two-mana dork that pings the whole table every time it ramps you.
// At a four-player game that is three damage a turn for free, which
// is why it shows up in Gruul and in group-slug decks rather than in
// ramp decks proper.
//
// The damage is a TRIGGERED ability, not part of the mana ability, and
// it uses the stack (#4928). CR 605.1b makes a triggered ability a mana
// ability only if "it could add mana to a player's mana pool when it
// resolves"; one that deals damage cannot, so it is an ordinary trigger.
// The {G} itself lands at once, because the activated mana ability
// "resolves immediately after it is activated" (CR 605.3b); the damage
// goes on the stack the next time a player would receive priority
// (CR 603.3), where the table can respond to it or counter it.
//
// It is declared with PingOpponentsTrigger, which watches the ACTIVATION
// of this row rather than the tap event: "whenever you tap this creature
// FOR MANA" fires only for a mana activation. A tap-event trigger — City
// of Brass's shape — would also fire when an opponent's Icy Manipulator
// taps the Druid, when it crews a Vehicle and when it attacks, which is
// a strictly better card and the direction this catalog does not ship
// (#259). The declaration also tells the auto-tapper the source acts on
// the table, so it plans the Druid and the move list counts its {G}
// (#4189), reaching for it only after the sources that do not.
//
// The trigger fires for this printed row. A {T} mana ability some other
// effect grants the Druid does not ping, which is weaker than printed.
func init() {
	Register(Spec{
		OracleID:     "5979310e-38c8-489f-ab9e-2723af98a3a5",
		Name:         "Zhur-Taa Druid",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:                 ManaAbilityCost{Tap: true},
			Produced:             "{G}",
			Label:                "Add {G}",
			PingOpponentsTrigger: 1,
		}},
	})
}
