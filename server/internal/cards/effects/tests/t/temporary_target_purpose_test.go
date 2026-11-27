package testst

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// temporary_target_purpose_test.go — #4302: a change to a target that
// ends with the turn declares which way it cuts, so the bot gives a
// benefit to its own creature and a harm to an opponent's.

func TestUntilEndOfTurnGuard(t *testing.T) {
	entries := func(ts ...game.TargetPurpose) game.Purpose { return effects.ForTargets(ts...) }
	creature := func() *game.TargetSpec { return effects.TargetCreature("target creature") }
	cases := []struct {
		name string
		spec effects.Spec
		want string // "" = accepted
	}{
		{"a keyword grant", effects.Spec{Targets: creature(), Purpose: entries(effects.BenefitUntilEOT(0, 0, 0, "lifelink"))}, ""},
		{"a pump", effects.Spec{Targets: creature(), Purpose: entries(effects.BenefitUntilEOT(0, 2, 2))}, ""},
		{"a direction alone", effects.Spec{Targets: creature(), Purpose: entries(effects.BenefitUntilEOT(0, 0, 0))}, ""},
		{"a shrink", effects.Spec{Targets: creature(), Purpose: entries(effects.HarmUntilEOT(0, -2, -2))}, ""},
		{"a harm alone", effects.Spec{Targets: creature(), Purpose: entries(effects.HarmUntilEOT(0, 0, 0))}, ""},
		{"can't be blocked", effects.Spec{Targets: creature(), Purpose: entries(effects.CantBeBlockedUntilEOT(0, 0, 0))}, ""},
		{"can't be blocked with a pump", effects.Spec{Targets: creature(), Purpose: entries(effects.CantBeBlockedUntilEOT(0, 2, 0))}, ""},
		{"can't block", effects.Spec{Targets: creature(), Purpose: entries(effects.CantBlockUntilEOT(0, 0, 0))}, ""},
		{"can't block with a shrink", effects.Spec{Targets: creature(), Purpose: entries(effects.CantBlockUntilEOT(0, -2, 0))}, ""},
		{"a benefit that can't block", effects.Spec{Targets: creature(), Purpose: entries(game.TargetPurpose{
			UntilEndOfTurn: game.TargetBenefit, Pump: &game.Pump{CantBlock: true}})}, "that is a harm"},
		{"a self pump that can't block", effects.Spec{Activated: []effects.ActivatedAbility{
			{Label: "a", Purpose: game.Purpose{Pump: &game.Pump{Power: 2, CantBlock: true}}},
		}}, "own source"},
		{"a harm that makes it unblockable", effects.Spec{Targets: creature(), Purpose: entries(game.TargetPurpose{
			UntilEndOfTurn: game.TargetHarm, Pump: &game.Pump{CantBeBlocked: true}})}, "unblockability"},
		{"on an activated row", effects.Spec{Activated: []effects.ActivatedAbility{
			{Label: "a", Targets: creature(), Purpose: entries(effects.BenefitUntilEOT(0, 0, 0, "trample"))},
		}}, ""},
		{"an unknown direction", effects.Spec{Targets: creature(), Purpose: entries(game.TargetPurpose{UntilEndOfTurn: "both"})}, "neither benefit nor harm"},
		{"a pump with no direction", effects.Spec{Targets: creature(), Purpose: entries(game.TargetPurpose{Pump: &game.Pump{Power: 1}})}, "without UntilEndOfTurn"},
		{"a benefit that shrinks", effects.Spec{Targets: creature(), Purpose: entries(game.TargetPurpose{
			UntilEndOfTurn: game.TargetBenefit, Pump: &game.Pump{Power: -1, Toughness: -1}})}, "that is a harm"},
		{"a harm that pumps", effects.Spec{Targets: creature(), Purpose: entries(game.TargetPurpose{
			UntilEndOfTurn: game.TargetHarm, Pump: &game.Pump{Power: 2}})}, "adds power"},
		{"a harm that grants a keyword", effects.Spec{Targets: creature(), Purpose: entries(game.TargetPurpose{
			UntilEndOfTurn: game.TargetHarm, Pump: &game.Pump{Keywords: []string{"flying"}}})}, "adds power"},
		{"a pump that gives nothing", effects.Spec{Targets: creature(), Purpose: entries(game.TargetPurpose{
			UntilEndOfTurn: game.TargetBenefit, Pump: &game.Pump{}})}, "gives nothing"},
		{"a capitalised keyword", effects.Spec{Targets: creature(), Purpose: entries(effects.BenefitUntilEOT(0, 0, 0, "Lifelink"))}, "not lowercase"},
		{"on a player", effects.Spec{Targets: effects.TargetPlayer("target player"), Purpose: entries(effects.BenefitUntilEOT(0, 0, 0))}, "cannot target a permanent"},
		{"on a spell", effects.Spec{Targets: effects.TargetSpell("target spell"), Purpose: entries(effects.HarmUntilEOT(0, 0, 0))}, "cannot target a permanent"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.spec.Name = "Test Card"
			var msg string
			func() {
				defer func() {
					if r := recover(); r != nil {
						msg, _ = r.(string)
					}
				}()
				effects.CheckSpecPurposes(c.spec)
			}()
			switch {
			case c.want == "" && msg != "":
				t.Errorf("refused: %s", msg)
			case c.want != "" && !strings.Contains(msg, c.want):
				t.Errorf("got %q, want a refusal containing %q", msg, c.want)
			}
		})
	}
}

