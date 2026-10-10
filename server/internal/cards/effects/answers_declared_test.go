package effects

import (
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answers_declared_test.go — ADR 0142 §4, at zero (S4): every catalog
// row smart autopass asks about declares what it answers.
//
// The enumerator reads an untargeted instant-speed activation's
// Purpose.Answers, and a mana ability's ManaAbility.Answers. There is no
// printed-text read behind them any more: an undeclared activated row
// counts as interacting (owner answer 3). So every catalog row in scope
// declares, and this test fails on one that does not, naming it:
//
//	<oracle_id>\t<card name>\t<ref>\t<label>
//
// <ref> is own:<i> for a card's own row, grant:<bundle>:<i> for a row of
// an ADR 0093 bundle, token:<slug>:<i> for a token's, and mana:<i> (or
// token:<slug>:mana:<i>) for a mana ability with a sacrifice-another
// cost. A token's oracle column is "token".
//
// The scope is answersInScope's: a row that may be activated at instant
// speed with no target, and a mana ability that sacrifices another
// permanent. Fix a failure by declaring the row's Answers
// (docs/adding-cards.md, "Declaring what an ability answers").

// answersRow is one catalog row the enumerator reads the answers of.
type answersRow struct {
	oracle, name, ref, label string
	// Exactly one of activated and mana is set.
	activated *game.ActivatedAbilityShape
	mana      *game.ManaAbilityShape
}

func (r answersRow) line() string {
	clean := func(s string) string {
		return strings.Join(strings.Fields(strings.NewReplacer("\t", " ", "\n", " ").Replace(s)), " ")
	}
	return r.oracle + "\t" + clean(r.name) + "\t" + r.ref + "\t" + clean(r.label)
}

func (r answersRow) declared() game.Answers {
	if r.mana != nil {
		return r.mana.Answers
	}
	return r.activated.Purpose.Answers
}

// answersKeyOwner is who declares a catalog key's rows: the oracle
// column, the name, and the ref prefix.
type answersKeyOwner struct {
	oracle, name, prefix string
}

func answersKeyOwners() map[string]answersKeyOwner {
	out := map[string]answersKeyOwner{}
	for id, spec := range registry {
		out[id] = answersKeyOwner{oracle: id, name: spec.Name, prefix: "own"}
		if spec.Emblem != nil {
			out[game.EmblemKey(id)] = answersKeyOwner{oracle: id, name: spec.Name + " emblem", prefix: "emblem"}
		}
		for _, gr := range spec.Grants {
			out[game.GrantKey(gr.Key)] = answersKeyOwner{oracle: id, name: spec.Name, prefix: "grant:" + gr.Key}
		}
	}
	for slug, build := range tokenTemplatesBySlug {
		t := build()
		out[game.TokenKey(slug)] = answersKeyOwner{oracle: "token", name: t.Card.Name + " token", prefix: "token:" + slug}
		for _, gr := range t.Grants {
			out[game.GrantKey(gr.Key)] = answersKeyOwner{oracle: "token", name: t.Card.Name + " token", prefix: "grant:" + gr.Key}
		}
	}
	return out
}

// answersRows is every production catalog row in scope.
func answersRows(t *testing.T) []answersRow {
	t.Helper()
	if len(productionDefKeys) == 0 {
		t.Fatal("TestMain did not capture the production catalog keys")
	}
	owners := answersKeyOwners()
	var rows []answersRow
	for _, key := range productionDefKeys {
		d := defs[key]
		if d == nil {
			continue
		}
		o, ok := owners[key]
		if !ok {
			o = answersKeyOwner{oracle: key, name: key, prefix: "own"}
		}
		for i := range d.Activated {
			ab := d.Activated[i]
			if !answersInScope(ab) {
				continue
			}
			rows = append(rows, answersRow{
				oracle: o.oracle, name: o.name, ref: o.prefix + ":" + strconv.Itoa(i), label: ab.Label, activated: &ab,
			})
		}
		manaPrefix := "mana"
		if o.prefix != "own" {
			manaPrefix = o.prefix + ":mana"
		}
		for i := range d.ManaAbilities {
			m := d.ManaAbilities[i]
			if m.SacrificeOther == nil {
				continue
			}
			rows = append(rows, answersRow{
				oracle: o.oracle, name: o.name, ref: manaPrefix + ":" + strconv.Itoa(i), label: m.Label, mana: &m,
			})
		}
	}
	return rows
}

func TestEveryAnswersRowDeclares(t *testing.T) {
	rows := answersRows(t)
	if len(rows) == 0 {
		t.Fatal("no catalog rows in scope: the test is reading the wrong registry")
	}
	var missing []string
	for _, r := range rows {
		if r.declared() == 0 {
			missing = append(missing, r.line())
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("%d catalog rows smart autopass reads declare no Answers. An undeclared row stops its controller on every opponent's spell; declare Purpose.Answers (ManaAbility.Answers) on each (docs/adding-cards.md, \"Declaring what an ability answers\", ADR 0142):\n  %s",
			len(missing), strings.Join(missing, "\n  "))
	}
	t.Logf("%d rows in scope, all declared", len(rows))
}
