package effects

import "testing"

// dredge_signal_test.go — #2390. The heuristic bot prices a dredge offer
// from the prompt's `dredge` count and its Source (the card the yes
// returns). Two claims about the catalog hold that up, and a new card
// that broke either would be priced wrongly without anything failing:
//
//   - every dredge — every graveyard "may" on a draw — declares its N,
//     so the prompt can say what the yes mills; and
//   - a dredge that names no card (its effect is not in the graveyard,
//     so the prompt has no Source) returns a LAND, because that is what
//     aiseat/heuristic/dredge.go prices it as. The Necrobloom's "land
//     cards in your graveyard have dredge 2" is the only one. A new
//     grant over another class of cards needs the wire to say which.
func TestOnlyTheNecrobloomGrantsDredgeWithoutACard(t *testing.T) {
	for _, s := range All() {
		for i, r := range s.Replacements {
			if r.FromGraveyard && r.Optional && r.Dredge <= 0 {
				t.Errorf("%s replacement %d is a graveyard \"may\" with no Dredge count; the bot cannot price its mill", s.Name, i)
			}
			if r.Dredge > 0 && !r.FromGraveyard && s.Name != "The Necrobloom" {
				t.Errorf("%s replacement %d grants dredge without naming the card; the bot prices that as returning a land (aiseat/heuristic/dredge.go)", s.Name, i)
			}
		}
	}
}
