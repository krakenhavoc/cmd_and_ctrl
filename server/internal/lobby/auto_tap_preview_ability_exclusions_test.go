package lobby

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_tap_preview_ability_exclusions_test.go — #1422. The ?ability=
// branch of the auto-tap preview excludes what the activation itself
// excludes from its auto-tap (game.ActivationAutoTapExclusions): the
// source when the cost taps or sacrifices it, the permanents named to
// the other cost components, and the waterbend taps — and it subtracts
// the waterbend taps from the price, as the activation does. Before,
// it excluded only ?exclude= and could plan the ability's own {T}
// source for its mana.
//
// Each test holds the preview to the activation on the same board
// (Strict + AutoTap): "ok" means the engine takes the activation and
// taps exactly the plan for mana; "not ok" means it refuses.

const (
	castleVantressOracle   = "cdf41cf4-4e77-453d-be5b-0abbbd358934"
	sceneOfTheCrimeOracle  = "ba11a517-1dbd-4797-9f5e-46ce0f6c77c0"
	kataraWaterTribeOracle = "234fb291-0b62-4092-9071-81311c71bd53"
	mindStoneOracle        = "c97361b5-af16-4a7b-af85-a429dbaf4ad2"
)

// activateWith runs ability 0 of `card` for Alice with `params`, the
// payment enforced and auto-tapped.
func (f *previewFixture) activateWith(card uuid.UUID, params game.ActivateAbilityParams) error {
	f.t.Helper()
	params.Strict = true
	params.AutoTap = true
	return f.game().ActivateCatalogAbility(f.alice, card, 0, params)
}

func containsID(ids []string, id uuid.UUID) bool {
	for _, s := range ids {
		if s == id.String() {
			return true
		}
	}
	return false
}

// Castle Vantress — "{T}: Add {U}. {2}{U}{U}, {T}: Scry 2." The {T} in
// the ability's cost is the Castle's, so its own {U} cannot pay the
// {2}{U}{U}. Three Islands plus the Castle: the old preview planned the
// Castle as the fourth source and said ok; the activation refuses.
// With a fourth Island both agree, and the plan is the four Islands.
func TestAutoTapPreviewAbilityDoesNotPlanItsOwnTapSource(t *testing.T) {
	f := newPreviewFixture(t)
	islands := f.lands("Island", 3)
	castle := f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Castle Vantress", TypeLine: "Land", OracleID: castleVantressOracle,
	}, 1)[0]
	if c, _ := f.game().LookupCardForEffect(castle); c.Tapped {
		t.Fatalf("Castle Vantress entered tapped beside three Islands")
	}

	short := f.previewAbility(castle, "")
	if short.OK {
		t.Fatalf("three Islands + the Castle for {2}{U}{U}, {T}: preview ok with plan %v — the Castle's own {T} is already spent", short.Plan)
	}
	if err := f.activateWith(castle, game.ActivateAbilityParams{}); err == nil {
		t.Fatalf("the preview said no and the activation went through")
	}

	islands = append(islands, f.lands("Island", 1)...)
	got := f.previewAbility(castle, "")
	if !got.OK || len(got.Plan) != 4 || containsID(got.Plan, castle) {
		t.Fatalf("four Islands + the Castle: ok=%v plan=%v missing=%v, want ok with the four Islands", got.OK, got.Plan, got.Missing)
	}
	if err := f.activateWith(castle, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the preview said ok and the activation was refused: %v", err)
	}
	if tapped, want := f.tappedAmong(islands), sortedCopy(got.Plan); !equalStrings(tapped, want) {
		t.Errorf("the activation tapped %v for mana, the preview planned %v", tapped, want)
	}
}

// Scene of the Crime — "{T}: Add {C}. {2}, Sacrifice this land: Draw a
// card." No {T} in the crack, so the land is untapped and a mana source
// — but it is the sacrifice, so it cannot also make mana for the {2}.
// One Mountain beside it is a miss for both; two is a hit for both,
// planned off the Mountains.
func TestAutoTapPreviewAbilityDoesNotPlanItsSacrificedSource(t *testing.T) {
	f := newPreviewFixture(t)
	mountains := f.lands("Mountain", 1)
	scene := f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Scene of the Crime", TypeLine: "Artifact Land — Clue", OracleID: sceneOfTheCrimeOracle,
	}, 1)[0]
	// "This land enters tapped" — untap it, so it IS a mana source the
	// plan could reach for.
	if err := f.game().TapCard(scene, false); err != nil {
		t.Fatalf("untap Scene of the Crime: %v", err)
	}

	short := f.previewAbility(scene, "")
	if short.OK {
		t.Fatalf("one Mountain + Scene of the Crime for {2}, sacrifice it: preview ok with plan %v — the land is the sacrifice", short.Plan)
	}
	if err := f.activateWith(scene, game.ActivateAbilityParams{}); err == nil {
		t.Fatalf("the preview said no and the activation went through")
	}

	mountains = append(mountains, f.lands("Mountain", 1)...)
	got := f.previewAbility(scene, "")
	if !got.OK || len(got.Plan) != 2 || containsID(got.Plan, scene) {
		t.Fatalf("two Mountains + the Scene: ok=%v plan=%v missing=%v, want ok with the two Mountains", got.OK, got.Plan, got.Missing)
	}
	if err := f.activateWith(scene, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the preview said ok and the activation was refused: %v", err)
	}
	if tapped, want := f.tappedAmong(mountains), sortedCopy(got.Plan); !equalStrings(tapped, want) {
		t.Errorf("the activation tapped %v for mana, the preview planned %v", tapped, want)
	}
}

