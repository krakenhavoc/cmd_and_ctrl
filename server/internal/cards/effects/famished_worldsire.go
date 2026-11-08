package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Famished Worldsire — Creature — Leviathan {5}{G}{G}{G}, 0/0 (Edge of
// Eternities, #4461):
//
//	"Ward {3}
//	 Devour land 3 (As this creature enters, you may sacrifice any number
//	 of lands. It enters with three times that many +1/+1 counters on
//	 it.)
//	 When this creature enters, look at the top X cards of your library,
//	 where X is this creature's power. Put any number of land cards from
//	 among them onto the battlefield tapped, then shuffle."
//
// Devour is Mycoloth's as-enters sacrifice (CR 614.12) over lands: zero
// is always a legal answer, and the counters ride the entry, so Doubling
// Season doubles them. The entry trigger reads the Worldsire's power as
// it resolves, or its last-known power if it has left (CR 608.2h). The
// cards are looked at, so only you see them, and any number of land cards
// among them (none included) enter tapped as one batch; the library is
// then shuffled whatever was chosen. A Worldsire that devoured nothing
// is a 0/0 and dies to the state-based actions before its trigger
// resolves, which then looks at no cards and only shuffles.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e7f53928-a4cc-41ed-877d-b72d786e77d9",
		Name:         "Famished Worldsire",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{DevourLand("Famished Worldsire", 3)},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{3}"), "Famished Worldsire — ward {3}"),
			WhenThisEnters("Famished Worldsire — look at the top X cards; put any number of lands onto the battlefield tapped, then shuffle",
				famishedWorldsireLookAndPut),
		},
	})
}

// famishedWorldsireLookAndPut looks at the top X cards, X being the
// Worldsire's power, and offers the land cards among them.
func famishedWorldsireLookAndPut(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	var looked []uuid.UUID
	if x := sourcePowerNow(ctx); x > 0 {
		looked = g.LookAtTopOfLibraryForEffect(player, x)
	}
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    looked,
		Match:    Land(),
		Max:      0,
		Optional: true,
		Tapped:   true,
		Label:    "Famished Worldsire — put any number of land cards from among them onto the battlefield tapped",
		ThenTail: thenShuffle,
	}.Apply(ctx)
}
