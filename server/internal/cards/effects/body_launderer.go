package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Body Launderer — Creature — Ogre Rogue {2}{B}{B}, 3/3 (Streets of New
// Capenna):
//
//	"Deathtouch
//	 Whenever another nontoken creature you control dies, this creature
//	 connives. (Draw a card, then discard a card. If you discarded a
//	 nonland card, put a +1/+1 counter on this creature.)
//	 When this creature dies, return another target non-Rogue creature
//	 card with equal or lesser power from your graveyard to the
//	 battlefield."
//
// Two printed abilities, two registered triggers. The connive trigger
// reads the died creature's last-known information for "nontoken" and "you
// control" (CR 603.10a), and never fires for the Launderer itself. The dies
// trigger's clause compares the candidate's power with the Launderer's
// last-known power (RelativeToSource, CR 608.2h) at announce and again on
// resolution; the Launderer is itself a Rogue, so it can never pick itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eaca2576-020b-4af4-b348-6d1b56116c41",
		Name:            "Body Launderer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b19AnotherNontokenCreatureYouControlDied(ev, source, g)
				},
				Key:    "Body Launderer — it connives",
				Effect: Do(Instruction(sourceConnives)),
			},
			Targeting(WhenThisDies("Body Launderer — return another target non-Rogue creature card with equal or lesser power from your graveyard to the battlefield",
				returnFirstLegalGraveyardTargetToBattlefield),
				RelativeToSource(TargetCardInGraveyard("another target non-Rogue creature card with equal or lesser power from your graveyard",
					Creature(), YouOwn(), Not(OfCreatureType("Rogue"))), PowerNoGreater())),
		},
	})
}