// untilEndOfTurnRow finds an activated row whose printed text changes a
// target until end of turn: "target creature gains lifelink until end of
// turn", "gets -2/-2 until end of turn", "can't be blocked this turn".
var untilEndOfTurnRow = regexp.MustCompile(`(?i)\btarget\b[^.:]*\b(gains?|gets|loses|can't block|can't be blocked)\b[^.]*\b(until end of turn|this turn)`)

// untilEndOfTurnExempt are the rows the pattern finds that do not declare
// a direction, each for a reason a direction cannot carry.
var untilEndOfTurnExempt = map[string]string{
	// "+X/+X or -X/-X": the controller picks the sign at resolution.
	"Liliana of the Dark Realms": "the sign is chosen at resolution",
	// Moves Zack Fair's counters and Equipment onto the creature for good.
	"Zack Fair": "part of what it does lasts",
	// "Target opponent loses X life, where X is the amount of life you
	// gained this turn": "this turn" bounds the count, not a change to the
	// target.
	"Midnight Snack": "this turn bounds the count, not a change to the target",
	// The target is a card in a graveyard, which a permanent direction
	// cannot describe: it gains embalm until end of turn.
	"Cursecloth Wrappings": "the target is a graveyard card, not a permanent",
	"Sapphire Collector":   "the target is a graveyard card, not a permanent",

	// The same: the target is a legendary creature card in a graveyard, and
	// it gains escape until end of turn (S82 GAP2 batch 2).
	"Confession Dial": "the target is a graveyard card, not a permanent",
	// The same: the target is a creature card in a graveyard, and it gains encore
	// until end of turn (S82 OLD batch 083).
	"Araumi of the Dead Tide": "the target is a graveyard card, not a permanent",
	// The target is the opponent given an emblem; the "until end of turn" is
	// the emblem's, about the creatures that attack them.
	"Garruk, Apex Predator": "the target is an opponent given an emblem",

	// Tales of Middle-earth Commander's lands (S82 GAP batch 120): each takes
	// an ability away from its target until end of turn. The pattern reads
	// "loses" as neither a harm nor a benefit and expects a benefit, but
	// removing swampwalk, first strike or a landwalk is a harm to the target.
	"Urborg":     "removes an ability from its target; the check reads 'loses' as a benefit",
	"Hammerheim": "removes an ability from its target; the check reads 'loses' as a benefit",

	// Secret Lair Drop partner commanders (S82 OLD batch 14): the target is a
	// player dealt damage; the "until end of turn" is the source's own
	// indestructible, not a change to the target.
	"Ellie, Vengeful Hunter": "the target is a player dealt damage; the until end of turn is the source's",

	// Coldsnap (S82 OLD batch 561): takes flying from its target until end of
	// turn, which is a harm; the pattern reads "loses" as a benefit (Urborg's case).
	"Adarkar Windform": "removes an ability from its target; the check reads 'loses' as a benefit",

	// Ravnica: City of Guilds (S82 OLD batch 575): the first target gets +2/+2
	// and the second gets -2/-2. The first target's entry declares the benefit,
	// which is true; the check reads the "gets -" of the second clause as a
	// harm on every target.
	"Drooling Groodion": "the first target benefits; the 'gets -' the check reads belongs to the second target",

	// Planeshift (S82 OLD batch 639): the first target is the opponent who reveals
	// a card; the creature that gets +X/+X or -X/-X is the second target, whose
	// entry (slot 1) declares the direction. The check reads slot 0 only.
	"Planeswalker's Favor": "the first target is the opponent who reveals a card; the pump's direction is declared on slot 1",
	"Planeswalker's Scorn": "the first target is the opponent who reveals a card; the shrink's direction is declared on slot 1",

	// Mercadian Masques (S82 OLD batch 681): takes the target for the turn,
	// so the haste it gains is the thief's, and the creature is priced as the
	// removal-for-a-turn it is; a benefit declaration would aim it at the
	// bot's own creatures.
	"Overtaker": "gains control of its target until end of turn; the target is an opposing creature, not a recipient",

	// Urza's Legacy (S82 OLD batch 716): takes flying, first strike or trample
	// from its target until end of turn, which is a harm; the pattern reads
	// "loses" as a benefit (Urborg's case).
	"Walking Sponge": "removes an ability from its target; the check reads 'loses' as a benefit",

	// Mirage (S82 OLD batch 793): deals damage to its target and takes flying
	// from it until end of turn, which is a harm; the pattern reads "loses" as
	// a benefit (Urborg's case).
	"Burning Palm Efreet": "removes an ability from its target; the check reads 'loses' as a benefit",
}

