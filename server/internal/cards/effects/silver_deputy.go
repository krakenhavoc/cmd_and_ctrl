package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silver Deputy — Artifact Creature — Mercenary {2}, 1/2 (Outlaws of
// Thunder Junction, #4468):
//
//	"When this creature enters, you may search your library for a basic
//	 land card or a Desert card, reveal it, then shuffle and put it on
//	 top.
//	 {T}: Target creature you control gets +1/+0 until end of turn.
//	 Activate only as a sorcery."
//
// The search is Outcaster Greenblade's predicate (a basic land card or a
// Desert card) under the one-mana tutors' "then shuffle and put it on
// top" (SearchLibrary.ToTop): the card never leaves the library, and is
// placed after the shuffle. "You may search" lets you decline the search
// and its shuffle. The tap ability is the set's Mercenary token's row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7bbede9a-8981-4435-9854-af61a18c6166",
		Name:         "Silver Deputy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Silver Deputy — you may search your library for a basic land card or a Desert card and put it on top",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: b34IsBasicLandOrDesertCard,
						Dest:      game.ZoneLibrary,
						ToTop:     true,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Silver Deputy — a basic land card or a Desert card, to the top of your library",
					}.Apply(NewContext(g, item))
				}),
		},
		// No Purpose.Answers: the row is sorcery speed and always names a
		// target (ADR 0142).
		Activated: []ActivatedAbility{{
			Label:        "{T}: Target creature you control gets +1/+0 until end of turn. Activate only as a sorcery.",
			Cost:         TapCost(),
			Targets:      TargetCreature("target creature you control", YouControl()),
			Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0)),
			SorcerySpeed: true,
			Effect:       boostTargetUntilEOT("Silver Deputy — +1/+0", 1, 0),
		}},
	})
}
