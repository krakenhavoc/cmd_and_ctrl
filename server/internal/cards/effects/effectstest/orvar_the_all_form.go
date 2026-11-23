package effectstest

// orvar_the_all_form_test.go — ADR 0153 PR 2 (#3168).

const OrvarOracle = "276c4643-13ad-43bf-840b-08c025ebc065"

func (d DboTable) CountByName(name string) int {
	n := 0
	for _, c := range d.G.BattlefieldCardsForEffect() {
		if c.Name == name {
			n++
		}
	}
	return n
}