// TestUntilEndOfTurnRowsDeclareTheirDirection holds every catalog
// activated row that changes a target until end of turn to declaring
// which way the change cuts on that target's entry (#4302), and the
// direction to the printed words: "gets -" and "can't block" are harms,
// "gets +", "gains" and "can't be blocked" benefits. A row that declares
// nothing is priced as though its target were removed, which is how
// Heliod, Sun-Crowned gave an opponent's creature lifelink.
func TestUntilEndOfTurnRowsDeclareTheirDirection(t *testing.T) {
	found := 0
	for _, spec := range effects.All() {
		for i, row := range spec.Activated {
			if row.Targets == nil || !untilEndOfTurnRow.MatchString(row.Label) {
				continue
			}
			if _, ok := untilEndOfTurnExempt[spec.Name]; ok {
				continue
			}
			found++
			var e *game.TargetPurpose
			for _, x := range row.Purpose.Targets.List() {
				if x.Slot == 0 {
					e = &x
				}
			}
			if e == nil || e.UntilEndOfTurn == "" {
				t.Errorf("%s ability %d %q declares no UntilEndOfTurn on its target (#4302)", spec.Name, i, row.Label)
				continue
			}
			harm := strings.Contains(row.Label, "gets -") || strings.Contains(row.Label, "can't block")
			checkDeclaredChange(t, fmt.Sprintf("%s ability %d %q", spec.Name, i, row.Label), row.Label, harm, e)
		}
	}
	if found < 70 {
		t.Errorf("found %d rows that change a target until end of turn, want at least the 74 #4302 declared", found)
	}
}

// TestHeliodDeclaresLifelinkForItsTarget: the card that started it.
func TestHeliodDeclaresLifelinkForItsTarget(t *testing.T) {
	s, ok := effects.Lookup("63e596a2-9126-4af6-8782-c38687d664ad")
	if !ok || len(s.Activated) != 1 {
		t.Fatalf("Heliod, Sun-Crowned is not registered with its one activated ability")
	}
	got := s.Activated[0].Purpose.Targets.List()
	if len(got) != 1 || got[0].UntilEndOfTurn != game.TargetBenefit || got[0].Pump == nil ||
		len(got[0].Pump.Keywords) != 1 || got[0].Pump.Keywords[0] != "lifelink" {
		t.Errorf("Heliod's target entry = %+v, want a lifelink benefit until end of turn", got)
	}
}

// checkDeclaredChange holds one declared change until end of turn to the
// words it was declared for: the direction (a harm when the words take
// something away), every keyword it grants among those words, and "can't
// be blocked this turn" declared as CantBeBlocked when the words say it
// with no "except" (#4312).
func checkDeclaredChange(t *testing.T, where, text string, harm bool, e *game.TargetPurpose) {
	t.Helper()
	if want := map[bool]game.TargetEffect{true: game.TargetHarm, false: game.TargetBenefit}[harm]; e.UntilEndOfTurn != want {
		t.Errorf("%s declares %q, want %q", where, e.UntilEndOfTurn, want)
	}
	lower := strings.ToLower(text)
	pm := e.Pump
	if pm == nil {
		pm = &game.Pump{}
	}
	for _, k := range pm.Keywords {
		if !strings.Contains(lower, k) {
			t.Errorf("%s declares the keyword %q its text does not grant", where, k)
		}
	}
	if pm.CantBeBlocked && !strings.Contains(lower, "can't be blocked") {
		t.Errorf("%s declares CantBeBlocked, which its text does not say", where)
	}
	unblockable := strings.Contains(lower, "can't be blocked this turn") && !strings.Contains(lower, "except")
	if unblockable && !harm && !pm.CantBeBlocked {
		t.Errorf("%s says \"can't be blocked this turn\" and does not declare CantBeBlocked (#4312)", where)
	}
	// #4977: "can't block" is declared as CantBlock, so it is priced by
	// the attack it opens and not as removal.
	if pm.CantBlock && !strings.Contains(lower, "can't block") {
		t.Errorf("%s declares CantBlock, which its text does not say (#4977)", where)
	}
	if harm && strings.Contains(lower, "can't block") && !pm.CantBlock {
		t.Errorf("%s says \"can't block\" and does not declare CantBlock (#4977)", where)
	}
}

