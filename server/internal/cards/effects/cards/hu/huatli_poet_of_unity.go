package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Huatli, Poet of Unity // Roar of the Fifth People — a transforming card
// (The Lost Caverns of Ixalan, #4470).
//
// Front, Huatli, Poet of Unity — Legendary Creature — Human Warrior Bard
// {2}{G}, 2/3:
//
//	"When Huatli enters, search your library for a basic land card, reveal
//	 it, put it into your hand, then shuffle.
//	 {3}{R/W}{R/W}: Exile Huatli, then return her to the battlefield
//	 transformed under her owner's control. Activate only as a sorcery."
//
// Back, Roar of the Fifth People — Enchantment — Saga:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I — Create two 3/3 green Dinosaur creature tokens.
//	 II — This Saga gains "Creatures you control have '{T}: Add {R}, {G},
//	      or {W}.'"
//	 III — Search your library for a Dinosaur card, reveal it, put it into
//	       your hand, then shuffle.
//	 IV — Dinosaurs you control gain double strike and trample until end
//	      of turn."
//
// The transform is the exile-and-return verb (CR 712.14a), so the Saga is a
// new object that enters with its first lore counter and chapter I triggers.
//
// Chapter II is Urza's Saga's shape one level deeper (ADR 0093 Decision 10,
// amendment 2026-11-10, #5402): a duration grant with no stated duration
// (CR 611.2a; the card's ruling: it applies as long as the Saga remains on
// the battlefield), pinned to the Saga, of the bundle roarChapterII. That
// bundle's one static is a layer-6 grant of a FIXED bundle, roarMana, which
// is mana only — so it opts in with GrantedFixedBundle and the cascade ends
// there. "You" is the Saga's controller, and "creatures you control" is read
// each layer pass, so a creature that enters later has the ability too. If
// the Saga loses all abilities after chapter II, its granted static stops
// existing and the creatures lose the mana ability (CR 613.6, 613.8a); a
// creature that loses all abilities after the grant loses it too (CR 613.7).
// When the Saga is sacrificed after IV, the pinned grant ends with it.
//
// Chapter III's "Dinosaur card" counts a changeling card (CR 702.73a).
// Chapter IV reads the Dinosaurs as it resolves (CR 611.2c).
//
// No simplification.
const (
	HuatliPoetOfUnityOracle = "2e6ed3e6-d2b8-4b75-b7e1-0773da4987a1"
	RoarChapterII           = "roar-of-the-fifth-people/creatures-have-mana"
	roarMana                = "roar-of-the-fifth-people/tap-for-r-g-or-w"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:     HuatliPoetOfUnityOracle,
		Name:         "Huatli, Poet of Unity",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Huatli, Poet of Unity — search for a basic land card, reveal it, put it into your hand", huatliSearchBasic),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{R/W}{R/W}: Exile Huatli, then return her to the battlefield transformed under her owner's control. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{3}{R/W}{R/W}"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
	const roar = "Roar of the Fifth People"
	effects.Register(effects.Spec{
		OracleID:     HuatliPoetOfUnityOracle + "#1",
		Name:         roar,
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: RoarChapterII,
				Static: []game.StaticAbility{effects.GrantedFixedBundle(
					"Creatures you control have \"{T}: Add {R}, {G}, or {W}.\"",
					roarCreaturesYouControl, roarMana)},
				Text: "Creatures you control have \"{T}: Add {R}, {G}, or {W}.\"",
			},
			effects.TapForManaGrant(roarMana, "{R|G|W}", "Add {R}, {G}, or {W}", "{T}: Add {R}, {G}, or {W}."),
		},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(roar, 1, "create two 3/3 green Dinosaur creature tokens"),
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Dinosaur"), N: 2})),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(roar, 2, "this Saga gains \"Creatures you control have '{T}: Add {R}, {G}, or {W}.'\""), roarGainsChapterII),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(roar, 3, "search for a Dinosaur card, reveal it, put it into your hand"), roarSearchDinosaur),
			effects.ChapterTrigger(4, effects.SagaChapterLabel(roar, 4, "Dinosaurs you control gain double strike and trample until end of turn"), roarDinosaursSurge),
		},
	})
}

// huatliSearchBasic is the front face's enter trigger.
func huatliSearchBasic(g *game.Game, item *game.StackItem) error {
	return effects.SearchLibrary{
		Player:    item.Controller,
		Predicate: effects.IsBasicLand,
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Huatli, Poet of Unity — a basic land card",
		Source:    item.SourceCardID,
	}.Apply(effects.NewContext(g, item))
}

// roarCreaturesYouControl is "creatures you control", read with the Saga
// that holds the grant as the source.
func roarCreaturesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return source != nil && target.IsCreature() && target.Controller == source.Controller
}

// roarGainsChapterII is chapter II: the Saga gives ITSELF the granting
// static for the rest of the game, pinned to this object.
func roarGainsChapterII(g *game.Game, item *game.StackItem) error {
	if effects.SourceIsNewObject(g, item) {
		return nil // "this Saga" left and came back: a new object
	}
	saga := item.SourceCardID
	return effects.GrantAbilitiesFor{
		Target:   saga,
		Keys:     []string{RoarChapterII},
		Duration: g.PinnedTo(game.IndefiniteDuration(), saga),
		Label:    "Roar of the Fifth People — gains \"Creatures you control have '{T}: Add {R}, {G}, or {W}.'\"",
	}.Apply(effects.NewContext(g, item))
}

// roarSearchDinosaur is chapter III.
func roarSearchDinosaur(g *game.Game, item *game.StackItem) error {
	return effects.SearchLibrary{
		Player:    item.Controller,
		Predicate: func(c game.Card) bool { return c.HasSubtypeWord("Dinosaur") },
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Roar of the Fifth People — a Dinosaur card",
		Source:    item.SourceCardID,
	}.Apply(effects.NewContext(g, item))
}

// roarDinosaursSurge is chapter IV.
func roarDinosaursSurge(g *game.Game, item *game.StackItem) error {
	return effects.GrantKeywordUntilEOT{
		Match:    effects.And(effects.HasSubtype("Dinosaur"), effects.YouControl()),
		Keywords: []string{"double strike", "trample"},
		Label:    "Roar of the Fifth People — Dinosaurs you control gain double strike and trample until end of turn",
	}.Apply(effects.NewContext(g, item))
}
