package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain Benjamin Sisko — Legendary Creature — Human Officer {1}{W}, 2/2
// (Star Trek, #4973):
//
//	"At the beginning of your end step, create a 1/1 red Officer
//	 creature token for each dilemma you've faced this turn.
//	 {2}{W}{U}{B}{R}{G}, {T}: Exile the top five cards of your library.
//	 You may cast any number of spells with mana value less than or equal
//	 to the number of Officers you control from among those cards without
//	 paying their mana costs."
//
// "Each dilemma you've faced this turn" is the per-turn count the engine
// keeps beside EventFacedDilemma (Game.DilemmasFacedThisTurn, cleared as
// each turn begins), read as the trigger resolves (CR 608.2h): a dilemma
// faced in response still counts, and none makes no token. Dilemmas
// faced before Sisko entered count too; the sentence is about you, not
// him. The token is the printed 1/1 red Officer.
//
// The activated ability is Hazoret's Undying Fury's exile
// (ExileTopWithPermission.EachManaValueAtMost, #4021): the top five are
// exiled with one pass-closed permission over the cards that landed,
// each spell free and capped on its own at the number of Officers you
// control as the ability resolves, counted through the layered view (a
// changeling is an Officer, and Sisko is one himself). The casts are made
// during the resolution (CR 608.2g); an {X} is 0 (CR 107.3b). A card not
// cast, a land and a spell over the cap stay in exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f997ad8a-6903-47b1-83af-837b3b55a3d4",
		Name:         "Captain Benjamin Sisko",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Captain Benjamin Sisko — create a 1/1 red Officer token for each dilemma you've faced this turn",
				captainBenjaminSiskoOfficers),
		},
		Activated: []ActivatedAbility{{
			Label:  "{2}{W}{U}{B}{R}{G}, {T}: Exile the top five cards of your library. You may cast any number of spells with mana value less than or equal to the number of Officers you control from among those cards without paying their mana costs.",
			Cost:   Plus(ManaCost("{2}{W}{U}{B}{R}{G}"), TapCost()),
			Effect: captainBenjaminSiskoExile,
		}},
	})
}

// captainBenjaminSiskoOfficers is the end-step trigger: one Officer per
// dilemma faced this turn, counted now.
func captainBenjaminSiskoOfficers(g *game.Game, item *game.StackItem) error {
	n := g.DilemmasFacedThisTurnFor(item.Controller)
	if n <= 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("1/1 red Officer"), N: n}.Apply(NewContext(g, item))
}

// captainBenjaminSiskoExile is the activated ability's resolution.
func captainBenjaminSiskoExile(g *game.Game, item *game.StackItem) error {
	limit := trekPermanentsWithSubtypeYouControl(g, item, item.Controller, "Officer")
	return ExileTopWithPermission{
		From:                item.Controller,
		GrantTo:             item.Controller,
		N:                   5,
		EachManaValueAtMost: &limit,
	}.Apply(NewContext(g, item))
}
