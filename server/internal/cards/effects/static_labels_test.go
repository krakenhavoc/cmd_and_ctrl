package effects

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// static_labels_test.go — #2410: every shared static / replacement
// constructor ships a player-facing label, and the labels read as the
// card does.

func TestFamilyConstructorLabels(t *testing.T) {
	goblins := TribeFilter{Tribes: []string{"Goblin"}, Others: true, YoursOnly: true}
	table := []struct {
		name string
		got  string
		want string
	}{
		{"TribalAnthem lord", TribalAnthem(goblins, 1, 1).Label, "Other Goblins you control get +1/+1."},
		{"TribalAnthem whole table", TribalAnthem(TribeFilter{Tribes: []string{"Merfolk"}, Others: true}, 1, 1).Label, "Other Merfolk get +1/+1."},
		{"TribalAnthem irregular plural", TribalAnthem(TribeFilter{Tribes: []string{"Elf"}, YoursOnly: true}, 1, 1).Label, "Elves you control get +1/+1."},
		{"TribalAnthem penalty", TribalAnthem(TribeFilter{YoursOnly: true}, -1, -1).Label, "Creatures you control get -1/-1."},
		{"TribalAnthem chosen", TribalAnthem(TribeFilter{Chosen: true, YoursOnly: true}, 1, 1).Label, "Creatures you control of the chosen type get +1/+1."},
		{"TribalAnthem two tribes", TribalAnthem(TribeFilter{Tribes: []string{"Goblin", "Orc"}}, 1, 0).Label, "Goblins and Orcs get +1/+0."},
		{"TribalKeywordGrant", TribalKeywordGrant(goblins, "haste").Label, "Other Goblins you control have haste."},
		{"TribalScalingAnthem", TribalScalingAnthem(goblins, nil).Label, "Other Goblins you control get +1/+1 for each one the card counts."},
		{"AllCreatureTypesGrant", AllCreatureTypesGrant(TribeFilter{YoursOnly: true}).Label, "Creatures you control are every creature type."},
		{"KeywordGrant", KeywordGrant(nil, "flying").Label, "Grants flying."},
		{"PumpAttached", PumpAttached(2, 2).Label, "Attached creature gets +2/+2."},
		{"PumpAttached penalty", PumpAttached(-1, -1).Label, "Attached creature gets -1/-1."},
		{"PumpAttached mixed", PumpAttached(3, -1).Label, "Attached creature gets +3/-1."},
		{"GrantToAttached", GrantToAttached("flying", "haste").Label, "Attached creature has flying and haste."},
		{"GrantToAttached three", GrantToAttached("flying", "haste", "trample").Label, "Attached creature has flying, haste and trample."},
		{"RemoveFromAttached", RemoveFromAttached("flying").Label, "Attached creature loses flying."},
		{"RestrictAttached", RestrictAttached(game.CantAttack | game.CantBlock).Label, "Attached creature can't attack or block."},
		{"RestrictAttached unblockable", RestrictAttached(game.CantBeBlocked).Label, "Attached creature can't be blocked."},
		{"RestrictSelf", RestrictSelf(game.CantBlock).Label, "Can't block."},
		{"RestrictSelf pair", RestrictSelf(game.CantAttack | game.CantBlock).Label, "Can't attack or block."},
		{"LoseKeywords", LoseKeywords(nil, "flying").Label, "Affected creatures lose flying."},
		{"LoseAndCantHave", LoseAndCantHave(nil, "haste").Label, "Affected creatures lose haste and can't have or gain it."},
		{"CanBlockAdditional", CanBlockAdditional(selfOnly, 1).Label, "Can block an additional creature each combat."},
		{"CanBlockAdditional two", CanBlockAdditional(selfOnly, 2).Label, "Can block 2 additional creatures each combat."},
		{"CanBlockAnyNumber", CanBlockAnyNumber(selfOnly).Label, "Can block any number of creatures."},
		{"LifeGatedPump", LifeGatedPump(nil, 5, 5).Label, "Gives +5/+5 while your life total is high enough."},
		{"LifeGatedKeyword", LifeGatedKeyword(nil, "flying").Label, "Gives flying while your life total is high enough."},
		{"SaddledKeywords", SaddledKeywords("trample").Label, "Has trample while saddled."},
		{"EachLandIsAlso", EachLandIsAlso("Swamp").Label, "Each land is also a Swamp."},
		{"NontokenCreaturesYouControlHave", NontokenCreaturesYouControlHave("haste").Label, "Nontoken creatures you control have haste."},
		{"CreaturesYouControlWithCountersHave", CreaturesYouControlWithCountersHave("+1/+1", true, "trample").Label, "Other creatures you control with a +1/+1 counter on them have trample."},
		{"ThisHasWhileItHasCounter", ThisHasWhileItHasCounter("", "trample").Label, "Has trample while it has a counter on it."},
		{"CantAttackItsOwner", CantAttackItsOwner().Label, "Can't attack its owner."},
		{"RestrictSelf cannot block and unblockable", RestrictSelf(game.CantBlock | game.CantBeBlocked).Label, "Can't block and can't be blocked."},
		{"CantAttackUnlessYouHaveTheCitysBlessing", CantAttackUnlessYouHaveTheCitysBlessing().Label, "Can't attack unless you have the city's blessing."},
		{"impendingNotACreature", impendingNotACreature().Label, "Isn't a creature while it has a time counter on it."},
		{"firstStrikeDuringYourTurn", firstStrikeDuringYourTurn().Label, "Has first strike during your turn."},
		{"ThisHasDuringYourTurn", ThisHasDuringYourTurn("hexproof").Label, "Has hexproof during your turn."},
		{"GoadAttached", GoadAttached().Label, "Attached creature is goaded."},
		{"b07TripleDamageFromYourSources", b07TripleDamageFromYourSources().RowLabel, "If a source you control would deal damage to a permanent or player, it deals triple that damage instead."},
		{"CantAttackUnlessDefendingPlayerControls", CantAttackUnlessDefendingPlayerControls(QuerySubtype("Island")).Label, "Can't attack unless defending player controls an Island."},
		{"CantAttackUnlessDefendingPlayerIsPoisoned", CantAttackUnlessDefendingPlayerIsPoisoned().Label, "Can't attack unless defending player is poisoned."},
		{"CantAttackUnlessDefendingPlayerHasCardsInGraveyard", CantAttackUnlessDefendingPlayerHasCardsInGraveyard(7).Label, "Can't attack unless defending player has 7 or more cards in their graveyard."},
		{"CantAttackAPlayerItAlreadyAttackedThisTurn", CantAttackAPlayerItAlreadyAttackedThisTurn().Label, "Can't attack a player it already attacked this turn."},
		{"AttacksEachCombatWhere", AttacksEachCombatWhere(selfOnly).Label, "Creatures it affects attack each combat if able."},
		{"IsAlsoTheChosenType", IsAlsoTheChosenType().Label, "Is also the chosen creature type."},
		{"ControlAttachedBySource", ControlAttachedBySource().Label, "You control the enchanted permanent."},
		{"LoseAllAbilities", LoseAllAbilities().Label, "Attached permanent loses all abilities."},
		{"b16Anthem", b16Anthem(nil, 1, 1).Label, "Gives +1/+1."},
		{"b16GrantKeywords", b16GrantKeywords(nil, "vigilance").Label, "Gives vigilance."},
		{"b31NotACreatureWhileGraveyardBelow", b31NotACreatureWhileGraveyardBelow(8).Label, "Isn't a creature unless you have 8 or more cards in your graveyard."},
		{"BlockRequirementWhere blocks", BlockRequirementWhere(game.BlockRequirementBlocks, selfOnly).Label, "Creatures it affects block each combat if able."},
		{"BlockRequirementWhere lure", BlockRequirementWhere(game.BlockRequirementLure, selfOnly).Label, "All creatures able to block creatures it affects do so."},
		{"BlockRequirementWhere exactly one", BlockRequirementWhere(game.BlockRequirementExactlyOne, selfOnly).Label, "Creatures it affects must be blocked by exactly one creature if able."},
		{"FilteredLure wall", FilteredLure(game.BlockerFilterWall).Label, "All Walls able to block it do so."},
		{"FilteredLure flying", FilteredLure(game.BlockerFilterFlying).Label, "All creatures with flying able to block it do so."},
		{"LosesAllLandTypes", LosesAllLandTypes(nil).Label, "Affected lands lose all land types."},
		{"LosesAllAbilitiesAndHas", LosesAllAbilitiesAndHas(nil, "x").Label, "Affected permanents lose all abilities and have what the card says they have."},
		{"fraThresholdSelfPT", fraThresholdSelfPT(1, 0).Label, "Gets +1/+0 while there are seven or more cards in your graveyard."},
		{"fraThresholdSelfPT both", fraThresholdSelfPT(2, 1).Label, "Gets +2/+1 while there are seven or more cards in your graveyard."},
		{"fraThresholdSelfKeywords", fraThresholdSelfKeywords("flying", "vigilance").Label, "Has flying and vigilance while there are seven or more cards in your graveyard."},
		{"equippedCreatureIs", equippedCreatureIs(game.Layer6Ability, "Equipped creature has menace.", nil).Label, "Equipped creature has menace."},
		{"whileExactlyTideCounters", whileExactlyTideCounters(3, thisCreatureOnly, 1, 1).Label, "While it has exactly 3 tide counters on it, affected creatures get +1/+1."},
		{"whileExactlyTideCounters one", whileExactlyTideCounters(1, thisCreatureOnly, -1, -1).Label, "While it has exactly 1 tide counter on it, affected creatures get -1/-1."},
		{"SpellsYouControlHave", SpellsYouControlHave(nil, "rebound").Label, "Spells you control have rebound."},
		{"SpellsYouControlHave narrowed", SpellsYouControlHave(Instant(), "flash").Label, "Certain spells you control have flash."},
		{"incarnationAnthem", incarnationAnthem("haste", "Mountain").Label, "While this is in your graveyard and you control a Mountain, creatures you control have haste."},
		{"incarnationAnthem article", incarnationAnthem("flying", "Island").Label, "While this is in your graveyard and you control an Island, creatures you control have flying."},
		// Replacements: the chip's wording is RowLabel, never the CR 616 prompt header.
		{"SelfEntersTapped", SelfEntersTapped().RowLabel, "Enters tapped."},
		{"SelfEntersTappedWithCounters", SelfEntersTappedWithCounters("charge", 2).RowLabel, "Enters tapped with 2 charge counters on it."},
		{"SelfEntersTappedWithCounters one", SelfEntersTappedWithCounters("charge", 1).RowLabel, "Enters tapped with a charge counter on it."},
		{"SelfEntersTappedUnless", SelfEntersTappedUnless(nil).RowLabel, "Enters tapped unless a condition is met."},
		{"EntersTappedUnless", EntersTappedUnless(nil).RowLabel, "Enters tapped unless a condition is met."},
		{"EntersTappedUnlessYouPayLife", EntersTappedUnlessYouPayLife("Blood Crypt", 2).RowLabel, "Enters tapped unless you pay 2 life."},
		{"SelfEntersPrepared", SelfEntersPrepared().RowLabel, "Enters prepared."},
		{"SelfEntersWithCountersIfNoManaSpent", SelfEntersWithCountersIfNoManaSpent("+1/+1", 3).RowLabel, "Enters with 3 +1/+1 counters on it if no mana was spent to cast it."},
		{"DoublesEntering", DoublesEntering(nil).RowLabel, "Abilities that trigger when a permanent enters trigger an additional time."},
		{"SuppressesEntering", SuppressesEntering(nil).RowLabel, "Some permanents entering don't cause abilities to trigger."},
		{"Devour", Devour("X", 2).RowLabel, "Devour 2: you may sacrifice any number of creatures as it enters; it enters with 2 +1/+1 counters on it for each."},
		{"Devour one", Devour("X", 1).RowLabel, "Devour 1: you may sacrifice any number of creatures as it enters; it enters with a +1/+1 counter on it for each."},
		{"DevourPaying", DevourPaying("X", 1, 1, 0).RowLabel, "Devour 1: you may sacrifice any number of creatures as it enters; it enters with a +1/+1 counter on it for each."},
		{"frEntersTappedUnlessYouBeholdAJace", frEntersTappedUnlessYouBeholdAJace("X").RowLabel, "Enters tapped unless you control a Jace or reveal one from your hand."},
		{"EntersAsCopyOfUntilEndOfTurn", EntersAsCopyOfUntilEndOfTurn("X", nil, nil).RowLabel, "May enter as a copy of another permanent until end of turn."},
		{"PreventDamageDealtTo", PreventDamageDealtTo(PreventionStatic{To: ToThisCreature, Then: removeACounterFromThisBody, Label: "Phantom — prevent damage to it and remove a +1/+1 counter"}).RowLabel, "Prevent damage to it and remove a +1/+1 counter."},
		{"PreventDamageDealtTo while", PreventDamageDealtTo(PreventionStatic{To: ToThisCreature, While: WhileItHasAPlusOneCounter, Then: removeACounterFromThisBody, Label: "Knight — prevent damage to it"}).RowLabel, "Prevent damage to it. Only while its condition on the card holds."},
		{"PreventDamageASourceWouldDeal", PreventDamageASourceWouldDeal(PreventionStatic{To: ToYou, Then: nineLivesCounterBody, Label: "Nine Lives — prevent the damage and put an incarnation counter on it"}).RowLabel, "Prevent the damage and put an incarnation counter on it."},
		{"redirectYourDamageToAttached", redirectYourDamageToAttached("Pariah — damage to you is dealt to enchanted creature instead").RowLabel, "Damage to you is dealt to enchanted creature instead."},
		{"redirectYourDamageToThis", redirectYourDamageToThis("Archangel — damage to you is dealt to it instead").RowLabel, "Damage to you is dealt to it instead."},
		{"CreaturesEnteringDontTrigger", CreaturesEnteringDontTrigger("Torpor Orb").RowLabel, "Creatures entering don't cause abilities to trigger."},
	}
	for _, c := range table {
		if c.got != c.want {
			t.Errorf("%s: label = %q, want %q", c.name, c.got, c.want)
		}
	}
}

