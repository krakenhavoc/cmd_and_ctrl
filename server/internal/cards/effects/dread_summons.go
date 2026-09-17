package effects

// Dread Summons — Sorcery {X}{B}{B} (EDHREC rank 3708):
//
//	"Each player mills X cards. For each creature card put into a
//	 graveyard this way, you create a tapped 2/2 black Zombie
//	 creature token. (To mill a card, a player puts the top card of
//	 their library into their graveyard.)"
//
// The mill-the-table Zombie maker. X is read off the announced
// value; every live player — the caster first, then the opponents
// in seat order — mills X, or the rest of their library when it
// holds fewer (a mill never loses a player the game, CR 701.17b), and
// the milled batch is read back so the creature cards among it are
// counted at the moment they land. The Zombies are all created
// after all the mills, as the printed "for each" reads, tapped and
// under the CASTER's control — "you create" — whoever's graveyard
// the creature card went to. The Zombie is Overseer of the Damned's
// tapped 2/2 black Zombie.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "642fa01d-025e-44dd-8360-5325e5a28282",
		Name:         "Dread Summons",
		Completeness: CompletenessFull,
		OnResolve:    b35EachPlayerMillsXThenZombiesPerCreature,
	})
}
