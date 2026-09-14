package effects

// Hour of Promise — Sorcery {4}{G} (EDHREC rank 3228):
//
//	"Search your library for up to two land cards, put them onto the
//	 battlefield tapped, then shuffle. Then if you control three or
//	 more Deserts, create two 2/2 black Zombie creature tokens."
//
// Explosive Vegetation for ANY two lands, with a Desert rider. The
// search is the shared library-search shape — lands of any kind,
// onto the battlefield tapped through the search's own entry flag,
// then shuffle — and the Desert count runs in the search's
// continuation, after the fetched lands have entered, so two fetched
// Deserts can make the third. The Zombies are the shared 2/2 black
// template.
//
// The catalog-wide search posture applies: with two or fewer land
// cards in the library there is nothing to choose and both are
// taken; a fuller library opens the search picker, where the
// searcher may take fewer than two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "51ff3e53-90bb-41bf-80e4-bb3fd51d574a",
		Name:         "Hour of Promise",
		Completeness: CompletenessFull,
		OnResolve:    b30SearchUpToTwoLandsTappedThenZombies,
	})
}