// untilEndOfTurnSpellExempt are the spells the pattern finds that declare
// no direction, each for a reason a direction cannot carry (#4312).
var untilEndOfTurnSpellExempt = map[string]string{
	// "Untap target permanent and gain control of it until end of turn":
	// the pattern reads "gain control" as a grant, but taking a permanent
	// is a change of control, which a direction cannot carry; Act of
	// Treason, whose words the pattern does not catch, declares none.
	"Word of Seizing": "a change of control, not a change to the target",
	// "Untap target creature an opponent controls and gain control of it
	// until end of turn" (Commander 2011): Word of Seizing's change of control.
	"Ray of Command": "a change of control, not a change to the target",
	// "Untap target creature and gain control of it until end of turn"
	// (Ravnica: City of Guilds; the white-mana rider grants the creature a
	// lifegain trigger): Word of Seizing's change of control.
	"Flash Conscription": "a change of control, not a change to the target",
	// "At the beginning of the next end step, destroy that creature if it
	// attacked this turn": the trample and power are a benefit, the
	// destruction a harm that lasts.
	"Berserk": "part of it is a lasting harm",
	// Both targets gain indestructible and then every creature is
	// destroyed: the grant is what the card's own sweep spares, which a
	// direction cannot say, and one of the two is an opponent's.
	"Fated Clash": "the grant answers its own sweep",
	// The granted ability redirects damage onto the creature.
	"Martyrdom": "the grant hurts the creature it helps",
	// "+3/-3 and loses all creature types": the power is a benefit, the
	// toughness and the type loss a harm (#4691).
	"Nameless Inversion": "partly a benefit and partly a harm",
	// "+1/-1 until end of turn. When it dies under your control this turn,
	// draw two cards": the power is a benefit, the toughness a harm, and the
	// card is cast on your own creature so that it dies.
	"Desperate Measures": "partly a benefit and partly a harm",
	// Mode one, "target creature gets +3/-3 until end of turn", is Nameless
	// Inversion's change without the type loss: the power is a benefit, the
	// toughness a harm, and it is cast as often to kill an X/3 as to pump.
	"Eldrazi Confluence": "partly a benefit and partly a harm",
	// "Target creature gets +2/-2 until end of turn" (#5502): the power is
	// a benefit, the toughness a harm, and it is cast as often to kill an
	// X/2 as to pump an attacker.
	"Lash of Malice": "partly a benefit and partly a harm",
	// "+3/-3 until end of turn" (S82 GAP): the same trade as Eldrazi
	// Confluence's first mode.
	"Spatial Contortion": "partly a benefit and partly a harm",
	// Mode two, "target creature gets +2/-1 until end of turn" (#5595): the
	// power is a benefit, the toughness a harm, and it is cast as often to
	// kill an X/1 as to pump an attacker.
	"Piracy Charm": "partly a benefit and partly a harm",

	// The target creature is copied, not changed: its copies are the lands it
	// does not target, until end of turn.
	"March from Velis Vel": "the target is copied, not changed",

	// "+4/-4 until end of turn" (S82 GAP): the same trade as Spatial
	// Contortion's, the power a benefit and the toughness a harm.
	"Auger Spree": "partly a benefit and partly a harm",
}

var targetWord = regexp.MustCompile(`(?i)\btarget\b`)

