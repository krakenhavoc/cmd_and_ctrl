package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// daynight_a_cards_test.go — the first half of the day/night card list
// (#2586, slice daynight-a), on real imported cards for the werewolves
// (so the importer's per-face daybound / nightbound stamping is on the
// road under test) and on pushed catalogue permanents for the rest.

const (
	DnDay   = "Daybound (If a player casts no spells during their own turn, it becomes night next turn.)"
	DnNight = "Nightbound (If a player casts at least two spells during their own turn, it becomes day next turn.)"

	DnaArlinn       = "f227ce07-7e96-4a36-ab7c-9be6e777d649"
	DnaAvabruck     = "730be0d6-2612-44d6-9d36-e1fc6510c6bd"
	DnaBallista     = "b6811c31-fcd3-4d00-89d5-cf974575a87c"
	DnaBaneblade    = "4f5da665-e880-4325-a9e2-6fc4ca1d807a"
	DnaBrutalCathar = "1ed2d8e0-462b-468e-8fd3-1f3c6d99fb8a"
	DnaBurlyBreaker = "4cbfb898-e1ca-4e34-9cb2-11ba50272984"
	DnaSanctifier   = "6885aa03-fbfe-4c2e-a785-0865788574b4"
	DnaChildOfPack  = "a9c5b155-7c09-4100-9679-8fca2b5e9222"
	DnaCollector    = "6950ab0f-f448-4ed1-9dd4-868cc79a5982"
	DnaFangblade    = "dbd22a65-4ccb-4435-ae27-03a47a86d630"
	DnaDawnguard    = "2788cc1e-18b6-4506-be4c-f4056600e33c"
	DnaTrespasser   = "0bbd6cad-9b6f-45a6-9f2e-d7b4853586ae"
	DnaHookhand     = "52def237-0374-4ab0-8c01-d0d00aaa324f"
	DnaHoundTamer   = "e9208fc2-616d-4c32-bd66-76a8bf85a6b5"
	DnaPiper        = "cede233b-4e27-4738-8099-c8e46862ba96"
	DnaLoner        = "6e0b3317-394d-42cc-a350-cb5ce051787a"
	DnaKessig       = "54f8acb1-58f3-49d1-bff8-c1b578245936"
	DnaLambholt     = "f3c104e2-470b-429f-a047-a21edc2adb3b"
)

// dnaWerewolf is one transforming werewolf: both faces' names, type
// lines, printed text and P/T, as the dump has them, and the keywords
// (beyond daybound and nightbound) each face is expected to carry.
type DnaWerewolf struct {
	Oracle                  string
	Front, Back             string
	FrontType, BackType     string
	Cost                    string
	FrontText, BackText     string
	Fp, Ft, Bp, Bt          string
	Colors                  []string
	ExtraKeywords           []string // the dump's top-level keywords
	FrontKeywords, BackKeys []string // expected on each face, beyond day/night
}

func (w DnaWerewolf) Row() cards.Card {
	return WerewolfRow(w.Oracle, w.Front, w.FrontType, w.Cost, w.FrontText, w.Back, w.BackType, w.BackText,
		w.Fp, w.Ft, w.Bp, w.Bt, w.Colors, w.ExtraKeywords...)
}

// dnaPower is current power, counters included (effectivePower is the
// layer result alone).
func DnaPower(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return c.CurrentPower()
}