// The spent-entry replacements are all marked, so the ◆ chip drops them
// once the permanent is in play.
func TestEntryReplacementsAreMarkedEntryOnly(t *testing.T) {
	for name, r := range map[string]game.ReplacementEffect{
		"SelfEntersTapped":             SelfEntersTapped(),
		"SelfEntersTappedWithCounters": SelfEntersTappedWithCounters("charge", 1),
		"SelfEntersTappedUnless":       SelfEntersTappedUnless(nil),
		"EntersTappedUnless":           EntersTappedUnless(nil),
		"EntersTappedUnlessYouPayLife": EntersTappedUnlessYouPayLife("X", 2),
		"SelfEntersPrepared":           SelfEntersPrepared(),
	} {
		if !r.EntryOnly {
			t.Errorf("%s is not EntryOnly", name)
		}
	}
}

// A tapland loses its ◆ chip once it is on the battlefield; the same
// land in hand keeps the promise. Run against the real catalog.
func TestRegisteredTaplandChipIsHiddenOnTheBattlefield(t *testing.T) {
	spec := specByName(t, "Temple of Abandon")
	c := game.Card{Name: spec.Name, OracleID: spec.OracleID}
	var inHand, inPlay []string
	for _, r := range game.AbilityRowsOf(c) {
		if r.Kind == game.AbilityRowStatic {
			inHand = append(inHand, r.Label)
		}
	}
	for _, r := range game.AbilityRowsOnBattlefield(c) {
		if r.Kind == game.AbilityRowStatic {
			inPlay = append(inPlay, r.Label)
		}
	}
	if len(inHand) != 1 || inHand[0] != "Enters tapped." {
		t.Errorf("in hand: static rows = %q, want [Enters tapped.]", inHand)
	}
	if len(inPlay) != 0 {
		t.Errorf("on the battlefield: static rows = %q, want none", inPlay)
	}
}