// TestUntilEndOfTurnSpellsDeclareTheirDirection is the spells' half of
// TestUntilEndOfTurnRowsDeclareTheirDirection (#4312): every catalog
// instant, sorcery and mode whose printed text changes a permanent it
// targets until end of turn declares the direction on that target's
// entry, with the keywords and "can't be blocked" its words say. A card's
// text is its oracle fixture, a mode's its label. The slot is the number
// of "target"s before the change in that text. A clause that cannot pick
// a permanent ("creatures target player controls") changes no target,
// and an Aura's enchant clause is not a target the spell changes. An
// undeclared spell is priced as though its target were removed, so a
// Giant Growth was worth casting at an opponent's creature.
func TestUntilEndOfTurnSpellsDeclareTheirDirection(t *testing.T) {
	found := 0
	check := func(where, text string, clauses []game.TargetClause, p game.Purpose) {
		for _, loc := range untilEndOfTurnRow.FindAllStringIndex(text, -1) {
			words := text[loc[0]:loc[1]]
			slot := len(targetWord.FindAllStringIndex(text[:loc[0]], -1))
			if len(clauses) == 1 {
				// Every "target" names the one clause: Divine
				// Resilience's kicked "instead any number of target
				// creatures" is the same pick, widened.
				slot = 0
			}
			if slot >= len(clauses) {
				t.Errorf("%s: the change %q is to slot %d of a statement with %d clauses", where, words, slot, len(clauses))
				continue
			}
			if c := clauses[slot]; !slices.Contains(c.Zones, game.ZoneBattlefield) || strings.HasPrefix(strings.ToLower(c.Label), "enchant") {
				continue
			}
			found++
			var e *game.TargetPurpose
			for _, x := range p.Targets.List() {
				if x.Slot == slot {
					e = &x
				}
			}
			if e == nil || e.UntilEndOfTurn == "" {
				t.Errorf("%s declares no UntilEndOfTurn on slot %d for %q (#4312)", where, slot, words)
				continue
			}
			harm := strings.Contains(words, "gets -") || strings.Contains(words, "can't block") || strings.Contains(words, "loses")
			checkDeclaredChange(t, fmt.Sprintf("%s slot %d", where, slot), words, harm, e)
		}
	}
	for _, spec := range effects.All() {
		if _, ok := untilEndOfTurnSpellExempt[spec.Name]; ok {
			continue
		}
		if spec.Modes != nil {
			for i, o := range spec.Modes.Options {
				check(fmt.Sprintf("%s mode %d %q", spec.Name, i, o.Label), o.Label, o.Targets.Clauses(), o.Purpose)
			}
			continue
		}
		if spec.Targets == nil {
			continue
		}
		if text := effectstest.SpellText(t, spec.OracleID); text != "" {
			check(spec.Name, text, spec.Targets.Clauses(), spec.Purpose)
		}
	}
	t.Logf("%d spell clauses change a target until end of turn", found)
	if found < 100 {
		t.Errorf("found %d spell clauses that change a target until end of turn, want at least the 106 #4312 declared", found)
	}
}

// TestDeclaredChangesOfTheTricksDeck: the cards the arena's tricks-battle
// deals (#4312), each declaring what the bot prices.
func TestDeclaredChangesOfTheTricksDeck(t *testing.T) {
	entry := func(oracle string, activated bool) game.TargetPurpose {
		t.Helper()
		s, ok := effects.Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", oracle)
		}
		p := s.Purpose
		if activated {
			p = s.Activated[len(s.Activated)-1].Purpose
		}
		got := p.Targets.List()
		if len(got) != 1 || got[0].Pump == nil {
			t.Fatalf("%s declares %+v, want one entry with a pump", s.Name, got)
		}
		return got[0]
	}
	if e := entry("9880ba09-d5b8-4675-bfb4-2161d86d2d41", false); e.UntilEndOfTurn != game.TargetBenefit || e.Pump.Power != 3 || e.Pump.Toughness != 3 {
		t.Errorf("Brute Force declares %+v %+v, want a +3/+3 benefit", e, *e.Pump)
	}
	if e := entry("fd74f8eb-0253-42dd-8277-186d4934da38", false); e.UntilEndOfTurn != game.TargetHarm || e.Pump.Toughness != -5 {
		t.Errorf("Dismember declares %+v %+v, want a -5/-5 harm", e, *e.Pump)
	}
	if e := entry("60433b48-d27f-413c-905c-43839b1943f1", true); !slices.Equal(e.Pump.Keywords, []string{"protection"}) {
		t.Errorf("Mother of Runes declares %+v, want protection", *e.Pump)
	}
	if e := entry("f29dc596-2121-4421-8463-15f6c2e8b9b3", true); !e.Pump.CantBeBlocked {
		t.Errorf("Rogue's Passage declares %+v, want can't be blocked", *e.Pump)
	}
}