// Katara, Water Tribe's Hope — "Waterbend {X}: …". X=2 with a Mind
// Stone named to the waterbend and one Forest: the Mind Stone's tap pays
// {1}, the Forest pays the other. The Mind Stone is spent on the
// waterbend, so the plan must not ALSO tap it for its {C} — the old
// preview priced the full {2} and planned both.
func TestAutoTapPreviewAbilityHonoursWaterbendTaps(t *testing.T) {
	f := newPreviewFixture(t)
	f.mainPhaseForAlice() // "Activate only during your turn."
	katara := f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Katara, Water Tribe's Hope", TypeLine: "Legendary Creature — Human Warrior Ally",
		OracleID: kataraWaterTribeOracle, Power: 3, Toughness: 3,
	}, 1)[0]
	stone := f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Mind Stone", TypeLine: "Artifact", OracleID: mindStoneOracle,
	}, 1)[0]
	forests := f.lands("Forest", 1)
	query := "&x=2&waterbend_ids=" + stone.String()

	got := f.previewAbility(katara, query)
	if !got.OK || len(got.Plan) != 1 || containsID(got.Plan, stone) {
		t.Fatalf("Waterbend {2}, Mind Stone waterbent, one Forest: ok=%v plan=%v missing=%v, want ok with the Forest alone", got.OK, got.Plan, got.Missing)
	}
	params := game.ActivateAbilityParams{XValue: 2, WaterbendIDs: []uuid.UUID{stone}}
	if err := f.activateWith(katara, params); err != nil {
		t.Fatalf("the preview said ok and the activation was refused: %v", err)
	}
	if tapped, want := f.tappedAmong(forests), sortedCopy(got.Plan); !equalStrings(tapped, want) {
		t.Errorf("the activation tapped %v for mana, the preview planned %v", tapped, want)
	}
}

// And the miss: X=3 with the same board owes {2} after the one
// waterbend tap, and the Forest is the only source the plan may use.
// The preview says no, and so does the activation.
func TestAutoTapPreviewAbilityWaterbendMiss(t *testing.T) {
	f := newPreviewFixture(t)
	f.mainPhaseForAlice()
	katara := f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Katara, Water Tribe's Hope", TypeLine: "Legendary Creature — Human Warrior Ally",
		OracleID: kataraWaterTribeOracle, Power: 3, Toughness: 3,
	}, 1)[0]
	stone := f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Mind Stone", TypeLine: "Artifact", OracleID: mindStoneOracle,
	}, 1)[0]
	f.lands("Forest", 1)

	if got := f.previewAbility(katara, "&x=3&waterbend_ids="+stone.String()); got.OK {
		t.Fatalf("Waterbend {3}, Mind Stone waterbent, one Forest: preview ok with plan %v, want a miss ({2} owed, one source)", got.Plan)
	}
	params := game.ActivateAbilityParams{XValue: 3, WaterbendIDs: []uuid.UUID{stone}}
	if err := f.activateWith(katara, params); err == nil {
		t.Fatalf("the preview said no and the activation went through")
	}
}

// The cast branch is untouched: a {2} spell off a Mind Stone and a
// Forest plans both, and CastSpell taps exactly those. The
// ability-only params (exile_ids, waterbend_ids) are not read there —
// naming the Mind Stone in them changes nothing.
func TestAutoTapPreviewCastBranchUnchangedByAbilityParams(t *testing.T) {
	f := newPreviewFixture(t)
	f.mainPhaseForAlice()
	spell := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Two Drop", TypeLine: "Creature — Construct", ManaCost: "{2}", Power: 2, Toughness: 2,
	}, 1)[0]
	stone := f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Mind Stone", TypeLine: "Artifact", OracleID: mindStoneOracle,
	}, 1)[0]
	forest := f.lands("Forest", 1)[0]
	sources := []uuid.UUID{stone, forest}

	plain := f.previewWith(spell, "")
	named := f.previewWith(spell, "&exile_ids="+stone.String()+"&waterbend_ids="+stone.String())
	if !plain.OK || len(plain.Plan) != 2 {
		t.Fatalf("{2} off a Mind Stone and a Forest: ok=%v plan=%v, want both", plain.OK, plain.Plan)
	}
	if !named.OK || !equalStrings(sortedCopy(named.Plan), sortedCopy(plain.Plan)) {
		t.Errorf("ability-only params changed the cast preview: %v vs %v", named.Plan, plain.Plan)
	}
	if !f.castsWithAutoTap(spell, game.CastSpellParams{}) {
		t.Fatalf("the preview said ok and the cast was refused")
	}
	if tapped, want := f.tappedAmong(sources), sortedCopy(plain.Plan); !equalStrings(tapped, want) {
		t.Errorf("the cast tapped %v, the preview planned %v", tapped, want)
	}
}
