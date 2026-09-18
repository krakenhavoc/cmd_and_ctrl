package effects

// Gigantosaurus — Creature — Dinosaur {G}{G}{G}{G}{G}, 10/10 (EDHREC
// rank 4468):
//
//	(no rules text)
//
// The largest French-vanilla body in the batch, and one of the
// largest in the format for five mana. It is here because a vanilla
// card still has to be REGISTERED to behave: a card the catalog does
// not know falls back to the manual sandbox path, and a Dinosaur
// deck's five-drop should resolve, route to the battlefield and
// attack without anyone clicking a manual trigger.
//
// The five green pips are the whole card and they are the deck
// importer's business, not this file's — the Spec carries no mana
// cost, because the printed cost rides game.Card.ManaCost from
// Scryfall. Nothing here but the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e666bae7-dd51-4921-8b89-7e8d423caba0",
		Name:         "Gigantosaurus",
		Completeness: CompletenessFull,
	})
}
