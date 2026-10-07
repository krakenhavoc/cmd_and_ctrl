package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestEveryColourExceptionDropsDevoidFromATokenCopy — #2322 / CR 707.9d.
// Each colour-setting copy exception, on both keyword roads, yields a
// token whose effective abilities carry no devoid and keep the rest; a
// plain token copy of the same card keeps it.
func TestEveryColourExceptionDropsDevoidFromATokenCopy(t *testing.T) {
	excepts := map[string]func(*game.Card){
		"scarab god": scarabGodZombieException,
		"embalm":     embalmException,
		"eternalize": eternalizeException,
		"hashaton":   hashatonZombieException,
		"sauron":     sauronWraithException,
		"plain":      nil,
	}
	roads := map[string]game.Card{
		"catalog": {Name: "Propagator Drone", TypeLine: "Creature — Eldrazi Drone",
			OracleID: oraclePropagatorDrone, ManaCost: "{1}{G}", Colors: []string{}, Power: 2, Toughness: 2},
		"import": {Name: "Fixture Drone", TypeLine: "Creature — Eldrazi Drone",
			ManaCost: "{2}{U}", Colors: []string{}, Keywords: []string{game.KeywordDevoid, "haste"}, Power: 2, Toughness: 2},
	}
	for en, except := range excepts {
		for rn, card := range roads {
			t.Run(en+"/"+rn, func(t *testing.T) {
				g := newCatalogGame(t)
				me := g.Seats[0]
				src := card
				src.InstanceID, src.Owner, src.Controller = uuid.New(), me.ID, me.ID
				me.Graveyard.PushTop(src)
				g.WithWriteLock(func() {
					if err := (CreateTokenCopy{Controller: me.ID, Copy: src.InstanceID, N: 1, Except: except}).Apply(NewContext(g, nil)); err != nil {
						t.Fatalf("CreateTokenCopy.Apply: %v", err)
					}
				})
				token := findBattlefieldByName(g, src.Name)
				if token == uuid.Nil {
					t.Fatal("no token copy was created")
				}
				got := battlefieldCardCopy(t, g, token)
				has := slices.Contains(got.Effective().Abilities, game.KeywordDevoid)
				if want := except == nil; has != want {
					t.Errorf("devoid in abilities = %v, want %v: %v", has, want, got.Effective().Abilities)
				}
				if except == nil {
					return
				}
				// The import road: Card.Keywords is what the
				// off-battlefield keyword read walks first.
				if slices.Contains(got.Keywords, game.KeywordDevoid) {
					t.Errorf("Card.Keywords still lists devoid: %v", got.Keywords)
				}
				if len(got.EffectiveColors()) != 1 {
					t.Errorf("colours = %v, want the exception's one colour", got.EffectiveColors())
				}
				if rn == "import" && !slices.Contains(got.Effective().Abilities, "haste") {
					t.Errorf("other keywords must survive: %v", got.Effective().Abilities)
				}
			})
		}
	}
}