// Card-specific wording on a handful of the most-played cards.
func TestMostPlayedCardsReadInPlainWords(t *testing.T) {
	for name, want := range map[string]string{
		"Glorious Anthem":             "Creatures you control get +1/+1.",
		"Chromatic Lantern":           "Lands you control have \"{T}: Add one mana of any color.\"",
		"Blood Moon":                  "Nonbasic lands are Mountains.",
		"Urborg, Tomb of Yawgmoth":    "Each land is a Swamp in addition to its other land types.",
		"Anger":                       "While this is in your graveyard and you control a Mountain, creatures you control have haste.",
		"Panharmonicon":               "Abilities that trigger when an artifact or creature enters trigger an additional time.",
		"Reliquary Tower":             "You have no maximum hand size.",
		"Lure":                        "All creatures able to block enchanted creature do so.",
		"Alpine Moon":                 "Lands your opponents control with the chosen name lose all land types and abilities and have \"{T}: Add one mana of any color.\"",
		"Homarid":                     "Gets -1/-1 while it has exactly 1 tide counter on it.",
		"Cast Through Time":           "Instant and sorcery spells you control have rebound.",
		"Psychosis Crawler":           "Power and toughness are each equal to the number of cards in your hand.",
		"Stridehangar Automaton":      "Thopters you control get +1/+1.",
		"Uthros Research Craft":       "Gets +1/+0 for each artifact you control.",
		"Cybermen Squadron":           "Nonlegendary artifact creatures you control have myriad.",
		"Gemcutter Buccaneer":         "Treasures you control have \"Equipped creature gets +2/+0\", \"Equip Pirate {1}\" and \"Equip {3}\".",
		"Changeling Outcast":          "Can't block and can't be blocked.",
		"Wayward Swordtooth":          "Can't attack unless you have the city's blessing.",
		"Adeline, Resplendent Cathar": "Power is equal to the number of creatures you control.",
		"Mirror Box":                  "Each legendary creature you control gets +1/+1.",
		"Overlord of the Hauntwoods":  "Isn't a creature while it has a time counter on it.",
		"Thassa, God of the Sea":      "Isn't a creature unless your devotion to blue is five or more.",
		"Tarmogoyf":                   "Power is the number of card types among cards in all graveyards; toughness is that plus 1.",
		// #3133: worded from their structured slots.
		"Omniscience":                   "You may cast spells from your hand without paying their mana costs.",
		"The Reality Chip":              "Play a land or cast a spell from the top of your library",
		"Chromatic Orrery":              "You may spend mana as though it were mana of any color.",
		"Mycosynth Lattice":             "Players may spend mana as though it were mana of any color.",
		"Sakashima of a Thousand Faces": "The legend rule doesn't apply to permanents you control.",
		"Mirror Gallery":                "The legend rule doesn't apply.",
		"Leyline Tyrant":                "You don't lose unspent red mana as steps and phases end.",
		"Frenzied Baloth":               "Combat damage can't be prevented.",
		"Questing Beast":                "Combat damage dealt by creatures you control can't be prevented.",
		"Sunspine Lynx":                 "Damage can't be prevented.",
		"Jodah, Archmage Eternal":       "You may pay {W}{U}{B}{R}{G} rather than pay the mana cost for spells you cast.",
		"Ulvenwald Hydra":               "Power and toughness are each equal to the number of lands you control.",
		"Spear of Heliod":               "Creatures you control get +1/+1.",
		"Toxrill, the Corrosive":        "Creatures you don't control get -1/-1 for each slime counter on them.",
	} {
		spec := specByName(t, name)
		found := false
		for _, r := range game.AbilityRowsOf(game.Card{Name: spec.Name, OracleID: spec.OracleID}) {
			if r.Kind == game.AbilityRowStatic && r.Label == want {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: no static row reads %q", name, want)
		}
	}
}

// engineWords are the vocabulary a label must not borrow from the
// engine (docs/adding-cards.md's Caveats tone).
var engineWords = regexp.MustCompile(`(?i)\b(layer|replacement|static ability|instanceid|oracle|closure|predicate|sublayer|timestamp)\b`)

// Every worded static row in the catalog reads as a sentence a player
// can use: not empty, no engine vocabulary.
func TestStaticRowLabelsUseNoEngineVocabulary(t *testing.T) {
	bad := map[string]string{}
	for _, s := range All() {
		for _, r := range game.AbilityRowsOf(game.Card{Name: s.Name, OracleID: s.OracleID}) {
			if r.Kind != game.AbilityRowStatic || r.Generic {
				continue
			}
			if engineWords.MatchString(r.Label) {
				bad[s.Name] = r.Label
			}
		}
	}
	for n, l := range bad {
		t.Errorf("%s: label %q uses engine vocabulary", n, l)
	}
}

// familyLabelExemptions are constructors that return a StaticAbility or
// ReplacementEffect and so would be expected to carry a label, with the
// reason they do not. A constructor not listed here and without a label
// fails the guard below: a whole family of cards would read "Power/
// toughness effect" until someone noticed (#2410).
var familyLabelExemptions = map[string]string{
	// The row is worded by the constructor this one calls.
	"TribalAbilityGrant":                                 "delegates to GrantAbilities, whose row reads the granted bundle's printed text",
	"GrantAbilities":                                     "the row reads the granted bundle's own printed text (game.grantedTextOf)",
	"GrantAbilitiesToAttached":                           "delegates to GrantAbilities",
	"GrantAbilitiesInZone":                               "delegates to GrantAbilities",
	"GrantAbilitiesToYourPlaneswalkers":                  "delegates to GrantAbilities",
	"AttacksEachCombat":                                  "sets the label on the value AttacksEachCombatWhere returns",
	"NontokenCreaturesYouControlHave":                    "sets the label on the value KeywordGrant returns",
	"LifeGatedKeyword":                                   "sets the label on the value KeywordGrant returns",
	"cantAttackOwnerStatic":                              "delegates to selfAttackTargetRestriction",
	"CantAttackItsOwner":                                 "delegates to selfAttackTargetRestriction",
	"CantAttackItsOwnerOrItsOwnersPlaneswalkers":         "delegates to selfAttackTargetRestriction",
	"CantAttackUnlessDefendingPlayerControls":            "delegates to selfAttackTargetRestriction",
	"CantAttackUnlessDefendingPlayerHasCardsInGraveyard": "delegates to selfAttackTargetRestriction",
	"CantAttackUnlessYouControlMoreThanDefendingPlayer":  "delegates to selfAttackTargetRestriction",
	"EntersOnlyIfYouDiscardFromHand":                     "delegates to entersOnlyIf",
	"EntersOnlyIfYouSacrifice":                           "delegates to entersOnlyIf",
	"EntersAsCopyOf":                                     "delegates to entersAsCopy",
	"EntersAsCopyOfFromCast":                             "delegates to entersAsCopy",
	"MaxSpeedStatic":                                     "wraps a static that carries its own label",
	"MaxSpeedReplacement":                                "wraps a replacement that carries its own label",
	"StaticWhenChosen":                                   "wraps a static that carries its own label",
	"ThresholdKeywords":                                  "a keyword row: the keyword chips show it, not the ◆ chip",
	"RestrictSelf":                                       "labelled through selfRestrictionLabel",
	// The wording is the card's own text.
	"CreaturesYouControlEnterWithCountersPerManaFrom": "takes the card's own label",
	"TokensDoubled":                    "takes the card's own label",
	"AnyPlayersTokensDoubled":          "takes the card's own label",
	"tokenMultiplier":                  "takes the card's own label",
	"OpponentsMillTwice":               "takes the card's own label",
	"OpponentsMillPlus":                "takes the card's own label",
	"ShuffleIntoOwnersLibraryInstead":  "takes the card's own label",
	"KeywordActionBecomes":             "takes the card's own label",
	"ProliferateTwice":                 "takes the card's own label",
	"ScryPlusOne":                      "takes the card's own label",
	"SurveilPlusOne":                   "takes the card's own label",
	"RegenerateIfThisWouldBeDestroyed": "a single-card ability with its own Label",
	"RestrictUntilEOT":                 "not a battlefield static",
	"GrantAbility":                     "a one-shot primitive, not a static row",
	"impendingNotACreature":            "keyword plumbing the chip hides",
	// #3129: delegating wrappers over a constructor that carries the wording.
	"BlockRequirementWhere":           "delegates to blockRequirementStatic, which words the row from the requirement kind",
	"Devour":                          "delegates to devourOver, which words the row from the count",
	"DevourPaying":                    "delegates to Devour",
	"grantToCommanderCreaturesYouOwn": "delegates to GrantAbilities, whose row reads the granted bundle's printed text",
	"phialGated":                      "narrows a replacement that already carries its own label",
	"PreventDamageDealtTo":            "delegates to PreventionStatic.replacement, which words the row from the card's header",
	"PreventDamageASourceWouldDeal":   "delegates to PreventionStatic.replacement, which words the row from the card's header",
	"redirectYourDamageToAttached":    "delegates to staticRedirection, which words the row from the card's header",
	"redirectYourDamageToThis":        "delegates to staticRedirection, which words the row from the card's header",
}

// familyLabelPending are constructors that should carry wording and do
// not yet: the remaining work of #2410, written down so it cannot be
// forgotten and so the guard stays green. The list only shrinks. An
// entry that gains a Label fails TestFamilyConstructorsCarryALabel
// until it is deleted from here. #3129 worded the last of them; a
// constructor whose meaning depends on the card goes here again only
// with a reason, and its cards carry Labelled(...) wording.
var familyLabelPending = map[string]bool{}

// Every function in this package that returns a StaticAbility or a
// ReplacementEffect (or a slice of them) names a Label or RowLabel, or
// is listed with a reason. This is the guard: a new family constructor
// that ships without wording fails here instead of reading generically
// on every card that calls it.
func TestFamilyConstructorsCarryALabel(t *testing.T) {
	fset := token.NewFileSet()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	var unlabelled []string
	stillPending := map[string]bool{}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		file, err := parser.ParseFile(fset, f, src, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range file.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Body == nil || fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
				continue
			}
			rt := fn.Type.Results.List[0].Type
			ret := string(src[fset.Position(rt.Pos()).Offset:fset.Position(rt.End()).Offset])
			if ret != "game.StaticAbility" && ret != "game.ReplacementEffect" {
				continue
			}
			// Only constructors: a function with arguments that builds the
			// value (a card file's zero-argument helper is the card's own).
			body := string(src[fset.Position(fn.Body.Pos()).Offset:fset.Position(fn.Body.End()).Offset])
			if strings.Contains(body, "Label") {
				continue
			}
			if fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
				continue // a card's own zero-argument helper; its Spec line carries the wording
			}
			if _, exempt := familyLabelExemptions[fn.Name.Name]; exempt {
				continue
			}
			if familyLabelPending[fn.Name.Name] {
				stillPending[fn.Name.Name] = true
				continue
			}
			unlabelled = append(unlabelled, f+": "+fn.Name.Name)
		}
	}
	for name := range familyLabelPending {
		if !stillPending[name] {
			t.Errorf("familyLabelPending lists %s, which now carries a label (or no longer exists): delete the entry", name)
		}
	}
	sort.Strings(unlabelled)
	for _, u := range unlabelled {
		t.Errorf("%s returns a static / replacement with no Label or RowLabel: give it player-facing wording, or list it in familyLabelExemptions with a reason", u)
	}
}
