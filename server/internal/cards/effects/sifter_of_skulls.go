package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sifter of Skulls — Creature — Eldrazi {3}{B}, 4/3:
//
//		"Devoid (This card has no color.)
//		 Whenever another nontoken creature you control dies, create a 1/1
//		 colorless Eldrazi Scion creature token. It has "Sacrifice this
//		 token: Add {C}.""
//
//	  - Devoid is the engine's colour-removing keyword (CR 702.114a).
//	  - The dies condition looks back in time (CR 603.10a): "you control"
//	    is the controller the exit stamped on the event
//	    (leftUnderControlOf), so a creature you stole and lost, or one an
//	    opponent stole from you, counts for whoever controlled it as it
//	    died. "Another" excludes the Sifter itself; "nontoken" excludes
//	    a token, which still dies but is not a card. Each qualifying
//	    creature is its own trigger, so a wipe that kills three of them
//	    makes three Scions.
//	  - The Scion is the catalog's own Eldrazi Scion (b28EldraziScionToken),
//	    a registered template whose "Sacrifice this token: Add {C}" is a
//	    real mana ability, not text on a vanilla token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "688e5e19-8f28-45c4-ace4-a38144f494f1",
		Name:            "Sifter of Skulls",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventLTB, sifterOfSkullsCondition,
				"Sifter of Skulls — create a 1/1 colorless Eldrazi Scion",
				Do(CreateToken{Template: b28EldraziScionToken(), N: 1})),
				game.Purpose{DeathPayoff: true, Tokens: 1}),
		},
	})
}

// sifterOfSkullsCondition is "another nontoken creature you control
// dies".
func sifterOfSkullsCondition(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
	if !AnotherCreatureDied(ev, source, ch, g) {
		return false
	}
	dead, _ := diedCreature(ev, g)
	return !IsToken(dead) && leftUnderControlOf(ev, dead) == source.Controller
}
