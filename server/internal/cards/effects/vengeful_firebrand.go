package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vengeful Firebrand — Creature — Elemental Warrior {3}{R}, 5/2
// (Morningtide, #5592):
//
//	"This creature has haste as long as a Warrior card is in your
//	 graveyard.
//	 {R}: This creature gets +1/+0 until end of turn."
//
// The haste is Aang, A Lot to Learn's graveyard-gated layer 6 grant
// (CR 613.1f), re-read on every recompute, so it comes and goes with the
// graveyard; "a Warrior card" reads the card's subtypes, so a changeling
// card counts. The pump is the shared thisGetsRow.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "920e21c4-0116-416e-a375-bb887cb7b44a",
		Name:         "Vengeful Firebrand",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfOnly(target, g, source) && morWarriorCardInGraveyard(g, source)
			}, "haste"), "This creature has haste as long as a Warrior card is in your graveyard."),
		},
		Activated: []ActivatedAbility{thisGetsRow("{R}", 1, 0, "Vengeful Firebrand")},
	})
}

// morWarriorCardInGraveyard is "a Warrior card is in your graveyard".
func morWarriorCardInGraveyard(g *game.Game, source *game.Card) bool {
	p := g.PlayerByIDForEffect(source.Controller)
	if p == nil || p.Graveyard == nil {
		return false
	}
	for _, c := range p.Graveyard.Cards {
		if c.HasSubtype("Warrior") {
			return true
		}
	}
	return false
}
